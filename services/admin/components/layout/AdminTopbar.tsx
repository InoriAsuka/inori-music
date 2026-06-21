"use client";

import { useRouter } from "next/navigation";
import { Menu, LogOut, Shield } from "lucide-react";
import { useAuthStore } from "@/store/auth";
import { adminClient } from "@/lib/api/client";

export function AdminTopbar({ onToggleSidebar }: { onToggleSidebar: () => void }) {
  const router = useRouter();
  const { token, user, clearSession } = useAuthStore();

  async function handleLogout() {
    if (token) {
      await adminClient(token).POST("/api/v1/auth/logout").catch(() => {});
    }
    clearSession();
    document.cookie = "inori_admin_session=; path=/; max-age=0";
    router.push("/login");
  }

  return (
    <header className="flex h-14 shrink-0 items-center justify-between border-b border-[var(--color-border)] bg-[var(--color-surface)] px-4">
      <div className="flex items-center gap-3">
        <button
          onClick={onToggleSidebar}
          className="rounded-md p-1.5 text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-raised)] hover:text-[var(--color-text)]"
        >
          <Menu size={18} />
        </button>
        <div className="flex items-center gap-2">
          <Shield size={18} className="text-[var(--color-primary)]" />
          <span className="font-display text-sm font-bold tracking-widest text-[var(--color-primary)]">
            INORI ADMIN
          </span>
        </div>
      </div>

      <div className="flex items-center gap-3">
        {user && (
          <span className="text-sm text-[var(--color-text-secondary)]">
            {user.username}
            <span className="ml-2 rounded-full bg-[var(--color-primary-dim)] border border-[var(--color-primary)] px-1.5 py-0.5 text-[10px] font-semibold text-[var(--color-primary)]">
              admin
            </span>
          </span>
        )}
        <button
          onClick={handleLogout}
          className="rounded-md p-1.5 text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-raised)] hover:text-[var(--color-danger)]"
          title="Log out"
        >
          <LogOut size={16} />
        </button>
      </div>
    </header>
  );
}
