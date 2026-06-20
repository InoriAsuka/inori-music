/**
 * AppShell — the persistent chrome around authenticated pages.
 *
 * Layout: vertical flex column.
 * ┌─────────────────────────────────┐
 * │  Topbar (fixed, 56px)           │
 * ├──────────┬──────────────────────┤
 * │ Sidebar  │  main content area   │
 * │ (240px)  │  (scrollable)        │
 * ├──────────┴──────────────────────┤
 * │  PlayerBar (fixed, 80px)        │
 * └─────────────────────────────────┘
 */
"use client";

import { Sidebar } from "./Sidebar";
import { Topbar } from "./Topbar";
import { PlayerBar } from "@/components/player/PlayerBar";

export function AppShell({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex h-screen flex-col overflow-hidden bg-[var(--color-background)]">
      {/* Fixed topbar */}
      <Topbar />

      {/* Content row */}
      <div className="flex flex-1 overflow-hidden">
        {/* Fixed sidebar */}
        <Sidebar />

        {/* Scrollable main area */}
        <main className="flex-1 overflow-y-auto px-6 py-6">{children}</main>
      </div>

      {/* Fixed bottom player bar */}
      <PlayerBar />
    </div>
  );
}
