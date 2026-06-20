"use client";

import { useState } from "react";
import { KeyRound, X } from "lucide-react";
import { useAdminStore } from "@/store/admin";

export function AdminTokenPanel() {
  const { adminToken, setAdminToken, clearAdminToken } = useAdminStore();
  const [draft, setDraft] = useState(adminToken ?? "");

  return (
    <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] p-4">
      <div className="mb-3 flex items-center gap-2">
        <KeyRound size={16} className="text-[var(--color-muted-foreground)]" />
        <h2 className="font-semibold">Admin token</h2>
      </div>
      <p className="mb-3 text-sm text-[var(--color-muted-foreground)]">
        Admin routes use an admin user JWT, or this static bootstrap token if configured.
      </p>
      <div className="flex gap-2">
        <input
          type="password"
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          placeholder="Paste bootstrap admin token"
          className="min-w-0 flex-1 rounded-md border border-[var(--color-border)] bg-[var(--color-background)] px-3 py-2 text-sm outline-none focus:border-[var(--color-primary)]"
        />
        <button
          onClick={() => setAdminToken(draft)}
          className="rounded-md bg-[var(--color-primary)] px-3 py-2 text-sm font-semibold text-[var(--color-primary-foreground)] hover:opacity-90"
        >
          Save
        </button>
        {adminToken && (
          <button
            onClick={() => { clearAdminToken(); setDraft(""); }}
            className="rounded-md border border-[var(--color-border)] px-3 py-2 text-sm hover:bg-[var(--color-muted)]"
            title="Clear token"
          >
            <X size={14} />
          </button>
        )}
      </div>
    </div>
  );
}
