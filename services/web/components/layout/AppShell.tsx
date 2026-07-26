/**
 * AppShell — persistent chrome around authenticated player pages.
 */
"use client";

import { useState } from "react";
import { Sidebar } from "./Sidebar";
import { Topbar } from "./Topbar";
import { MobileSidebar } from "./MobileSidebar";
import { BottomNav } from "./BottomNav";
import { PetalDrift } from "./PetalDrift";
import { PlayerBar } from "@/components/player/PlayerBar";
import { usePlayerKeyboard } from "@/hooks/usePlayerKeyboard";

export function AppShell({ children }: { children: React.ReactNode }) {
  const [drawerOpen, setDrawerOpen] = useState(false);
  usePlayerKeyboard();

  return (
    <div className="relative flex h-screen flex-col overflow-hidden bg-[var(--color-void)]">
      <PetalDrift />

      <Topbar onMenuClick={() => setDrawerOpen(true)} />
      <MobileSidebar open={drawerOpen} onClose={() => setDrawerOpen(false)} />

      <div className="relative z-10 flex flex-1 overflow-hidden">
        <Sidebar />
        <main className="aurora-veil flex-1 overflow-y-auto text-[var(--color-text)]">
          <div className="relative z-10 px-4 py-6 sm:px-8 sm:py-8">{children}</div>
        </main>
      </div>

      <div className="relative z-10">
        <PlayerBar />
        <BottomNav />
      </div>
    </div>
  );
}
