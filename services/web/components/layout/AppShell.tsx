/**
 * AppShell — the persistent chrome around authenticated pages.
 * Includes mobile hamburger menu and responsive layout.
 */
"use client";

import { useState } from "react";
import { Sidebar } from "./Sidebar";
import { Topbar } from "./Topbar";
import { MobileSidebar } from "./MobileSidebar";
import { PlayerBar } from "@/components/player/PlayerBar";

export function AppShell({ children }: { children: React.ReactNode }) {
  const [drawerOpen, setDrawerOpen] = useState(false);

  return (
    <div className="flex h-screen flex-col overflow-hidden bg-[var(--color-background)]">
      {/* Fixed topbar — passes hamburger handler */}
      <Topbar onMenuClick={() => setDrawerOpen(true)} />

      {/* Mobile drawer */}
      <MobileSidebar open={drawerOpen} onClose={() => setDrawerOpen(false)} />

      {/* Content row */}
      <div className="flex flex-1 overflow-hidden">
        {/* Desktop sidebar (hidden on mobile) */}
        <Sidebar />

        {/* Scrollable main area */}
        <main className="flex-1 overflow-y-auto px-4 py-4 sm:px-6 sm:py-6">{children}</main>
      </div>

      {/* Fixed bottom player bar */}
      <PlayerBar />
    </div>
  );
}
