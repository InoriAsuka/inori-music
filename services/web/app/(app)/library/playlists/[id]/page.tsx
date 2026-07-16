/**
 * User playlist detail — /library/playlists/[id]
 *
 * Data model note on DUPLICATES: `UserPlaylist.trackIds` is the authoritative
 * ordered list and MAY contain the same id more than once. The tracks endpoint
 * returns metadata but we treat `trackIds` as the source of truth for order and
 * multiplicity — display rows are built by joining each trackId (by position)
 * against a metadata map, so duplicates and order always survive.
 *
 * Each row carries a stable `uid` (`<id>#<occurrence>`) so dnd-kit can identify
 * duplicate-id rows uniquely; reordering permutes rows then persists the derived
 * id order via PUT (replace-order). Removing a row removes that exact occurrence:
 * the first occurrence uses the DELETE endpoint (which drops the first match),
 * any later occurrence is removed via a replace-order PUT built from removeAt().
 */
"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Play, GripVertical, Trash2 } from "lucide-react";
import {
  DndContext,
  closestCenter,
  type DragEndEvent,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import {
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { useAuthStore } from "@/store/auth";
import { authedApi } from "@/lib/api/client";
import { Skeleton, TrackRowSkeleton } from "@/components/ui/Skeleton";
import { EmptyState } from "@/components/ui/EmptyState";
import { usePlayerStore } from "@/store/player";
import { TrackRow } from "@/components/ui/TrackRow";
import { cn, formatDuration } from "@/lib/utils";
import { resolveArtistNames } from "@/lib/api/catalog-cache";
import { removeAt, firstIndexOf, moveAt, occurrenceAt } from "@/lib/playlist-order";
import {
  getUserPlaylist,
  getUserPlaylistTracks,
  replaceUserPlaylistTracks,
  removeFirstUserPlaylistTrack,
  UserPlaylistNotFoundError,
} from "@/lib/api/me-playlists";

interface Row {
  uid: string; // `<trackId>#<occurrence>` — stable across reorder, unique per duplicate
  id: string;
  title: string;
  artistName: string;
  durationMs: number;
}

/** Assign a per-occurrence uid so duplicate ids stay individually addressable. */
function buildRows(trackIds: string[], meta: Map<string, { title: string; artistName: string; durationMs: number }>): Row[] {
  const seen = new Map<string, number>();
  return trackIds.map((id) => {
    const n = seen.get(id) ?? 0;
    seen.set(id, n + 1);
    const m = meta.get(id);
    return {
      uid: `${id}#${n}`,
      id,
      title: m?.title ?? id,
      artistName: m?.artistName ?? "",
      durationMs: m?.durationMs ?? 0,
    };
  });
}

export default function UserPlaylistDetailPage() {
  const { id } = useParams<{ id: string }>();
  const token = useAuthStore((s) => s.token);
  const playQueue = usePlayerStore((s) => s.playQueue);

  const [name, setName] = useState("");
  const [rows, setRows] = useState<Row[]>([]);
  const [loading, setLoading] = useState(true);
  const [notFound, setNotFound] = useState(false);
  const [error, setError] = useState(false);
  const [persisting, setPersisting] = useState(false);
  const rowsRef = useRef<Row[]>([]);
  const mutatingRef = useRef(false);

  const commitRows = useCallback((next: Row[]) => {
    rowsRef.current = next;
    setRows(next);
  }, []);

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates })
  );

  const load = useCallback(async () => {
    if (!token || !id) return;
    setLoading(true);
    setNotFound(false);
    setError(false);
    try {
      const pl = await getUserPlaylist(token, id);
      setName(pl.name);
      const tracks = await getUserPlaylistTracks(token, id);
      const client = authedApi(token);
      const names = await resolveArtistNames(client, tracks.map((t) => t.artistId));
      const meta = new Map(
        tracks.map((t) => [
          t.id,
          { title: t.title, artistName: names.get(t.artistId) ?? "", durationMs: t.durationMs ?? 0 },
        ])
      );
      commitRows(buildRows(pl.trackIds, meta));
    } catch (cause) {
      if (cause instanceof UserPlaylistNotFoundError) setNotFound(true);
      else setError(true);
    } finally {
      setLoading(false);
    }
  }, [token, id, commitRows]);

  useEffect(() => {
    load();
  }, [load]);

  const totalMs = useMemo(() => rows.reduce((sum, r) => sum + r.durationMs, 0), [rows]);

  function playFrom(idx: number) {
    const q = rows.map((r) => ({
      id: r.id,
      title: r.title,
      artistName: r.artistName,
      albumTitle: "",
      durationSeconds: Math.round(r.durationMs / 1000),
      playbackUrl: "",
    }));
    playQueue(q, idx);
  }

  async function persistOrder(nextRows: Row[]) {
    if (!token || !id || mutatingRef.current) return;
    const prev = rowsRef.current;
    mutatingRef.current = true;
    commitRows(nextRows);
    setPersisting(true);
    setError(false);
    try {
      await replaceUserPlaylistTracks(token, id, nextRows.map((r) => r.id));
    } catch {
      commitRows(prev);
      setError(true);
    } finally {
      mutatingRef.current = false;
      setPersisting(false);
    }
  }

  function onDragEnd(event: DragEndEvent) {
    const { active, over } = event;
    if (!over || active.id === over.id || mutatingRef.current) return;
    const currentRows = rowsRef.current;
    const from = currentRows.findIndex((r) => r.uid === active.id);
    const to = currentRows.findIndex((r) => r.uid === over.id);
    if (from < 0 || to < 0) return;
    void persistOrder(moveAt(currentRows, from, to));
  }

  async function removeRow(uid: string) {
    if (!token || !id || mutatingRef.current) return;
    const currentRows = rowsRef.current;
    const position = currentRows.findIndex((row) => row.uid === uid);
    const row = currentRows[position];
    if (!row) return;
    const ids = currentRows.map((r) => r.id);
    const occurrence = occurrenceAt(ids, position);
    if (!occurrence) return;
    if (!window.confirm(`Remove "${row.title}" (occurrence ${occurrence.occurrence} of ${occurrence.total}) from "${name}"?`)) return;
    const isFirst = firstIndexOf(ids, row.id) === position;
    const nextRows = removeAt(currentRows, position);
    mutatingRef.current = true;
    commitRows(nextRows);
    setPersisting(true);
    setError(false);
    try {
      if (isFirst) {
        await removeFirstUserPlaylistTrack(token, id, row.id);
      } else {
        await replaceUserPlaylistTracks(token, id, removeAt(ids, position));
      }
    } catch {
      commitRows(currentRows);
      setError(true);
    } finally {
      mutatingRef.current = false;
      setPersisting(false);
    }
  }

  // ── Render ────────────────────────────────────────────────────────────
  if (notFound) {
    return (
      <div className="space-y-6">
        <BackLink />
        <EmptyState title="Playlist not found" description="This playlist may have been deleted." />
      </div>
    );
  }

  return (
    <div className="space-y-8">
      <BackLink />

      <div className="flex items-end gap-4">
        <div>
          {loading ? <Skeleton className="h-8 w-48 mb-2" /> : <h1 className="text-3xl font-bold">{name}</h1>}
          <p className="text-sm text-[var(--color-muted-foreground)]">
            {rows.length} {rows.length === 1 ? "track" : "tracks"}
            {rows.length > 0 && <> · {formatDuration(totalMs / 1000)}</>}
          </p>
          <button
            type="button"
            onClick={() => playFrom(0)}
            disabled={rows.length === 0}
            className="mt-3 flex items-center gap-2 rounded-full bg-[var(--color-primary)] px-6 py-2 text-sm font-semibold text-[var(--color-primary-foreground)] hover:opacity-90 disabled:opacity-50 transition-opacity"
          >
            <Play size={14} fill="currentColor" /> Play all
          </button>
        </div>
      </div>

      {error && (
        <div className="flex items-center justify-between rounded-lg border border-[var(--color-destructive)]/40 bg-[var(--color-destructive)]/10 px-4 py-3 text-sm text-[var(--color-destructive)]">
          <span>Something went wrong. Your last change may not have been saved.</span>
          <button type="button" onClick={load} className="font-medium underline">
            Reload
          </button>
        </div>
      )}

      <div
        className={cn(
          "rounded-xl border border-[var(--color-border)] bg-[var(--color-card)]",
          persisting && "opacity-70"
        )}
      >
        {loading ? (
          Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="border-b border-[var(--color-border)] px-4 last:border-0">
              <TrackRowSkeleton />
            </div>
          ))
        ) : rows.length === 0 ? (
          <div className="p-8 text-center text-sm text-[var(--color-muted-foreground)]">
            This playlist is empty. Add tracks from any track list’s ⋯ menu.
          </div>
        ) : (
          <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={onDragEnd}>
            <SortableContext items={rows.map((r) => r.uid)} strategy={verticalListSortingStrategy}>
              {rows.map((row, idx) => (
                <SortableTrackRow
                  key={row.uid}
                  row={row}
                  position={idx}
                  onPlay={() => playFrom(idx)}
                  onRemove={() => removeRow(row.uid)}
                />
              ))}
            </SortableContext>
          </DndContext>
        )}
      </div>
    </div>
  );
}

function BackLink() {
  return (
    <Link
      href="/library/playlists"
      className="flex items-center gap-1 text-sm text-[var(--color-muted-foreground)] hover:text-[var(--color-foreground)]"
    >
      <ArrowLeft size={14} /> My Playlists
    </Link>
  );
}

function SortableTrackRow({
  row,
  position,
  onPlay,
  onRemove,
}: {
  row: Row;
  position: number;
  onPlay: () => void;
  onRemove: () => void;
}) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: row.uid });
  const style = { transform: CSS.Transform.toString(transform), transition };

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={cn(isDragging && "relative z-10 bg-[var(--color-muted)] shadow-lg")}
    >
      <TrackRow
        track={row}
        index={position + 1}
        onPlay={onPlay}
        showFavorite={false}
        showMenu={false}
        className="px-4"
        leading={
          <button
            type="button"
            {...attributes}
            {...listeners}
            aria-label={`Reorder ${row.title}`}
            className="shrink-0 cursor-grab touch-none rounded p-1 text-[var(--color-muted-foreground)] hover:text-[var(--color-foreground)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-[var(--color-primary)]"
          >
            <GripVertical size={15} />
          </button>
        }
        trailing={
          <button
            type="button"
            onClick={onRemove}
            aria-label={`Remove ${row.title}`}
            className="shrink-0 rounded p-1.5 text-[var(--color-muted-foreground)] hover:text-[var(--color-destructive)]"
          >
            <Trash2 size={15} />
          </button>
        }
      />
    </div>
  );
}
