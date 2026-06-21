"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  Music2, Users, Disc3, ListMusic,
  Heart, History, LayoutDashboard, Search,
} from "lucide-react";
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
  { href: "/library/favorites", label: "Favorites", icon: <Heart size={16} /> },
  { href: "/library/history", label: "History", icon: <History size={16} /> },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="hidden w-60 shrink-0 flex-col border-r border-[var(--color-border)] bg-[var(--color-surface)] md:flex">
      <nav className="flex flex-col gap-1 overflow-y-auto px-3 py-4">
        <SectionLabel>Library</SectionLabel>
        {NAV.map((item) => (
          <NavLink key={item.href} item={item} active={pathname === item.href} />
        ))}

        <SectionLabel className="mt-4">My Music</SectionLabel>
        {LIBRARY_NAV.map((item) => (
          <NavLink key={item.href} item={item} active={pathname.startsWith(item.href)} />
        ))}
      </nav>
    </aside>
  );
}

function SectionLabel({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <span className={cn("px-2 py-1 text-xs font-semibold uppercase tracking-wider text-[var(--color-text-muted)]", className)}>
      {children}
    </span>
  );
}

function NavLink({ item, active }: { item: { href: string; label: string; icon: React.ReactNode }; active: boolean }) {
  return (
    <Link
      href={item.href}
      className={cn(
        "flex items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors",
        active
          ? "bg-[var(--color-primary)] text-[var(--color-primary-fg)]"
          : "text-[var(--color-text)] hover:bg-[var(--color-surface-raised)]"
      )}
    >
      {item.icon}
      {item.label}
    </Link>
  );
}
