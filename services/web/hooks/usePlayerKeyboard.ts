"use client";

import { useEffect } from "react";
import { usePlayerStore } from "@/store/player";

function isTypingTarget(el: EventTarget | null): boolean {
  if (!(el instanceof HTMLElement)) return false;
  const tag = el.tagName.toLowerCase();
  return tag === "input" || tag === "textarea" || tag === "select" || el.isContentEditable;
}

/** Global player keyboard shortcuts. */
export function usePlayerKeyboard() {
  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if (isTypingTarget(e.target)) return;
      const s = usePlayerStore.getState();

      switch (e.key.toLowerCase()) {
        case " ":
          e.preventDefault();
          if (s.status === "playing") s.pause();
          else s.play();
          break;
        case "arrowleft":
          e.preventDefault();
          window.dispatchEvent(new CustomEvent("inori:seek", { detail: Math.max(0, s.positionSeconds - 5) }));
          break;
        case "arrowright":
          e.preventDefault();
          window.dispatchEvent(new CustomEvent("inori:seek", { detail: s.positionSeconds + 5 }));
          break;
        case "arrowup":
          e.preventDefault();
          s.setVolume(Math.min(1, s.volume + 0.1));
          break;
        case "arrowdown":
          e.preventDefault();
          s.setVolume(Math.max(0, s.volume - 0.1));
          break;
        case "n":
          s.skipToNext();
          break;
        case "p":
          s.skipToPrevious();
          break;
      }
    }

    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, []);
}
