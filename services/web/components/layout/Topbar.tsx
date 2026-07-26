"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { Music2, LogOut, User, Search, Menu } from "lucide-react";
import { useAuthStore } from "@/store/auth";
import { authedApi } from "@/lib/api/client";

export function Topbar({ onMenuClick }: { onMenuClick?: () => void }) {
  const router = useRouter();
  const { token, user, clearSession } = useAuthStore();

  // ⌘K / Ctrl+K → /search
  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        router.push("/search");
      }
    }
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [router]);

  async function handleLogout() {
    if (token) {
      const client = authedApi(token);
      await client.POST("/api/v1/auth/logout").catch(() => {});
    }
    clearSession();
    router.push("/login");
  }

  return (
    <header className="relative z-10 flex h-14 shrink-0 items-center justify-between border-b border-[var(--color-border)] bg-[var(--color-void)] px-4">
      {/* Logo + hamburger */}
      <div className="flex items-center gap-2">
        {onMenuClick && (
          <button
            type="button"
            onClick={onMenuClick}
            className="rounded-lg p-2 text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-raised)] hover:text-[var(--color-text)] transition-colors md:hidden"
            title="Menu"
          >
            <Menu size={18} />
          </button>
        )}
        <Link href="/" className="flex items-center gap-2">
          <Music2 size={20} className="text-[var(--color-primary)]" />
          <span className="font-display text-base font-bold tracking-tight">Inori Music</span>
        </Link>
      </div>

      {/* Search shortcut — clicking navigates too */}
      <Link
        href="/search"
        className="hidden flex-1 max-w-sm mx-8 items-center gap-2 rounded-full border border-[var(--color-border-strong)] bg-[var(--color-surface-raised)] px-4 py-2 text-sm text-[var(--color-text-secondary)] hover:border-[var(--color-primary)] hover:text-[var(--color-text)] transition-colors md:flex"
      >
        <Search size={14} />
        Search tracks, artists…
        <kbd className="ml-auto rounded border border-[var(--color-border-strong)] bg-[var(--color-surface)] px-1.5 py-0.5 font-mono text-[10px] text-[var(--color-text-muted)]">
          ⌘K
        </kbd>
      </Link>

      {/* User menu */}
      <div className="flex items-center gap-3">
        {user && (
          <span className="flex items-center gap-1.5 text-sm text-[var(--color-text-secondary)]">
            <User size={14} />
            {user.username}
            {user.role === "admin" && (
              <span className="rounded-full bg-[var(--color-primary)] px-2 py-0.5 text-[10px] font-semibold text-[var(--color-primary-ink)]">
                admin
              </span>
            )}
          </span>
        )}
        <button
          type="button"
          onClick={handleLogout}
          className="flex items-center gap-1 rounded-lg p-2 text-sm text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-raised)] hover:text-[var(--color-danger)] transition-colors"
          title="Log out"
        >
          <LogOut size={16} />
        </button>
      </div>
    </header>
  );
}
