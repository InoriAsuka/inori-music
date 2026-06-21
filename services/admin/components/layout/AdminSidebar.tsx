"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  LayoutDashboard, Users, Database, Upload, HardDrive,
  Activity, FileBox, ChevronLeft, ChevronRight,
} from "lucide-react";
import { cn } from "@/lib/utils";

const NAV = [
  { href: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { href: "/users", label: "Users", icon: Users },
  { href: "/catalog", label: "Catalog", icon: Database },
  { href: "/import", label: "Import", icon: Upload },
  { href: "/storage", label: "Storage", icon: HardDrive },
  { href: "/media-objects", label: "Media Objects", icon: FileBox },
  { href: "/history", label: "History", icon: Activity },
];

export function AdminSidebar({ collapsed }: { collapsed: boolean }) {
  const pathname = usePathname();

  return (
    <aside
      className={cn(
        "flex flex-col border-r border-[var(--color-border)] bg-[var(--color-surface)] transition-all duration-200",
        collapsed ? "w-14" : "w-52"
      )}
    >
      <nav className="flex flex-col gap-0.5 p-2 pt-3">
        {NAV.map(({ href, label, icon: Icon }) => {
          const active = pathname === href || pathname.startsWith(href + "/");
          return (
            <Link
              key={href}
              href={href}
              title={collapsed ? label : undefined}
              className={cn(
                "flex items-center gap-3 rounded-md px-3 py-2.5 text-sm font-medium transition-colors",
                active
                  ? "nav-active"
                  : "text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-raised)] hover:text-[var(--color-text)]"
              )}
            >
              <Icon size={16} className="shrink-0" />
              {!collapsed && <span className="truncate">{label}</span>}
            </Link>
          );
        })}
      </nav>
    </aside>
  );
}
