/**
 * Pure helpers for user-playlist track ordering.
 *
 * The backend stores a playlist body as `trackIds: string[]` where order is
 * significant and DUPLICATE ids are allowed (the same track may appear twice).
 * These helpers operate on that raw id array so the detail page can reorder /
 * remove by *position* without ever collapsing duplicates.
 *
 * Kept free of React / API imports so they're trivially unit-testable.
 */

/**
 * Move the id at `from` to `to`, preserving every other element (including
 * duplicates) and their relative order. Out-of-range indices return the input
 * unchanged. Mirrors the array-move used by the dnd-kit reorder handler.
 */
export function moveAt<T>(items: readonly T[], from: number, to: number): T[] {
  if (from === to) return items.slice();
  if (from < 0 || from >= items.length || to < 0 || to >= items.length) return items.slice();
  const next = items.slice();
  const [moved] = next.splice(from, 1);
  next.splice(to, 0, moved);
  return next;
}

export function moveId(ids: readonly string[], from: number, to: number): string[] {
  return moveAt(ids, from, to);
}

/**
 * Remove exactly the element at `index` (a specific occurrence), preserving all
 * other duplicates. Used when removing a track that is NOT the first occurrence —
 * the DELETE endpoint would drop the first match instead, so we persist the full
 * array via the replace-order endpoint. Out-of-range index returns input unchanged.
 */
export function removeAt<T>(items: readonly T[], index: number): T[] {
  if (index < 0 || index >= items.length) return items.slice();
  const next = items.slice();
  next.splice(index, 1);
  return next;
}

/**
 * Index of the first occurrence of `trackId`, or -1. When the row the user
 * clicked is the first occurrence, the DELETE `.../tracks/{trackId}` endpoint
 * (which removes the first match) does exactly the right thing; otherwise the
 * caller falls back to a replace-order PUT built from {@link removeAt}.
 */
export function firstIndexOf(ids: readonly string[], trackId: string): number {
  return ids.indexOf(trackId);
}

/**
 * One-based occurrence number and total for the item at `index`. This gives a
 * destructive confirmation enough context to distinguish duplicate rows.
 */
export function occurrenceAt<T>(items: readonly T[], index: number): { occurrence: number; total: number } | null {
  if (index < 0 || index >= items.length) return null;
  const item = items[index];
  let occurrence = 0;
  let total = 0;
  for (let i = 0; i < items.length; i += 1) {
    if (items[i] !== item) continue;
    total += 1;
    if (i <= index) occurrence += 1;
  }
  return { occurrence, total };
}
