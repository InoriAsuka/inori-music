/**
 * SleepTimerControl — sleep-timer menu for the player bar.
 *
 * Subscribes only to the sleep-timer store, so its once-a-second countdown
 * updates re-render this small control and nothing else in the player tree.
 * Offers the exact 15/30/45/60-minute presets plus "After current track",
 * shows a live countdown / badge while armed, and can be cancelled.
 */
"use client";

import { Moon } from "lucide-react";
import { SLEEP_TIMER_PRESET_MINUTES, useSleepTimerStore } from "@/store/sleepTimer";
import { formatSleepCountdown } from "@/lib/player/controls";
import {
  PlayerPopover,
  PlayerPopoverItem,
  type PlayerPopoverAlign,
  type PlayerPopoverPlacement,
} from "./PlayerPopover";

export function SleepTimerControl({
  placement = "above",
  align = "right",
}: {
  placement?: PlayerPopoverPlacement;
  align?: PlayerPopoverAlign;
}) {
  const mode = useSleepTimerStore((s) => s.mode);
  const presetMinutes = useSleepTimerStore((s) => s.presetMinutes);
  const remainingMs = useSleepTimerStore((s) => s.remainingMs);
  const startFixed = useSleepTimerStore((s) => s.startFixed);
  const startAfterTrack = useSleepTimerStore((s) => s.startAfterTrack);
  const cancel = useSleepTimerStore((s) => s.cancel);

  const active = mode !== "off";
  const badge =
    mode === "fixed" ? formatSleepCountdown(remainingMs) : mode === "after-track" ? "•" : undefined;

  return (
    <PlayerPopover
      label={
        mode === "fixed"
          ? `Sleep timer — ${formatSleepCountdown(remainingMs)} remaining`
          : mode === "after-track"
            ? "Sleep timer — stops after current track"
            : "Sleep timer"
      }
      active={active}
      badge={badge}
      align={align}
      placement={placement}
      title={<Moon size={16} />}
    >
      <p className="px-3 pb-1 pt-1 text-[10px] font-bold uppercase tracking-widest text-[var(--color-text-muted)]">
        Sleep timer
      </p>
      {SLEEP_TIMER_PRESET_MINUTES.map((minutes, i) => {
        const selected = mode === "fixed" && presetMinutes === minutes;
        return (
          <PlayerPopoverItem
            key={minutes}
            selected={selected}
            autoFocus={selected || (i === 0 && mode !== "after-track")}
            onSelect={() => startFixed(minutes)}
          >
            <span>{minutes} minutes</span>
          </PlayerPopoverItem>
        );
      })}
      <PlayerPopoverItem
        selected={mode === "after-track"}
        autoFocus={mode === "after-track"}
        onSelect={startAfterTrack}
      >
        <span>After current track</span>
      </PlayerPopoverItem>

      {active && (
        <>
          <div className="my-1 h-px bg-[var(--color-border)]" />
          {mode === "fixed" && (
            <p className="px-3 py-1 text-center font-mono text-xs text-[var(--color-primary)]">
              {formatSleepCountdown(remainingMs)} left
            </p>
          )}
          <PlayerPopoverItem onSelect={cancel}>
            <span className="text-[var(--color-danger)]">Cancel timer</span>
          </PlayerPopoverItem>
        </>
      )}
    </PlayerPopover>
  );
}
