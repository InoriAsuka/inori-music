/**
 * audioGraph.ts — WebAudio gain pipeline wrapping a single HTMLAudioElement.
 *
 * Each dual-element gapless engine slot gets one AudioGraphNode: the audio
 * element feeds a MediaElementAudioSourceNode -> GainNode -> destination.
 * ReplayGain and crossfade ramps both apply through the same GainNode,
 * independent of `audio.volume` (which carries the user's volume slider).
 *
 * CORS: MediaElementAudioSourceNode "taints" cross-origin audio unless the
 * element has crossOrigin="anonymous" AND the response carries a matching
 * Access-Control-Allow-Origin header. A tainted element still plays audio
 * normally through <audio> — WebAudio just can't read/process the samples,
 * which throws on `createMediaElementSource` in some browsers or silently
 * produces no sound in others. We detect this and fall back to direct
 * (ungained) playback so users never lose audio because of a CORS gap; see
 * docs/operations/audio-cors.md for the server-side requirements.
 */
"use client";

let sharedContext: AudioContext | null = null;
let resumeListenersInstalled = false;

/** Lazily creates (or returns) the single shared AudioContext for the page. */
function getSharedContext(): AudioContext | null {
  if (typeof window === "undefined") return null;
  const Ctor =
    window.AudioContext ?? (window as unknown as { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;
  if (!Ctor) return null;
  if (!sharedContext) {
    sharedContext = new Ctor();
  }
  installResumeOnGesture(sharedContext);
  return sharedContext;
}

/**
 * Browsers suspend new AudioContexts until a user gesture. Resume on the
 * first pointerdown/keydown/touchstart so playback isn't silently gained-to-zero
 * by a context stuck in "suspended".
 */
function installResumeOnGesture(ctx: AudioContext) {
  if (resumeListenersInstalled || typeof window === "undefined") return;
  resumeListenersInstalled = true;
  const resume = () => {
    if (ctx.state === "suspended") ctx.resume().catch(() => {});
  };
  const opts = { passive: true } as const;
  window.addEventListener("pointerdown", resume, opts);
  window.addEventListener("keydown", resume, opts);
  window.addEventListener("touchstart", resume, opts);
}

export interface AudioGraphNode {
  /** True when WebAudio graph is active for this element; false = direct-playback fallback. */
  readonly active: boolean;
  readonly gainNode: GainNode | null;
  readonly eqFilters: BiquadFilterNode[];
  /** Set the gain immediately (no ramp). */
  setGain(value: number): void;
  /** Linearly ramp gain from its current value to `target` over `seconds`. */
  rampGain(target: number, seconds: number): void;
  /** Update one EQ band's peaking filter gain in dB. */
  setEqBand(index: number, db: number): void;
  disconnect(): void;
}

/**
 * Wraps `audio` with an AudioContext -> MediaElementAudioSourceNode ->
 * BiquadFilter[] -> GainNode -> destination graph.
 * Sets `crossOrigin = "anonymous"` on the element (must be set before
 * the graph is created / before `src` triggers a fetch that doesn't request
 * CORS headers). Falls back to a no-op passthrough node (audio plays directly,
 * ReplayGain/crossfade/EQ become inert) if AudioContext is unavailable or graph
 * creation throws (CORS-tainted or otherwise unsupported).
 */
export function createAudioGraph(audio: HTMLAudioElement): AudioGraphNode {
  audio.crossOrigin = "anonymous";

  const ctx = getSharedContext();
  if (!ctx) {
    return createFallbackNode();
  }

  try {
    const source = ctx.createMediaElementSource(audio);
    const eqFilters = createEqFilterChain(ctx);
    const gainNode = ctx.createGain();
    gainNode.gain.value = 1.0;

    let chain: AudioNode = source;
    for (const filter of eqFilters) {
      chain.connect(filter);
      chain = filter;
    }
    chain.connect(gainNode);
    gainNode.connect(ctx.destination);

    return {
      active: true,
      gainNode,
      eqFilters,
      setGain(value) {
        gainNode.gain.cancelScheduledValues(ctx.currentTime);
        gainNode.gain.setValueAtTime(value, ctx.currentTime);
      },
      rampGain(target, seconds) {
        const now = ctx.currentTime;
        gainNode.gain.cancelScheduledValues(now);
        gainNode.gain.setValueAtTime(gainNode.gain.value, now);
        gainNode.gain.linearRampToValueAtTime(target, now + seconds);
      },
      setEqBand(index, db) {
        if (index < 0 || index >= eqFilters.length) return;
        const filter = eqFilters[index];
        filter.gain.cancelScheduledValues(ctx.currentTime);
        filter.gain.setValueAtTime(db, ctx.currentTime);
      },
      disconnect() {
        try {
          source.disconnect();
          for (const filter of eqFilters) {
            filter.disconnect();
          }
          gainNode.disconnect();
        } catch {
          // Already disconnected — non-fatal.
        }
      },
    };
  } catch {
    // createMediaElementSource throws if this element is already connected to
    // another graph, or in rare CORS/codec edge cases. Degrade gracefully.
    return createFallbackNode();
  }
}

const EQ_FREQUENCIES = [31, 62, 125, 250, 500, 1000, 2000, 4000, 8000, 16000];

function createEqFilterChain(ctx: BaseAudioContext): BiquadFilterNode[] {
  return EQ_FREQUENCIES.map((freq) => {
    const filter = ctx.createBiquadFilter();
    filter.type = "peaking";
    filter.frequency.value = freq;
    filter.Q.value = 1.4;
    filter.gain.value = 0;
    return filter;
  });
}

/** No-op graph used when WebAudio is unavailable — audio plays directly via the element. */
function createFallbackNode(): AudioGraphNode {
  return {
    active: false,
    gainNode: null,
    eqFilters: [],
    setGain() {},
    rampGain() {},
    setEqBand() {},
    disconnect() {},
  };
}
