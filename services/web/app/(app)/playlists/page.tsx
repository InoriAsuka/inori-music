/**
 * Playlists list page — /playlists
 * Playlist: { id, name, description, trackIds, createdAt, updatedAt }
 * No trackCount directly, but trackIds.length gives it.
 */
"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { ListMusic } from "lucide-react";
import { useAuthStore } from "@/store/auth";
import { authedApi } from "@/lib/api/client";
import { PaginationBar, type OffsetPagination, offsetFromPage } from "@/components/ui/PaginationBar";
import { CardSkeleton } from "@/components/ui/Skeleton";

interface Playlist {
  id: string;
  name: string;
  trackCount: number;
}

const PAGE_SIZE = 40;

export default function PlaylistsPage() {
  const token = useAuthStore((s) => s.token);
  const [playlists, setPlaylists] = useState<Playlist[]>([]);
  const [pagination, setPagination] = useState<OffsetPagination | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!token) return;
    setLoading(true);
    authedApi(token)
      .GET("/api/v1/catalog/playlists", {
        params: { query: { limit: PAGE_SIZE, offset: offsetFromPage(page, PAGE_SIZE) } },
      })
      .then(({ data }) => {
        if (data) {
          setPlaylists((data.playlists ?? []).map((p) => ({
            id: p.id,
            name: p.name,
            trackCount: p.trackIds.length,
          })));
          if (data.pagination) setPagination(data.pagination);
        }
      })
      .finally(() => setLoading(false));
  }, [token, page]);

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Playlists</h1>
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
        {loading
          ? Array.from({ length: 12 }).map((_, i) => <CardSkeleton key={i} />)
          : playlists.map((pl) => (
              <Link
                key={pl.id}
                href={`/playlists/${pl.id}`}
                className="group flex flex-col gap-2 rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] p-4 hover:bg-[var(--color-muted)] transition-colors"
              >
                <div className="flex aspect-square w-full items-center justify-center rounded-lg bg-[var(--color-muted)] text-[var(--color-muted-foreground)]">
                  <ListMusic size={32} />
                </div>
                <p className="truncate font-medium text-sm">{pl.name}</p>
                <p className="text-xs text-[var(--color-muted-foreground)]">{pl.trackCount} tracks</p>
              </Link>
            ))}
      </div>
      {pagination && <PaginationBar pagination={pagination} onPageChange={setPage} />}
    </div>
  );
}
