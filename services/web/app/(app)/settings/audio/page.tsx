/**
 * settings/audio — ReplayGain + crossfade toggles.
 *
 * Both settings are pure localStorage toggles (no server round-trip);
 * ReplayGain changes are broadcast live via REPLAY_GAIN_CHANGE_EVENT so an
 * already-playing track picks up the new gain immediately (see useAudio.ts).
 * Crossfade only affects the next track transition (no live signal needed —
 * useAudio reads the toggle at swap time).
 */
"use client";

import { useEffect, useState } from "react";
import { Volume2, Waves } from "lucide-react";
import {
  isReplayGainEnabled,
  setReplayGainEnabled,
  dispatchReplayGainChange,
} from "@/lib/audio/replayGain";
import { isCrossfadeEnabled, setCrossfadeEnabled, CROSSFADE_SECONDS } from "@/lib/audio/crossfade";
import { cn } from "@/lib/utils";

export default function AudioSettingsPage() {
  const [replayGain, setReplayGain] = useState(false);
  const [crossfade, setCrossfade] = useState(false);

  useEffect(() => {
    setReplayGain(isReplayGainEnabled());
    setCrossfade(isCrossfadeEnabled());
  }, []);

  function toggleReplayGain() {
    const next = !replayGain;
    setReplayGain(next);
    setReplayGainEnabled(next);
    dispatchReplayGainChange(next);
  }

  function toggleCrossfade() {
    const next = !crossfade;
    setCrossfade(next);
    setCrossfadeEnabled(next);
  }

  return (
    <div className="mx-auto max-w-md space-y-6">
      <h1 className="text-2xl font-bold text-[var(--color-text)]">Audio</h1>

      <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)] divide-y divide-[var(--color-border)]">
        <ToggleRow
          icon={<Volume2 size={16} />}
          label="ReplayGain"
          description="Normalize loudness across tracks using each track's analyzed ReplayGain value."
          enabled={replayGain}
          onToggle={toggleReplayGain}
        />
        <ToggleRow
          icon={<Waves size={16} />}
          label="Crossfade"
          description={`Fade out the current track while fading in the next (${CROSSFADE_SECONDS}s ramp).`}
          enabled={crossfade}
          onToggle={toggleCrossfade}
        />
      </div>
    </div>
  );
}

function ToggleRow({
  icon,
  label,
  description,
  enabled,
  onToggle,
}: {
  icon: React.ReactNode;
  label: string;
  description: string;
  enabled: boolean;
  onToggle: () => void;
}) {
  return (
    <div className="flex items-start justify-between gap-4 px-4 py-3.5">
      <div className="flex items-start gap-3">
        <span className="mt-0.5 text-[var(--color-text-muted)]">{icon}</span>
        <div>
          <p className="text-sm font-medium text-[var(--color-text)]">{label}</p>
          <p className="text-xs text-[var(--color-text-muted)]">{description}</p>
        </div>
      </div>
      <button
        type="button"
        role="switch"
        aria-checked={enabled}
        onClick={onToggle}
        className={cn(
          "relative h-6 w-11 shrink-0 rounded-full transition-colors",
          enabled ? "bg-[var(--color-primary)]" : "bg-[var(--color-border)]"
        )}
      >
        <span
          className={cn(
            "absolute top-0.5 h-5 w-5 rounded-full bg-white transition-transform",
            enabled ? "translate-x-5" : "translate-x-0.5"
          )}
        />
      </button>
    </div>
  );
}
