/**
 * useAudio — dual-element gapless playback with per-element WebAudio gain.
 *
 * A slot load is identified by both its queue occurrence ({queueIndex,
 * trackId}) and a monotonically increasing loadId. That prevents duplicate
 * track IDs, stale async descriptor responses, old media readiness events,
 * and delayed crossfade cleanup from being mistaken for the slot's current
 * load.
 *
 * A slot's PLAYBACK CYCLE is a finer identity — (loadId, playGen) — that
 * distinguishes each `ended` cycle of the same loaded media. Resuming the same
 * media after its cycle was terminated by an `ended` (after-track shutdown or
 * end-of-queue) bumps playGen, so a fresh legitimate `ended` is not mistaken
 * for the stale event of the already-terminated cycle. See PlaybackCycle in
 * gaplessEngine.
 */
"use client";

import { authedApi } from "@/lib/api/client";
import { type AudioGraphNode, createAudioGraph } from "@/lib/audio/audioGraph";
import { CROSSFADE_SECONDS, isCrossfadeEnabled } from "@/lib/audio/crossfade";
import {
  type PlaybackCycle,
  type QueueOccurrence,
  canEndedAdvance,
  canFinalizeReservedSlot,
  canSettleActivePlay,
  cyclesMatch,
  hasCurrentEndedEvidence,
  isSlotReusable,
  isStandbyReadyForOccurrence,
  occurrencesMatch,
  shouldOpenFreshEndedCycle,
  shouldRestartRepeatOne,
  shouldStartNaturalCrossfade,
  shouldTriggerPreload,
} from "@/lib/audio/gaplessEngine";
import { REPLAY_GAIN_CHANGE_EVENT, computeReplayGain, isReplayGainEnabled } from "@/lib/audio/replayGain";
import { resolveReplayGainDb } from "@/lib/audio/trackGainCache";
import { useAuthStore } from "@/store/auth";
import { type QueueTrack, useCurrentTrack, usePlayerStore } from "@/store/player";
import { isAfterTrackArmed, useSleepTimerStore } from "@/store/sleepTimer";
import { useEffect, useRef } from "react";

type ApiClient = ReturnType<typeof authedApi>;
type SlotIndex = 0 | 1;

interface Slot {
  audio: HTMLAudioElement;
  graph: AudioGraphNode;
  occurrence: QueueOccurrence | null;
  /** Incremented whenever this element is assigned a new intended load. */
  loadId: number;
  /** True only after this load generation emits canplay. */
  ready: boolean;
  resolvedAtMs: number;
  /** ReplayGain multiplier restored after a fade. */
  targetGain: number;
  /** Epoch ms when an in-progress fade-in ends (0 = not fading). */
  rampEndsAtMs: number;
  /** The exact load generation currently fading out and not yet reusable. */
  reservedLoadId: number | null;
  /** Invalidates pending play() settlements without replacing the loaded media. */
  playRequestId: number;
  /**
   * Playback-cycle generation for this element's currently loaded media. A new
   * media load leaves it as-is (identity carried by loadId); an explicit
   * replay of the SAME already-loaded media (resume after its `ended` cycle was
   * consumed/transitioned) bumps it, opening a fresh `ended` cycle so stale
   * cycle records stop matching. See PlaybackCycle in gaplessEngine.
   */
  playGen: number;
  readyListener: (() => void) | null;
}

interface PreparedSlot {
  slot: Slot;
  loadId: number;
}

async function resolvePlaybackUrl(api: ApiClient, trackId: string): Promise<string | null> {
  const { data, error } = await api.GET("/api/v1/catalog/tracks/{id}/playback", {
    params: { path: { id: trackId } },
  });
  if (error || !data) return null;
  if (data.presignedUrl) return data.presignedUrl;
  if (data.streamUrl) {
    return data.streamUrl.startsWith("/") ? `${window.location.origin}${data.streamUrl}` : data.streamUrl;
  }
  return null;
}

function setCurrentTimeSafely(audio: HTMLAudioElement, seconds: number) {
  try {
    audio.currentTime = seconds;
  } catch {
    // Metadata may not be available yet. Restore also retries on loadedmetadata.
  }
}

/**
 * Pin an element's playback rate to `speed`. Sets both `defaultPlaybackRate`
 * (which the media load algorithm restores `playbackRate` from on every
 * load()/src change) and `playbackRate` (the live rate) so a freshly-loaded
 * or swapped-in element never silently reverts to 1×.
 */
function applyPlaybackRate(audio: HTMLAudioElement, speed: number) {
  audio.defaultPlaybackRate = speed;
  audio.playbackRate = speed;
}

/**
 * Immediately silence every element and cancel crossfade state. Incrementing
 * playRequestId makes already-returned play() promises unable to settle state;
 * clearing reservations makes delayed fade cleanup idempotent.
 */
function neutralizePlayback(slots: [Slot, Slot]) {
  for (const slot of slots) {
    slot.playRequestId++;
    slot.audio.pause();
    slot.graph.setGain(slot.targetGain);
    slot.rampEndsAtMs = 0;
    slot.reservedLoadId = null;
  }
}

export function useAudio() {
  const slotsRef = useRef<[Slot, Slot] | null>(null);
  const activeSlotRef = useRef<SlotIndex>(0);
  const positionTickRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const restoreSeekedRef = useRef(false);
  const replayGainEnabledRef = useRef(isReplayGainEnabled());
  /** Globally unique across both slots and repeat-one playback cycles. */
  const nextLoadIdRef = useRef(1);
  /**
   * Last source playback cycle advanced off; blocks timeupdate + ended double
   * transitions. Records the whole cycle (loadId + playGen) so an explicit
   * replay of the same media (which bumps playGen) is not treated as the
   * already-transitioned cycle.
   */
  const lastTransitionSourceCycleRef = useRef<PlaybackCycle | null>(null);
  /**
   * Playback cycle whose `ended` event was already consumed by an
   * after-current-track sleep shutdown. A duplicate `ended` on the same
   * element and cycle must not advance/revive playback once the timer has
   * cleared. An explicit replay opens a fresh cycle (bumped playGen), so its
   * legitimate next `ended` is no longer matched here.
   */
  const endedConsumedCycleRef = useRef<PlaybackCycle | null>(null);

  const token = useAuthStore((s) => s.token);
  const currentTrack = useCurrentTrack();
  const { status, volume, speed, currentIndex, setStatus, setPosition, restoredPending, acknowledgeRestore } =
    usePlayerStore();

  /**
   * Latest playback speed, mirrored into a ref so slot-loading helpers
   * (init, prepareSlot, swaps, restore) can apply the current rate without
   * depending on a stale render closure. The `[speed]` effect below handles
   * live changes to already-loaded elements.
   */
  const speedRef = useRef(speed);
  speedRef.current = speed;

  // ── Initialise both audio elements once ─────────────────────────────────
  useEffect(() => {
    if (typeof window === "undefined") return;

    function makeSlot(): Slot {
      const audio = new Audio();
      audio.preload = "auto";
      // Set both: the media load algorithm resets playbackRate to
      // defaultPlaybackRate, so pinning both keeps the rate across load().
      audio.defaultPlaybackRate = speedRef.current;
      audio.playbackRate = speedRef.current;
      return {
        audio,
        graph: createAudioGraph(audio),
        occurrence: null,
        loadId: 0,
        ready: false,
        resolvedAtMs: 0,
        targetGain: 1,
        rampEndsAtMs: 0,
        reservedLoadId: null,
        playRequestId: 0,
        playGen: 0,
        readyListener: null,
      };
    }

    const slots: [Slot, Slot] = [makeSlot(), makeSlot()];
    slotsRef.current = slots;

    // Zustand notifies subscribers synchronously inside pause(). This closes
    // the window where a fixed sleep-timer expiry has updated player intent,
    // but a pending audio.play() microtask could settle before React's status
    // effect gets a chance to pause and invalidate both elements.
    const unsubscribePlayer = usePlayerStore.subscribe((state, previousState) => {
      if (state.status === "paused" && previousState.status !== "paused") {
        neutralizePlayback(slots);
      }
    });

    return () => {
      unsubscribePlayer();
      for (const slot of slots) {
        if (slot.readyListener) slot.audio.removeEventListener("canplay", slot.readyListener);
        slot.audio.pause();
        slot.audio.src = "";
        slot.graph.disconnect();
      }
      slotsRef.current = null;
      if (positionTickRef.current) clearInterval(positionTickRef.current);
    };
  }, []);

  function occurrenceFor(index: number, track: QueueTrack | undefined): QueueOccurrence | null {
    return track && index >= 0 ? { queueIndex: index, trackId: track.id } : null;
  }

  function resetSlotForLoad(slot: Slot, occurrence: QueueOccurrence): number {
    if (slot.readyListener) {
      slot.audio.removeEventListener("canplay", slot.readyListener);
      slot.readyListener = null;
    }
    slot.audio.pause();
    slot.loadId = nextLoadIdRef.current++;
    slot.reservedLoadId = null;
    slot.playRequestId++;
    slot.occurrence = occurrence;
    slot.ready = false;
    slot.resolvedAtMs = 0;
    slot.targetGain = 1;
    slot.rampEndsAtMs = 0;
    slot.graph.setGain(1);
    setCurrentTimeSafely(slot.audio, 0);
    slot.audio.removeAttribute("src");
    slot.audio.load();
    return slot.loadId;
  }

  function clearFailedLoad(slot: Slot, loadId: number) {
    if (slot.loadId !== loadId) return;
    if (slot.readyListener) {
      slot.audio.removeEventListener("canplay", slot.readyListener);
      slot.readyListener = null;
    }
    slot.occurrence = null;
    slot.ready = false;
    slot.resolvedAtMs = 0;
  }

  async function applyGainToSlot(slotIdx: SlotIndex, occurrence: QueueOccurrence, loadId: number, authToken: string) {
    const slots = slotsRef.current;
    if (!slots) return;
    const slot = slots[slotIdx];
    let gain = 1;
    if (replayGainEnabledRef.current) {
      const db = await resolveReplayGainDb(authedApi(authToken), occurrence.trackId);
      gain = computeReplayGain(db);
    }
    if (slot.loadId !== loadId || !occurrencesMatch(slot.occurrence, occurrence)) return;
    slot.targetGain = gain;
    if (slot.rampEndsAtMs > Date.now()) {
      slot.graph.rampGain(gain, Math.max(0, (slot.rampEndsAtMs - Date.now()) / 1000));
    } else {
      slot.rampEndsAtMs = 0;
      slot.graph.setGain(gain);
    }
  }

  async function prepareSlot(
    slotIdx: SlotIndex,
    occurrence: QueueOccurrence,
    authToken: string
  ): Promise<PreparedSlot | null> {
    const slots = slotsRef.current;
    if (!slots) return null;
    const slot = slots[slotIdx];
    if (!isSlotReusable(slot.reservedLoadId, slot.loadId)) return null;
    const loadId = resetSlotForLoad(slot, occurrence);
    const url = await resolvePlaybackUrl(authedApi(authToken), occurrence.trackId).catch(() => null);

    if (slot.loadId !== loadId || !occurrencesMatch(slot.occurrence, occurrence)) return null;
    if (!url) {
      clearFailedLoad(slot, loadId);
      return null;
    }

    const markReady = () => {
      if (slot.loadId !== loadId || !occurrencesMatch(slot.occurrence, occurrence)) return;
      slot.ready = true;
      slot.readyListener = null;
    };
    slot.readyListener = markReady;
    slot.audio.addEventListener("canplay", markReady, { once: true });
    slot.audio.src = url;
    slot.resolvedAtMs = Date.now();
    setCurrentTimeSafely(slot.audio, 0);
    slot.audio.load();
    applyPlaybackRate(slot.audio, speedRef.current);
    if (slot.audio.readyState >= HTMLMediaElement.HAVE_FUTURE_DATA) markReady();

    void applyGainToSlot(slotIdx, occurrence, loadId, authToken);
    return { slot, loadId };
  }

  function updateMediaSession(track: QueueTrack) {
    if ("mediaSession" in navigator) {
      navigator.mediaSession.metadata = new MediaMetadata({
        title: track.title,
        artist: track.artistName,
        album: track.albumTitle,
        artwork: track.artworkUrl ? [{ src: track.artworkUrl, sizes: "512x512", type: "image/jpeg" }] : [],
      });
    }
  }

  function recordHistory(trackId: string, authToken: string | null) {
    if (!authToken) return;
    authedApi(authToken)
      .POST("/api/v1/me/history", { body: { trackId, playedAt: new Date().toISOString() } })
      .catch(() => {});
  }

  function startSlotPlay(slotIdx: SlotIndex, expectedLoadId: number) {
    const slots = slotsRef.current;
    if (!slots) return null;
    const slot = slots[slotIdx];
    const playRequestId = slot.playRequestId + 1;
    if (
      !canSettleActivePlay(
        usePlayerStore.getState().status,
        slotIdx,
        activeSlotRef.current,
        expectedLoadId,
        slot.loadId,
        playRequestId,
        playRequestId
      )
    ) {
      return null;
    }
    slot.playRequestId = playRequestId;
    return { playRequestId, promise: slot.audio.play() };
  }

  function canSettleSlotPlay(slotIdx: SlotIndex, loadId: number, playRequestId: number): boolean {
    const slots = slotsRef.current;
    if (!slots) return false;
    return canSettleActivePlay(
      usePlayerStore.getState().status,
      slotIdx,
      activeSlotRef.current,
      loadId,
      slots[slotIdx].loadId,
      playRequestId,
      slots[slotIdx].playRequestId
    );
  }

  function activateStandby(
    standbyIdx: SlotIndex,
    intended: QueueOccurrence,
    track: QueueTrack,
    crossfade: boolean,
    advanceStore: boolean
  ): boolean {
    const slots = slotsRef.current;
    if (!slots) return false;
    const prevIdx = activeSlotRef.current;
    if (prevIdx === standbyIdx) return false;
    const prevSlot = slots[prevIdx];
    const nextSlot = slots[standbyIdx];
    if (!isSlotReusable(nextSlot.reservedLoadId, nextSlot.loadId)) return false;
    if (!isStandbyReadyForOccurrence(nextSlot, intended, Date.now())) return false;
    if (cyclesMatch(lastTransitionSourceCycleRef.current, { loadId: prevSlot.loadId, playGen: prevSlot.playGen }))
      return false;

    const sourceLoadId = prevSlot.loadId;
    lastTransitionSourceCycleRef.current = { loadId: sourceLoadId, playGen: prevSlot.playGen };
    activeSlotRef.current = standbyIdx;
    applyPlaybackRate(nextSlot.audio, speedRef.current);
    setCurrentTimeSafely(nextSlot.audio, 0);
    if (advanceStore) usePlayerStore.getState().skipToNext();

    const useOverlap = crossfade && nextSlot.graph.active && prevSlot.graph.active;
    if (useOverlap) {
      const cleanupSlotIdx = prevIdx;
      const cleanupLoadId = sourceLoadId;
      const destinationLoadId = nextSlot.loadId;
      const finalizeSource = () => {
        const currentSlots = slotsRef.current;
        if (!currentSlots) return false;
        const sourceSlot = currentSlots[cleanupSlotIdx];
        if (
          !canFinalizeReservedSlot(
            cleanupSlotIdx,
            activeSlotRef.current,
            cleanupLoadId,
            sourceSlot.loadId,
            sourceSlot.reservedLoadId
          )
        ) {
          return false;
        }
        sourceSlot.audio.pause();
        setCurrentTimeSafely(sourceSlot.audio, 0);
        sourceSlot.graph.setGain(sourceSlot.targetGain);
        sourceSlot.rampEndsAtMs = 0;
        sourceSlot.reservedLoadId = null;
        return true;
      };

      prevSlot.reservedLoadId = sourceLoadId;
      nextSlot.graph.setGain(0);
      nextSlot.rampEndsAtMs = Date.now() + CROSSFADE_SECONDS * 1000;
      const play = startSlotPlay(standbyIdx, destinationLoadId);
      if (!play) {
        prevSlot.reservedLoadId = null;
        nextSlot.graph.setGain(nextSlot.targetGain);
        nextSlot.rampEndsAtMs = 0;
        return false;
      }
      play.promise
        .then(() => {
          if (!canSettleSlotPlay(standbyIdx, destinationLoadId, play.playRequestId)) return;
          const currentSlots = slotsRef.current;
          if (!currentSlots) return;
          currentSlots[standbyIdx].graph.rampGain(currentSlots[standbyIdx].targetGain, CROSSFADE_SECONDS);
          currentSlots[cleanupSlotIdx].graph.rampGain(0, CROSSFADE_SECONDS);
          setStatus("playing");
        })
        .catch(() => {
          finalizeSource();
          if (!canSettleSlotPlay(standbyIdx, destinationLoadId, play.playRequestId)) return;
          const currentSlots = slotsRef.current;
          if (!currentSlots) return;
          currentSlots[standbyIdx].graph.setGain(currentSlots[standbyIdx].targetGain);
          currentSlots[standbyIdx].rampEndsAtMs = 0;
          setStatus("paused");
        });

      setTimeout(finalizeSource, CROSSFADE_SECONDS * 1000);
    } else {
      prevSlot.audio.pause();
      setCurrentTimeSafely(prevSlot.audio, 0);
      prevSlot.graph.setGain(prevSlot.targetGain);
      nextSlot.graph.setGain(nextSlot.targetGain);
      const destinationLoadId = nextSlot.loadId;
      const play = startSlotPlay(standbyIdx, destinationLoadId);
      if (!play) return false;
      play.promise
        .then(() => {
          if (canSettleSlotPlay(standbyIdx, destinationLoadId, play.playRequestId)) setStatus("playing");
        })
        .catch(() => {
          if (canSettleSlotPlay(standbyIdx, destinationLoadId, play.playRequestId)) setStatus("paused");
        });
    }

    updateMediaSession(track);
    return true;
  }

  function restartRepeatOne(slotIdx: SlotIndex, occurrence: QueueOccurrence, authToken: string | null) {
    const slots = slotsRef.current;
    if (!slots) return;
    const slot = slots[slotIdx];
    lastTransitionSourceCycleRef.current = { loadId: slot.loadId, playGen: slot.playGen };
    recordHistory(occurrence.trackId, authToken);
    const loadId = nextLoadIdRef.current++;
    slot.loadId = loadId;
    slot.playGen = 0;
    slot.playRequestId++;
    setCurrentTimeSafely(slot.audio, 0);
    slot.graph.setGain(slot.targetGain);
    usePlayerStore.getState().setPosition(0);
    const play = startSlotPlay(slotIdx, loadId);
    if (!play) return;
    play.promise
      .then(() => {
        if (canSettleSlotPlay(slotIdx, loadId, play.playRequestId)) setStatus("playing");
      })
      .catch(() => {
        if (canSettleSlotPlay(slotIdx, loadId, play.playRequestId)) setStatus("paused");
      });
  }

  function advanceFromActive(reason: "natural-crossfade" | "ended", authToken: string | null): boolean {
    const slots = slotsRef.current;
    if (!slots) return false;
    const sourceIdx = activeSlotRef.current;
    const sourceSlot = slots[sourceIdx];
    const sourceOccurrence = sourceSlot.occurrence;
    const sourceCycle: PlaybackCycle = { loadId: sourceSlot.loadId, playGen: sourceSlot.playGen };
    if (!sourceOccurrence || cyclesMatch(lastTransitionSourceCycleRef.current, sourceCycle)) return false;

    const state = usePlayerStore.getState();
    if (state.repeat === "one") {
      if (shouldRestartRepeatOne(state.repeat, reason)) {
        restartRepeatOne(sourceIdx, sourceOccurrence, authToken);
        return true;
      }
      return false;
    }

    const intended = computeNextOccurrence(state.queue, state.currentIndex, state.repeat, state.shuffle);
    if (!intended) {
      if (reason === "ended") {
        lastTransitionSourceCycleRef.current = sourceCycle;
        recordHistory(sourceOccurrence.trackId, authToken);
        state.skipToNext();
      }
      return false;
    }

    const standbyIdx: SlotIndex = sourceIdx === 0 ? 1 : 0;
    const nextTrack = state.queue[intended.queueIndex];
    if (!nextTrack) return false;
    const activated = activateStandby(standbyIdx, intended, nextTrack, reason === "natural-crossfade", true);
    if (activated) recordHistory(sourceOccurrence.trackId, authToken);
    if (!activated && reason === "ended") {
      lastTransitionSourceCycleRef.current = sourceCycle;
      recordHistory(sourceOccurrence.trackId, authToken);
      state.skipToNext();
    }
    return activated;
  }

  // ── React to ReplayGain toggle changes live ─────────────────────────────
  useEffect(() => {
    function onChange(e: Event) {
      replayGainEnabledRef.current = (e as CustomEvent<boolean>).detail;
      const slots = slotsRef.current;
      if (!slots || !token) return;
      const slotIdx = activeSlotRef.current;
      const slot = slots[slotIdx];
      if (slot.occurrence) void applyGainToSlot(slotIdx, slot.occurrence, slot.loadId, token);
    }
    window.addEventListener(REPLAY_GAIN_CHANGE_EVENT, onChange);
    return () => window.removeEventListener(REPLAY_GAIN_CHANGE_EVENT, onChange);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

  // ── Restore-from-persistence: seek without autoplay ─────────────────────
  useEffect(() => {
    if (!restoredPending || restoreSeekedRef.current) return;
    const intended = occurrenceFor(currentIndex, currentTrack ?? undefined);
    if (!intended || !currentTrack || !token) return;
    restoreSeekedRef.current = true;
    setStatus("loading");

    const slotIdx = activeSlotRef.current;
    void prepareSlot(slotIdx, intended, token).then((prepared) => {
      if (!prepared) {
        setStatus("error");
        return;
      }
      const restoredPosition = usePlayerStore.getState().positionSeconds;
      const seekRestoredPosition = () => setCurrentTimeSafely(prepared.slot.audio, restoredPosition);
      if (prepared.slot.audio.readyState >= HTMLMediaElement.HAVE_METADATA) seekRestoredPosition();
      else prepared.slot.audio.addEventListener("loadedmetadata", seekRestoredPosition, { once: true });
      setStatus("paused");
      updateMediaSession(currentTrack);
      acknowledgeRestore();
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [restoredPending, currentIndex, currentTrack?.id, token]);

  // ── Load/swap when the selected queue occurrence changes ────────────────
  useEffect(() => {
    if (restoredPending || !token) return;
    const intended = occurrenceFor(currentIndex, currentTrack ?? undefined);
    const slots = slotsRef.current;
    if (!intended || !currentTrack || !slots) return;

    const active = slots[activeSlotRef.current];
    if (occurrencesMatch(active.occurrence, intended)) return;

    const standbyIdx: SlotIndex = activeSlotRef.current === 0 ? 1 : 0;
    const standby = slots[standbyIdx];
    if (
      isStandbyReadyForOccurrence(standby, intended, Date.now()) &&
      activateStandby(standbyIdx, intended, currentTrack, isCrossfadeEnabled(), false)
    ) {
      return;
    }

    // The matching standby may still be reserved by a fade-out. In that
    // case, load the explicit selection into the active slot rather than
    // stalling until a timer releases the other element.
    const activeIdx = activeSlotRef.current;
    setStatus("loading");
    void prepareSlot(activeIdx, intended, token).then((prepared) => {
      if (!prepared) {
        setStatus("error");
        return;
      }
      const play = startSlotPlay(activeIdx, prepared.loadId);
      if (!play) return;
      play.promise
        .then(() => {
          if (canSettleSlotPlay(activeIdx, prepared.loadId, play.playRequestId)) setStatus("playing");
        })
        .catch(() => {
          if (canSettleSlotPlay(activeIdx, prepared.loadId, play.playRequestId)) setStatus("paused");
        });
      updateMediaSession(currentTrack);
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentIndex, currentTrack?.id, token, restoredPending]);

  // ── Preload the exact next queue occurrence into standby ────────────────
  useEffect(() => {
    if (!token || status !== "playing") return;
    const authToken = token;

    function maybePreload() {
      const slots = slotsRef.current;
      if (!slots) return;
      const active = slots[activeSlotRef.current];
      if (!shouldTriggerPreload(active.audio.currentTime, active.audio.duration)) return;

      const state = usePlayerStore.getState();
      const intended = computeNextOccurrence(state.queue, state.currentIndex, state.repeat, state.shuffle);
      if (!intended) return;
      const standbyIdx: SlotIndex = activeSlotRef.current === 0 ? 1 : 0;
      const standby = slots[standbyIdx];
      if (occurrencesMatch(standby.occurrence, intended)) return;
      void prepareSlot(standbyIdx, intended, authToken);
    }

    const id = setInterval(maybePreload, 1000);
    return () => clearInterval(id);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [status, token]);

  // ── Play / Pause ───────────────────────────────────────────────────────
  useEffect(() => {
    const slots = slotsRef.current;
    if (!slots) return;
    const activeIdx = activeSlotRef.current;
    const active = slots[activeIdx];
    if (status === "playing") {
      // Explicit resume/replay boundary: if this slot's current playback cycle
      // was already terminated by a prior `ended` (consumed by after-track
      // shutdown, or transitioned off at end-of-queue), replaying the SAME
      // media must open a fresh ended cycle. Bump playGen here — and only here
      // — so the legitimate next `ended` of the replayed media is no longer
      // matched by the stale consumed/transition records. A normal mid-track
      // pause→play and a fixed-timer resume leave playGen untouched, preserving
      // stale-event inertness.
      if (
        shouldOpenFreshEndedCycle(
          { loadId: active.loadId, playGen: active.playGen },
          endedConsumedCycleRef.current,
          lastTransitionSourceCycleRef.current
        )
      ) {
        active.playGen++;
      }
      const play = startSlotPlay(activeIdx, active.loadId);
      play?.promise.catch(() => {
        if (canSettleSlotPlay(activeIdx, active.loadId, play.playRequestId)) setStatus("paused");
      });
    } else if (status === "paused") {
      neutralizePlayback(slots);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [status]);

  // ── User volume stays orthogonal to ReplayGain ─────────────────────────
  useEffect(() => {
    const slots = slotsRef.current;
    if (!slots) return;
    for (const slot of slots) slot.audio.volume = volume;
  }, [volume]);

  // ── Playback speed applies to BOTH elements (current + preloaded standby) ─
  useEffect(() => {
    const slots = slotsRef.current;
    if (!slots) return;
    for (const slot of slots) applyPlaybackRate(slot.audio, speed);
  }, [speed]);

  // ── Position + natural crossfade ticker ─────────────────────────────────
  useEffect(() => {
    if (positionTickRef.current) clearInterval(positionTickRef.current);
    if (status === "playing") {
      positionTickRef.current = setInterval(() => {
        const slots = slotsRef.current;
        if (!slots) return;
        const active = slots[activeSlotRef.current];
        setPosition(active.audio.currentTime);

        // When after-current-track sleep mode is armed, the current track must
        // be the last thing that plays — a lead-time crossfade would advance
        // to the next track before `ended` fires, bypassing that semantic. Let
        // the track run out and let the ended handler pause.
        if (isAfterTrackArmed(useSleepTimerStore.getState())) return;
        if (!isCrossfadeEnabled() || !active.graph.active) return;
        const state = usePlayerStore.getState();
        const intended = computeNextOccurrence(state.queue, state.currentIndex, state.repeat, state.shuffle);
        if (!intended) return;
        const standbyIdx: SlotIndex = activeSlotRef.current === 0 ? 1 : 0;
        const standby = slots[standbyIdx];
        const standbyReady = standby.graph.active && isStandbyReadyForOccurrence(standby, intended, Date.now());
        if (
          shouldStartNaturalCrossfade(active.audio.currentTime, active.audio.duration, CROSSFADE_SECONDS, standbyReady)
        ) {
          advanceFromActive("natural-crossfade", token);
        }
      }, 250);
    }
    return () => {
      if (positionTickRef.current) clearInterval(positionTickRef.current);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [status, token]);

  // ── Ended fallback + repeat-one restart ─────────────────────────────────
  useEffect(() => {
    const slots = slotsRef.current;
    if (!slots) return;

    function makeOnEnded(slotIdx: SlotIndex) {
      return (event: Event) => {
        if (slotIdx !== activeSlotRef.current) return;
        const slots = slotsRef.current;
        if (!slots) return;
        const sourceSlot = slots[slotIdx];
        if (!hasCurrentEndedEvidence(event.isTrusted, sourceSlot.audio.ended)) return;
        const sourceCycle: PlaybackCycle = { loadId: sourceSlot.loadId, playGen: sourceSlot.playGen };

        // An ended event may advance/repeat only while playback intent still
        // holds for this exact source cycle. A fixed sleep-timer expiry pauses
        // synchronously, so a late ended on the still-active source is rejected
        // here (status is "paused"); a duplicate after-track ended is rejected
        // because the cycle was marked consumed below.
        if (
          !canEndedAdvance(
            usePlayerStore.getState().status,
            sourceCycle,
            endedConsumedCycleRef.current,
            lastTransitionSourceCycleRef.current
          )
        ) {
          return;
        }

        // Sleep timer "after current track": handleTrackEnded pauses the
        // player and clears the timer. Mark this source cycle consumed FIRST
        // so a duplicate ended cannot slip through after the timer clears, then
        // do NOT advance/repeat/crossfade past this track.
        if (useSleepTimerStore.getState().handleTrackEnded()) {
          endedConsumedCycleRef.current = sourceCycle;
          return;
        }
        advanceFromActive("ended", token);
      };
    }

    const handlerA = makeOnEnded(0);
    const handlerB = makeOnEnded(1);
    slots[0].audio.addEventListener("ended", handlerA);
    slots[1].audio.addEventListener("ended", handlerB);
    return () => {
      slots[0].audio.removeEventListener("ended", handlerA);
      slots[1].audio.removeEventListener("ended", handlerB);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentIndex, currentTrack?.id, token]);

  // ── External seek events (keyboard shortcuts / player UI) ──────────────
  useEffect(() => {
    function onSeek(e: Event) {
      const seconds = (e as CustomEvent<number>).detail;
      const slots = slotsRef.current;
      if (typeof seconds === "number" && slots) {
        setCurrentTimeSafely(slots[activeSlotRef.current].audio, Math.max(0, seconds));
        setPosition(Math.max(0, seconds));
      }
    }
    window.addEventListener("inori:seek", onSeek);
    return () => window.removeEventListener("inori:seek", onSeek);
  }, [setPosition]);

  // ── MediaSession action handlers ───────────────────────────────────────
  useEffect(() => {
    if (typeof window === "undefined" || !("mediaSession" in navigator)) return;
    navigator.mediaSession.setActionHandler("play", () => usePlayerStore.getState().play());
    navigator.mediaSession.setActionHandler("pause", () => usePlayerStore.getState().pause());
    navigator.mediaSession.setActionHandler("nexttrack", () => usePlayerStore.getState().skipToNext());
    navigator.mediaSession.setActionHandler("previoustrack", () => usePlayerStore.getState().skipToPrevious());
    return () => {
      navigator.mediaSession.setActionHandler("play", null);
      navigator.mediaSession.setActionHandler("pause", null);
      navigator.mediaSession.setActionHandler("nexttrack", null);
      navigator.mediaSession.setActionHandler("previoustrack", null);
    };
  }, []);

  function seek(seconds: number) {
    const slots = slotsRef.current;
    if (!slots) return;
    setCurrentTimeSafely(slots[activeSlotRef.current].audio, seconds);
    setPosition(seconds);
  }

  return { seek };
}

function computeNextOccurrence(
  queue: QueueTrack[],
  currentIndex: number,
  repeat: "off" | "one" | "all",
  shuffle: boolean
): QueueOccurrence | null {
  if (queue.length === 0 || repeat === "one" || shuffle) return null;
  let nextIndex = currentIndex + 1;
  if (nextIndex >= queue.length) {
    if (repeat !== "all") return null;
    nextIndex = 0;
  }
  const nextTrack = queue[nextIndex];
  return nextTrack ? { queueIndex: nextIndex, trackId: nextTrack.id } : null;
}
