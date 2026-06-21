/**
 * AppShell — persistent chrome around authenticated player pages.
 */
"use client";

import { useState } from "react";
import { Sidebar } from "./Sidebar";
import { Topbar } from "./Topbar";
import { MobileSidebar } from "./MobileSidebar";
import { BottomNav } from "./BottomNav";
import { PlayerBar } from "@/components/player/PlayerBar";
import { usePlayerKeyboard } from "@/hooks/usePlayerKeyboard";

export function AppShell({ children }: { children: React.ReactNode }) {
  const [drawerOpen, setDrawerOpen] = useState(false);
  usePlayerKeyboard();

  return (
    <div className="flex h-screen flex-col overflow-hidden bg-[var(--color-void)]">
      <Topbar onMenuClick={() => setDrawerOpen(true)} />
      <MobileSidebar open={drawerOpen} onClose={() => setDrawerOpen(false)} />

      <div className="flex flex-1 overflow-hidden">
        <Sidebar />
        <main className="flex-1 overflow-y-auto px-4 py-4 text-[var(--color-text)] sm:px-6 sm:py-6">
          {children}
        </main>
      </div>

      <PlayerBar />
      <BottomNav />
    </div>
  );
}
