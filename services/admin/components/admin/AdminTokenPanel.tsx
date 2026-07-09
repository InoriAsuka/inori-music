"use client";

import { useState } from "react";
import { KeyRound, X, CheckCircle } from "lucide-react";
import { useAuthStore } from "@/store/auth";

export function AdminTokenPanel() {
  const { bootstrapToken, setBootstrapToken, clearBootstrapToken } = useAuthStore();
  const [draft, setDraft] = useState(bootstrapToken ?? "");
  const [saved, setSaved] = useState(false);

  function save() {
    setBootstrapToken(draft);
    document.cookie = draft.trim()
      ? "inori_admin_session=1; path=/; max-age=86400; SameSite=Lax"
      : "inori_admin_session=; path=/; max-age=0";
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  }

  return (
    <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)] p-4">
      <div className="mb-3 flex items-center gap-2">
        <KeyRound size={15} className="text-[var(--color-primary)]" />
        <h2 className="text-sm font-semibold text-[var(--color-text)]">Bootstrap Admin Token</h2>
        {bootstrapToken && (
          <span className="rounded-full bg-[var(--color-success)] bg-opacity-15 px-2 py-0.5 text-[10px] font-semibold text-[var(--color-success)]">
            active
          </span>
        )}
      </div>
      <p className="mb-3 text-xs text-[var(--color-text-muted)]">
        Admin routes accept an admin user JWT or this static bootstrap token. Stored in browser localStorage only.
      </p>
      <div className="flex gap-2">
        <input
          type="password"
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && save()}
          placeholder="Paste bootstrap token…"
          className="min-w-0 flex-1 rounded-md border border-[var(--color-border)] bg-[var(--color-void)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)] transition-colors placeholder:text-[var(--color-text-muted)]"
        />
        <button
          type="button"
          onClick={save}
          className="flex items-center gap-1.5 rounded-md bg-[var(--color-primary)] px-3 py-2 text-sm font-semibold text-[var(--color-primary-fg)] hover:opacity-90 transition-opacity"
        >
          {saved ? <CheckCircle size={14} /> : "Save"}
        </button>
        {bootstrapToken && (
          <button
            type="button"
            onClick={() => {
              clearBootstrapToken();
              setDraft("");
            }}
            className="rounded-md border border-[var(--color-border)] p-2 text-[var(--color-text-muted)] hover:border-[var(--color-danger)] hover:text-[var(--color-danger)] transition-colors"
            title="Clear token"
          >
            <X size={14} />
          </button>
        )}
      </div>
    </div>
  );
}
