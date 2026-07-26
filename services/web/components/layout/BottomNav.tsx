"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Home, Search, Heart, History, Settings } from "lucide-react";
import { cn } from "@/lib/utils";

const TABS = [
  { href: "/", label: "Home", icon: Home },
  { href: "/search", label: "Search", icon: Search },
  { href: "/library/favorites", label: "Fav", icon: Heart },
  { href: "/library/history", label: "History", icon: History },
  { href: "/settings/security", label: "Settings", icon: Settings },
];

export function BottomNav() {
  const pathname = usePathname();
  return (
    <nav className="grid h-14 shrink-0 grid-cols-5 border-t border-[var(--color-border)] bg-[var(--color-void)] sm:hidden">
      {TABS.map(({ href, label, icon: Icon }) => {
        const active = pathname === href || pathname.startsWith(`${href}/`);
        return (
          <Link
            key={href}
            href={href}
            aria-current={active ? "page" : undefined}
            className={cn(
              "relative flex flex-col items-center justify-center gap-1 text-[10px] transition-colors",
              active ? "text-[var(--color-primary)]" : "text-[var(--color-text-muted)]"
            )}
          >
            {active && (
              <span
                aria-hidden="true"
                className="absolute top-0 h-[3px] w-8 rounded-full bg-[var(--color-primary)]"
              />
            )}
            <Icon size={18} />
            {label}
          </Link>
        );
      })}
    </nav>
  );
}
