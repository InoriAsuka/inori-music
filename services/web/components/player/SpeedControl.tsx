/**
 * SpeedControl — playback-speed menu for the player bar.
 *
 * Subscribes only to `speed` / `setSpeed` from the player store, so speed
 * changes never rerender the whole PlayerBar tree. The non-1× state stays
 * visibly indicated via a badge over the trigger.
 */
"use client";

import { Gauge } from "lucide-react";
import { usePlayerStore } from "@/store/player";
import { SPEED_PRESETS, formatSpeedLabel, isDefaultSpeed } from "@/lib/player/controls";
import {
  PlayerPopover,
  PlayerPopoverItem,
  type PlayerPopoverAlign,
  type PlayerPopoverPlacement,
} from "./PlayerPopover";

export function SpeedControl({
  placement = "above",
  align = "center",
}: {
  placement?: PlayerPopoverPlacement;
  align?: PlayerPopoverAlign;
}) {
  const speed = usePlayerStore((s) => s.speed);
  const setSpeed = usePlayerStore((s) => s.setSpeed);
  const nonDefault = !isDefaultSpeed(speed);

  return (
    <PlayerPopover
      label="Playback speed"
      active={nonDefault}
      badge={nonDefault ? formatSpeedLabel(speed) : undefined}
      placement={placement}
      align={align}
      title={<Gauge size={16} />}
    >
      <p className="px-3 pb-1 pt-1 text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-muted)]">
        Speed
      </p>
      {SPEED_PRESETS.map((preset, i) => (
        <PlayerPopoverItem
          key={preset}
          selected={speed === preset}
          // Focus the active preset on open, or the first item if the current
          // speed isn't one of the presets.
          autoFocus={speed === preset || (i === 0 && !SPEED_PRESETS.includes(speed as (typeof SPEED_PRESETS)[number]))}
          onSelect={() => setSpeed(preset)}
        >
          <span>{formatSpeedLabel(preset)}</span>
          {preset === 1 && <span className="text-xs text-[var(--color-text-muted)]">Normal</span>}
        </PlayerPopoverItem>
      ))}
    </PlayerPopover>
  );
}
