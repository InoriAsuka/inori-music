"use client";

import { useState } from "react";
import { AdminSidebar } from "./AdminSidebar";
import { AdminTopbar } from "./AdminTopbar";

export function AdminShell({ children }: { children: React.ReactNode }) {
  const [collapsed, setCollapsed] = useState(false);

  return (
    <div className="flex h-screen flex-col overflow-hidden bg-[var(--color-void)]">
      <AdminTopbar onToggleSidebar={() => setCollapsed((c) => !c)} />
      <div className="flex flex-1 overflow-hidden">
        <AdminSidebar collapsed={collapsed} />
        <main className="flex-1 overflow-y-auto p-6 text-[var(--color-text)]">
          {children}
        </main>
      </div>
    </div>
  );
}
