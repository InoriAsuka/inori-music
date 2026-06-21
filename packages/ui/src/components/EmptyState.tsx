export function EmptyState({ title, description }: { title: string; description?: string }) {
  return (
    <div className="rounded-xl border border-dashed border-[var(--color-border)] p-10 text-center">
      <p className="font-medium text-[var(--color-text)]">{title}</p>
      {description && <p className="mt-1 text-sm text-[var(--color-text-muted)]">{description}</p>}
    </div>
  );
}
