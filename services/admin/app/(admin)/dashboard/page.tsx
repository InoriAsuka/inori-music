"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Users, Database, Upload, HardDrive, FileBox, Activity } from "lucide-react";
import { AdminTokenPanel } from "@/components/admin/AdminTokenPanel";
import { useAdminClient } from "@/hooks/useAdminClient";
import { useIsAdminLoggedIn } from "@/store/auth";

interface Stat {
  label: string;
  value?: number;
}

function StatCard({ label, value }: Stat) {
  return (
    <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)] p-4">
      <p className="text-xs text-[var(--color-text-muted)]">{label}</p>
      <p className="mt-2 font-mono text-2xl font-bold text-[var(--color-text)]">
        {value?.toLocaleString() ?? <span className="opacity-30">—</span>}
      </p>
    </div>
  );
}

const QUICK_LINKS = [
  { href: "/users", label: "Users", icon: Users, desc: "Manage accounts and sessions" },
  { href: "/catalog", label: "Catalog", icon: Database, desc: "Artists, albums, tracks, playlists" },
  { href: "/import", label: "Import", icon: Upload, desc: "Import from media objects" },
  { href: "/storage", label: "Storage", icon: HardDrive, desc: "Backends, health, capacity" },
  { href: "/media-objects", label: "Media Objects", icon: FileBox, desc: "Registry and lifecycle" },
  { href: "/history", label: "History", icon: Activity, desc: "Global playback stats" },
];

export default function DashboardPage() {
  const client = useAdminClient();
  const hasAccess = useIsAdminLoggedIn();
  const [catalog, setCatalog] = useState<{ artists: number; albums: number; tracks: number; playlists: number } | null>(
    null
  );
  const [hist, setHist] = useState<{ totalEvents: number; uniqueUsers: number; uniqueTracks: number } | null>(null);

  useEffect(() => {
    if (!client) return;
    Promise.all([client.GET("/api/v1/admin/catalog/stats"), client.GET("/api/v1/admin/history/stats")]).then(
      ([c, h]) => {
        if (c.data) setCatalog(c.data);
        if (h.data) setHist(h.data);
      }
    );
  }, [client]);

  return (
    <div className="space-y-8">
      <div>
        <h1 className="font-display text-2xl font-bold tracking-wider text-[var(--color-primary)]">DASHBOARD</h1>
        <p className="mt-1 text-sm text-[var(--color-text-muted)]">Inori Music administration overview</p>
      </div>

      {!hasAccess && (
        <div className="max-w-lg">
          <AdminTokenPanel />
        </div>
      )}

      {/* Catalog stats */}
      <section>
        <p className="mb-3 text-xs font-semibold uppercase tracking-wider text-[var(--color-text-muted)]">Catalog</p>
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
          <StatCard label="Artists" value={catalog?.artists} />
          <StatCard label="Albums" value={catalog?.albums} />
          <StatCard label="Tracks" value={catalog?.tracks} />
          <StatCard label="Playlists" value={catalog?.playlists} />
        </div>
      </section>

      {/* History stats */}
      <section>
        <p className="mb-3 text-xs font-semibold uppercase tracking-wider text-[var(--color-text-muted)]">Playback</p>
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
          <StatCard label="Play events" value={hist?.totalEvents} />
          <StatCard label="Unique users" value={hist?.uniqueUsers} />
          <StatCard label="Unique tracks" value={hist?.uniqueTracks} />
        </div>
      </section>

      {/* Quick links */}
      <section>
        <p className="mb-3 text-xs font-semibold uppercase tracking-wider text-[var(--color-text-muted)]">Sections</p>
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {QUICK_LINKS.map(({ href, label, icon: Icon, desc }) => (
            <Link
              key={href}
              href={href}
              className="group flex items-start gap-4 rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)] p-4 transition-colors hover:border-[var(--color-border-glow)] hover:bg-[var(--color-surface-raised)]"
            >
              <div className="mt-0.5 rounded-lg border border-[var(--color-border-glow)] bg-[var(--color-primary-dim)] p-2">
                <Icon size={16} className="text-[var(--color-primary)]" />
              </div>
              <div>
                <p className="font-semibold text-[var(--color-text)] group-hover:text-[var(--color-primary-hover)]">
                  {label}
                </p>
                <p className="text-xs text-[var(--color-text-muted)]">{desc}</p>
              </div>
            </Link>
          ))}
        </div>
      </section>
    </div>
  );
}
