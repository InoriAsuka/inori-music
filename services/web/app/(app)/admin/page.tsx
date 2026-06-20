/**
 * Admin dashboard — /admin
 */
"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Users, Database, Upload, HardDrive, Activity, Shield } from "lucide-react";
import { AdminTokenPanel } from "@/components/admin/AdminTokenPanel";
import { useAdminApi, useHasAdminAccess } from "@/hooks/useAdminApi";
import { Skeleton } from "@/components/ui/Skeleton";

interface CatalogStats { artists: number; albums: number; tracks: number; playlists: number; }
interface HistoryStats { totalEvents: number; uniqueUsers: number; uniqueTracks: number; }

export default function AdminDashboardPage() {
  const admin = useAdminApi();
  const hasAccess = useHasAdminAccess();
  const [catalog, setCatalog] = useState<CatalogStats | null>(null);
  const [history, setHistory] = useState<HistoryStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!admin) { setLoading(false); return; }
    setLoading(true);
    Promise.all([
      admin.GET("/api/v1/admin/catalog/stats"),
      admin.GET("/api/v1/admin/history/stats"),
    ]).then(([catalogRes, historyRes]) => {
      if (catalogRes.data) setCatalog(catalogRes.data);
      if (historyRes.data) setHistory(historyRes.data);
    }).finally(() => setLoading(false));
  }, [admin]);

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <Shield size={22} className="text-[var(--color-primary)]" />
        <h1 className="text-2xl font-bold">Admin</h1>
      </div>

      <AdminTokenPanel />

      {!hasAccess && (
        <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] p-4 text-sm text-[var(--color-muted-foreground)]">
          Sign in as an admin user, or paste the bootstrap admin token above.
        </div>
      )}

      <div className="grid gap-4 sm:grid-cols-4">
        <Stat label="Artists" value={catalog?.artists} loading={loading} />
        <Stat label="Albums" value={catalog?.albums} loading={loading} />
        <Stat label="Tracks" value={catalog?.tracks} loading={loading} />
        <Stat label="Playlists" value={catalog?.playlists} loading={loading} />
      </div>

      <div className="grid gap-4 sm:grid-cols-3">
        <Stat label="Play events" value={history?.totalEvents} loading={loading} />
        <Stat label="Unique users" value={history?.uniqueUsers} loading={loading} />
        <Stat label="Unique tracks" value={history?.uniqueTracks} loading={loading} />
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
        <AdminCard href="/admin/users" icon={<Users size={22} />} title="Users" description="Create, disable, delete, and update users." />
        <AdminCard href="/admin/catalog" icon={<Database size={22} />} title="Catalog" description="Manage artists, albums, tracks." />
        <AdminCard href="/admin/import" icon={<Upload size={22} />} title="Import" description="Import tracks from media objects." />
        <AdminCard href="/admin/storage" icon={<HardDrive size={22} />} title="Storage" description="Probe and manage storage backends." />
        <AdminCard href="/admin/history" icon={<Activity size={22} />} title="History" description="Inspect and clean playback history." />
      </div>
    </div>
  );
}

function Stat({ label, value, loading }: { label: string; value?: number; loading: boolean }) {
  return (
    <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] p-4">
      <p className="text-sm text-[var(--color-muted-foreground)]">{label}</p>
      {loading ? <Skeleton className="mt-2 h-7 w-16" /> : <p className="mt-2 text-2xl font-bold">{value ?? 0}</p>}
    </div>
  );
}

function AdminCard({ href, icon, title, description }: { href: string; icon: React.ReactNode; title: string; description: string }) {
  return (
    <Link href={href} className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] p-4 hover:bg-[var(--color-muted)] transition-colors">
      <div className="mb-3 text-[var(--color-primary)]">{icon}</div>
      <h2 className="font-semibold">{title}</h2>
      <p className="mt-1 text-sm text-[var(--color-muted-foreground)]">{description}</p>
    </Link>
  );
}
