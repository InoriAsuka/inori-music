/**
 * Artists list page — /artists
 * v1 API: GET /api/v1/catalog/artists?limit=&offset=
 * Response: { artists: CatalogArtist[], pagination: CatalogPaginationMeta }
 * CatalogArtist: { id, name, sortName, createdAt, updatedAt }
 * Pagination: { limit, offset, total, hasMore }
 */
"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useAuthStore } from "@/store/auth";
import { authedApi } from "@/lib/api/client";
import { PaginationBar, type OffsetPagination, offsetFromPage } from "@/components/ui/PaginationBar";
import { CardSkeleton } from "@/components/ui/Skeleton";

interface Artist {
  id: string;
  name: string;
}

const PAGE_SIZE = 40;

export default function ArtistsPage() {
  const token = useAuthStore((s) => s.token);
  const [artists, setArtists] = useState<Artist[]>([]);
  const [pagination, setPagination] = useState<OffsetPagination | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!token) return;
    setLoading(true);
    authedApi(token)
      .GET("/api/v1/catalog/artists", {
        params: { query: { limit: PAGE_SIZE, offset: offsetFromPage(page, PAGE_SIZE) } },
      })
      .then(({ data }) => {
        if (data) {
          setArtists((data.artists ?? []).map((a) => ({ id: a.id, name: a.name })));
          if (data.pagination) setPagination(data.pagination);
        }
      })
      .finally(() => setLoading(false));
  }, [token, page]);

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Artists</h1>

      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
        {loading
          ? Array.from({ length: 20 }).map((_, i) => <CardSkeleton key={i} />)
          : artists.map((artist) => (
              <Link
                key={artist.id}
                href={`/artists/${artist.id}`}
                className="group flex flex-col gap-2 rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] p-4 hover:bg-[var(--color-muted)] transition-colors"
              >
                <div className="flex aspect-square w-full items-center justify-center rounded-lg bg-[var(--color-muted)] text-4xl font-bold text-[var(--color-muted-foreground)]">
                  {artist.name.charAt(0).toUpperCase()}
                </div>
                <p className="truncate font-medium">{artist.name}</p>
              </Link>
            ))}
      </div>

      {pagination && <PaginationBar pagination={pagination} onPageChange={setPage} />}
    </div>
  );
}
