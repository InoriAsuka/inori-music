/**
 * Albums list page — /albums
 * CatalogAlbum: { id, title, artistId, releaseYear, sortTitle, createdAt, updatedAt }
 * No denormalized artistName or trackCount.
 */
"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useAuthStore } from "@/store/auth";
import { authedApi } from "@/lib/api/client";
import { PaginationBar, type OffsetPagination, offsetFromPage } from "@/components/ui/PaginationBar";
import { CardSkeleton } from "@/components/ui/Skeleton";
import { Artwork } from "@/components/ui/Artwork";

interface Album {
  id: string;
  title: string;
  year?: number;
}

const PAGE_SIZE = 40;

export default function AlbumsPage() {
  const token = useAuthStore((s) => s.token);
  const [albums, setAlbums] = useState<Album[]>([]);
  const [pagination, setPagination] = useState<OffsetPagination | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!token) return;
    setLoading(true);
    authedApi(token)
      .GET("/api/v1/catalog/albums", {
        params: { query: { limit: PAGE_SIZE, offset: offsetFromPage(page, PAGE_SIZE) } },
      })
      .then(({ data }) => {
        if (data) {
          setAlbums(
            (data.albums ?? []).map((a) => ({
              id: a.id,
              title: a.title,
              year: a.releaseYear ?? undefined,
            }))
          );
          if (data.pagination) setPagination(data.pagination);
        }
      })
      .finally(() => setLoading(false));
  }, [token, page]);

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Albums</h1>
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
        {loading
          ? Array.from({ length: 20 }).map((_, i) => <CardSkeleton key={i} />)
          : albums.map((album) => (
              <Link
                key={album.id}
                href={`/albums/${album.id}`}
                className="group flex flex-col gap-2 rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] p-3 hover:bg-[var(--color-muted)] transition-colors"
              >
                <Artwork alt={album.title} size="lg" className="w-full h-auto aspect-square" />
                <p className="truncate font-medium text-sm">{album.title}</p>
                {album.year && <p className="truncate text-xs text-[var(--color-muted-foreground)]">{album.year}</p>}
              </Link>
            ))}
      </div>
      {pagination && <PaginationBar pagination={pagination} onPageChange={setPage} />}
    </div>
  );
}
