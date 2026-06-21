/**
 * useAudio — wires the HTMLAudioElement to the PlayerStore.
 *
 * Playback URL resolution (in priority order):
 *   1. presignedUrl  — S3-compatible presigned GET URL (16-min TTL)
 *   2. streamUrl     — server-proxied /stream endpoint for local/NFS/SMB backends
 *                      Uses ?token= query param because <audio> cannot set headers.
 *   3. error         — no URL available
 */
"use client";

import { useEffect, useRef } from "react";
import { usePlayerStore, useCurrentTrack } from "@/store/player";
import { useAuthStore } from "@/store/auth";
import { authedApi } from "@/lib/api/client";

export function useAudio() {
  const audioRef = useRef<HTMLAudioElement | null>(null);
  const trackRef = useRef<string | null>(null);
  const positionTickRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const token = useAuthStore((s) => s.token);
  const currentTrack = useCurrentTrack();
  const { status, volume, setStatus, setPosition, skipToNext } = usePlayerStore();

  // ── Initialise audio element once ──────────────────────────────────────
  useEffect(() => {
    if (typeof window === "undefined") return;
    const audio = new Audio();
    audio.preload = "auto";
    audioRef.current = audio;

    return () => {
      audio.pause();
      audio.src = "";
      if (positionTickRef.current) clearInterval(positionTickRef.current);
    };
  }, []);

  // ── Load track when currentTrack changes ───────────────────────────────
  useEffect(() => {
    const audio = audioRef.current;
    if (!audio || !currentTrack || !token) return;
    if (trackRef.current === currentTrack.id) return;

    trackRef.current = currentTrack.id;
    setStatus("loading");

    authedApi(token)
      .GET("/api/v1/catalog/tracks/{id}/playback", {
        params: { path: { id: currentTrack.id } },
      })
      .then(({ data, error }) => {
        if (error || !data) {
          setStatus("error");
          return;
        }

        // Prefer presigned URL; fall back to server-proxied stream endpoint.
        let playUrl: string | null = null;
        if (data.presignedUrl) {
          playUrl = data.presignedUrl;
        } else if (data.streamUrl) {
          // Append the viewer JWT as a query param so the server can
          // authenticate without a custom Authorization header.
          const base = data.streamUrl.startsWith("/")
            ? `${window.location.origin}${data.streamUrl}`
            : data.streamUrl;
          const u = new URL(base);
          u.searchParams.set("token", token);
          playUrl = u.toString();
        }

        if (!playUrl) {
          setStatus("error");
          return;
        }

        audio.src = playUrl;
        audio.load();
        audio
          .play()
          .then(() => setStatus("playing"))
          .catch(() => setStatus("paused"));

        if ("mediaSession" in navigator) {
          navigator.mediaSession.metadata = new MediaMetadata({
            title: currentTrack.title,
            artist: currentTrack.artistName,
            album: currentTrack.albumTitle,
            artwork: currentTrack.artworkUrl
              ? [{ src: currentTrack.artworkUrl, sizes: "512x512", type: "image/jpeg" }]
              : [],
          });
        }
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentTrack?.id, token]);

  // ── Play / Pause ───────────────────────────────────────────────────────
  useEffect(() => {
    const audio = audioRef.current;
    if (!audio) return;
    if (status === "playing") {
      audio.play().catch(() => setStatus("paused"));
    } else if (status === "paused") {
      audio.pause();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [status]);

  // ── Volume ─────────────────────────────────────────────────────────────
  useEffect(() => {
    if (audioRef.current) audioRef.current.volume = volume;
  }, [volume]);

  // ── Position ticker ────────────────────────────────────────────────────
  useEffect(() => {
    if (positionTickRef.current) clearInterval(positionTickRef.current);
    if (status === "playing") {
      positionTickRef.current = setInterval(() => {
        const audio = audioRef.current;
        if (audio) setPosition(audio.currentTime);
      }, 250);
    }
    return () => { if (positionTickRef.current) clearInterval(positionTickRef.current); };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [status]);

  // ── Track ended ────────────────────────────────────────────────────────
  useEffect(() => {
    const audio = audioRef.current;
    if (!audio) return;

    function onEnded() {
      if (token && currentTrack) {
        authedApi(token)
          .POST("/api/v1/me/history", {
            body: {
              trackId: currentTrack.id,
              playedAt: new Date().toISOString(),
              durationSeconds: Math.round(currentTrack.durationSeconds),
              source: "web",
            },
          })
          .catch(() => {});
      }
      skipToNext();
    }

    audio.addEventListener("ended", onEnded);
    return () => audio.removeEventListener("ended", onEnded);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentTrack?.id, token]);

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
    if (audioRef.current) {
      audioRef.current.currentTime = seconds;
      setPosition(seconds);
    }
  }

  return { seek };
}
