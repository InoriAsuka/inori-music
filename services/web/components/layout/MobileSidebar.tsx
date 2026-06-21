"use client";

import { useEffect } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { X, Music2, Users, Disc3, ListMusic, Heart, History, LayoutDashboard, Search } from "lucide-react";
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

interface MobileSidebarProps {
  open: boolean;
  onClose: () => void;
}

export function MobileSidebar({ open, onClose }: MobileSidebarProps) {
  const pathname = usePathname();

  useEffect(() => { onClose(); }, [pathname, onClose]); // eslint-disable-line react-hooks/exhaustive-deps
  useEffect(() => {
    if (open) document.body.style.overflow = "hidden";
    else document.body.style.overflow = "";
    return () => { document.body.style.overflow = ""; };
  }, [open]);

  return (
    <>
      <div
        className={cn("fixed inset-0 z-40 bg-black/60 transition-opacity md:hidden", open ? "opacity-100 pointer-events-auto" : "opacity-0 pointer-events-none")}
        onClick={onClose}
      />
      <aside className={cn("fixed inset-y-0 left-0 z-50 w-72 transform overflow-y-auto border-r border-[var(--color-border)] bg-[var(--color-surface)] transition-transform duration-200 md:hidden",
        open ? "translate-x-0" : "-translate-x-full")}>
        <div className="flex items-center justify-between border-b border-[var(--color-border)] px-4 py-3">
          <span className="font-semibold text-[var(--color-text)]">Inori Music</span>
          <button onClick={onClose} className="rounded-md p-1.5 hover:bg-[var(--color-surface-raised)]"><X size={18} /></button>
        </div>
        <nav className="flex flex-col gap-1 px-3 py-4">
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
    </>
  );
}

function SectionLabel({ children, className }: { children: React.ReactNode; className?: string }) {
  return <span className={cn("px-2 py-1 text-xs font-semibold uppercase tracking-wider text-[var(--color-text-muted)]", className)}>{children}</span>;
}

function NavLink({ item, active }: { item: { href: string; label: string; icon: React.ReactNode }; active: boolean }) {
  return (
    <Link href={item.href} className={cn("flex items-center gap-2 rounded-md px-3 py-2.5 text-sm transition-colors",
      active ? "bg-[var(--color-primary)] text-[var(--color-primary-fg)]" : "text-[var(--color-text)] hover:bg-[var(--color-surface-raised)]")}>
      {item.icon}{item.label}
    </Link>
  );
}
