import { useCallback, useEffect, useState } from "react";

/**
 * Persisted toggle for bilingual (translation) lyrics display.
 * Mirrors the Flutter `BilingualLyricsNotifier` pattern using localStorage.
 */
const KEY = "inori.lyrics.bilingual";

export function useBilingualToggle(): [enabled: boolean, setEnabled: (v: boolean) => void] {
  const [enabled, setEnabled] = useState(false);

  useEffect(() => {
    if (typeof window !== "undefined") {
      setEnabled(localStorage.getItem(KEY) === "true");
    }
  }, []);

  const toggle = useCallback((v: boolean) => {
    setEnabled(v);
    if (typeof window !== "undefined") {
      localStorage.setItem(KEY, v ? "true" : "false");
    }
  }, []);

  return [enabled, toggle];
}
