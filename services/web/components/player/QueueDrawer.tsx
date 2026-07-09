"use client";

import { AnimatePresence, motion } from "motion/react";
import { DndContext, closestCenter, type DragEndEvent } from "@dnd-kit/core";
import { SortableContext, useSortable, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { GripVertical, X, Trash2 } from "lucide-react";
import { usePlayerStore, type QueueTrack } from "@/store/player";
import { cn, formatDuration } from "@/lib/utils";

export function QueueDrawer({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { queue, currentIndex, skipToIndex, removeFromQueue, reorderQueue } = usePlayerStore();

  function onDragEnd(event: DragEndEvent) {
    const { active, over } = event;
    if (!over || active.id === over.id) return;
    const from = queue.findIndex((t) => t.id === active.id);
    const to = queue.findIndex((t) => t.id === over.id);
    if (from >= 0 && to >= 0) reorderQueue(from, to);
  }

  return (
    <AnimatePresence>
      {open && (
        <>
          <motion.div
            className="fixed inset-0 z-40 bg-black/50"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
          />
          <motion.aside
            className="fixed right-0 top-0 z-50 h-full w-full max-w-md border-l border-[var(--color-border)] bg-[var(--color-surface)] shadow-2xl"
            initial={{ x: "100%" }}
            animate={{ x: 0 }}
            exit={{ x: "100%" }}
            transition={{ duration: 0.18 }}
          >
            <div className="flex items-center justify-between border-b border-[var(--color-border)] px-4 py-3">
              <h2 className="font-display text-sm font-bold tracking-widest text-[var(--color-primary)]">QUEUE</h2>
              <button
                type="button"
                onClick={onClose}
                className="rounded p-1.5 text-[var(--color-text-muted)] hover:bg-[var(--color-surface-raised)] hover:text-[var(--color-text)]"
              >
                <X size={16} />
              </button>
            </div>

            <div className="h-[calc(100%-52px)] overflow-y-auto p-2">
              {queue.length === 0 ? (
                <p className="py-12 text-center text-sm text-[var(--color-text-muted)]">Queue is empty.</p>
              ) : (
                <DndContext collisionDetection={closestCenter} onDragEnd={onDragEnd}>
                  <SortableContext items={queue.map((t) => t.id)} strategy={verticalListSortingStrategy}>
                    {queue.map((track, idx) => (
                      <SortableQueueItem
                        key={track.id}
                        track={track}
                        index={idx}
                        active={idx === currentIndex}
                        onPlay={() => skipToIndex(idx)}
                        onRemove={() => removeFromQueue(idx)}
                      />
                    ))}
                  </SortableContext>
                </DndContext>
              )}
            </div>
          </motion.aside>
        </>
      )}
    </AnimatePresence>
  );
}

function SortableQueueItem({
  track,
  index,
  active,
  onPlay,
  onRemove,
}: {
  track: QueueTrack;
  index: number;
  active: boolean;
  onPlay: () => void;
  onRemove: () => void;
}) {
  const { attributes, listeners, setNodeRef, transform, transition } = useSortable({ id: track.id });
  const style = { transform: CSS.Transform.toString(transform), transition };

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={cn(
        "mb-1 flex items-center gap-2 rounded-lg border border-transparent px-2 py-2 transition-colors",
        active
          ? "border-[var(--color-primary)] bg-[var(--color-primary-dim)]"
          : "hover:bg-[var(--color-surface-raised)]"
      )}
    >
      <button type="button" {...attributes} {...listeners} className="cursor-grab text-[var(--color-text-muted)]">
        <GripVertical size={14} />
      </button>
      <button type="button" onClick={onPlay} className="min-w-0 flex-1 text-left">
        <p className="truncate text-sm font-medium text-[var(--color-text)]">
          {index + 1}. {track.title}
        </p>
        <p className="truncate text-xs text-[var(--color-text-muted)]">{track.artistName || track.albumTitle}</p>
      </button>
      <span className="font-mono text-xs text-[var(--color-text-muted)]">{formatDuration(track.durationSeconds)}</span>
      <button
        type="button"
        onClick={onRemove}
        className="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-danger)]"
      >
        <Trash2 size={13} />
      </button>
    </div>
  );
}
