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
    <header className="flex h-14 shrink-0 items-center justify-between border-b border-[var(--color-border)] bg-[var(--color-card)] px-4">
      {/* Logo + hamburger */}
      <div className="flex items-center gap-2">
        {onMenuClick && (
          <button
            type="button"
            onClick={onMenuClick}
            className="rounded-md p-1.5 hover:bg-[var(--color-muted)] md:hidden"
            title="Menu"
          >
            <Menu size={18} />
          </button>
        )}
        <Link href="/" className="flex items-center gap-2 font-semibold">
          <Music2 size={20} className="text-[var(--color-primary)]" />
          <span className="text-sm">Inori Music</span>
        </Link>
      </div>

      {/* Search shortcut — clicking navigates too */}
      <Link
        href="/search"
        className="hidden flex-1 max-w-sm mx-8 items-center gap-2 rounded-md border border-[var(--color-border)] bg-[var(--color-muted)] px-3 py-1.5 text-sm text-[var(--color-muted-foreground)] hover:bg-[var(--color-border)] transition-colors md:flex"
      >
        <Search size={14} />
        Search tracks, artists…
        <kbd className="ml-auto rounded border border-[var(--color-border)] bg-[var(--color-card)] px-1 text-xs opacity-60">
          ⌘K
        </kbd>
      </Link>

      {/* User menu */}
      <div className="flex items-center gap-3">
        {user && (
          <span className="flex items-center gap-1.5 text-sm text-[var(--color-muted-foreground)]">
            <User size={14} />
            {user.username}
            {user.role === "admin" && (
              <span className="rounded-full bg-[var(--color-primary)] px-1.5 py-0.5 text-[10px] font-semibold text-[var(--color-primary-foreground)]">
                admin
              </span>
            )}
          </span>
        )}
        <button
          type="button"
          onClick={handleLogout}
          className="flex items-center gap-1 rounded-md p-1.5 text-sm text-[var(--color-muted-foreground)] hover:bg-[var(--color-muted)] transition-colors"
          title="Log out"
        >
          <LogOut size={16} />
        </button>
      </div>
    </header>
  );
}
