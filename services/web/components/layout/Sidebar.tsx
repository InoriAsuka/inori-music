"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Music2, Users, Disc3, ListMusic, Heart, History, LayoutDashboard, Search, Settings, Volume2 } from "lucide-react";
import { cn } from "@/lib/utils";

const NAV = [
  { href: "/", label: "Home", icon: <LayoutDashboard size={16} /> },
  { href: "/artists", label: "Artists", icon: <Users size={16} /> },
  { href: "/albums", label: "Albums", icon: <Disc3 size={16} /> },
  { href: "/tracks", label: "Tracks", icon: <Music2 size={16} /> },
  { href: "/playlists", label: "Playlists", icon: <ListMusic size={16} /> },
  { href: "/search", label: "Search", icon: <Search size={16} /> },
];

const LIBRARY_NAV = [
  { href: "/library/playlists", label: "My Playlists", icon: <ListMusic size={16} /> },
  { href: "/library/favorites", label: "Favorites", icon: <Heart size={16} /> },
  { href: "/library/history", label: "History", icon: <History size={16} /> },
];

const SETTINGS_NAV = [
  { href: "/settings/security", label: "Password", icon: <Settings size={16} /> },
  { href: "/settings/sessions", label: "Sessions", icon: <Settings size={16} /> },
  { href: "/settings/language", label: "Language", icon: <Settings size={16} /> },
  { href: "/settings/audio", label: "Audio", icon: <Volume2 size={16} /> },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="hidden w-60 shrink-0 flex-col border-r border-[var(--color-border)] bg-[var(--color-void)] md:flex">
      <nav className="flex flex-col gap-1 overflow-y-auto px-3 py-4">
        <SectionLabel>Library</SectionLabel>
        {NAV.map((item) => (
          <NavLink key={item.href} item={item} active={pathname === item.href} />
        ))}

        <SectionLabel className="mt-4">My Music</SectionLabel>
        {LIBRARY_NAV.map((item) => (
          <NavLink key={item.href} item={item} active={pathname.startsWith(item.href)} />
        ))}

        <SectionLabel className="mt-4">Settings</SectionLabel>
        {SETTINGS_NAV.map((item) => (
          <NavLink key={item.href} item={item} active={pathname === item.href} />
        ))}
      </nav>
    </aside>
  );
}

function SectionLabel({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <span
      className={cn(
        "px-2 py-1 font-display text-[11px] font-bold uppercase tracking-[0.14em] text-[var(--color-text-muted)]",
        className
      )}
    >
      {children}
    </span>
  );
}

function NavLink({ item, active }: { item: { href: string; label: string; icon: React.ReactNode }; active: boolean }) {
  return (
    <Link
      href={item.href}
      aria-current={active ? "page" : undefined}
      className={cn(
        "relative flex items-center gap-2.5 rounded-lg py-2.5 pl-4 pr-3 text-sm transition-colors",
        active
          ? "bg-[var(--color-primary-dim)] font-medium text-[var(--color-primary-on-dim)]"
          : "text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-raised)] hover:text-[var(--color-text)]"
      )}
    >
      {active && (
        <span
          aria-hidden="true"
          className="absolute left-0 top-1/2 h-5 w-[3px] -translate-y-1/2 rounded-full bg-[var(--color-primary)]"
        />
      )}
      {item.icon}
      {item.label}
    </Link>
  );
}
