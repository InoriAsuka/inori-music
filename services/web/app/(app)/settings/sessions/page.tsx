/**
 * settings/sessions — Active sessions
 * SessionView: { userId, expiresAt, createdAt }
 * No id, lastSeenAt, userAgent, or isCurrent in the spec.
 */
"use client";

import { useEffect, useState } from "react";
import { Monitor, Loader2 } from "lucide-react";
import { useAuthStore } from "@/store/auth";
import { authedApi } from "@/lib/api/client";
import { Skeleton } from "@/components/ui/Skeleton";

interface Session {
  userId: string;
  createdAt: string;
  expiresAt: string;
}

export default function SessionsPage() {
  const token = useAuthStore((s) => s.token);
  const [sessions, setSessions] = useState<Session[]>([]);
  const [loading, setLoading] = useState(true);
  const [revoking, setRevoking] = useState(false);

  async function load() {
    if (!token) return;
    const { data } = await authedApi(token).GET("/api/v1/me/sessions");
    if (data?.sessions) {
      setSessions(data.sessions.map((s) => ({
        userId: s.userId,
        createdAt: s.createdAt,
        expiresAt: s.expiresAt,
      })));
    }
    setLoading(false);
  }

  useEffect(() => { load(); }, [token]); // eslint-disable-line react-hooks/exhaustive-deps

  async function revokeOthers() {
    if (!token) return;
    setRevoking(true);
    await authedApi(token).POST("/api/v1/me/sessions/revoke-all");
    await load();
    setRevoking(false);
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Active Sessions</h1>
        <button
          onClick={revokeOthers}
          disabled={revoking || sessions.length <= 1}
          className="flex items-center gap-1.5 rounded-md border border-[var(--color-border)] px-3 py-1.5 text-sm hover:bg-[var(--color-muted)] disabled:opacity-50 transition-colors"
        >
          {revoking && <Loader2 size={12} className="animate-spin" />}
          Revoke other sessions
        </button>
      </div>

      <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] divide-y divide-[var(--color-border)]">
        {loading
          ? Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="flex items-center gap-3 px-4 py-3">
                <Skeleton className="h-8 w-8 rounded-md" />
                <div className="flex-1 space-y-1.5">
                  <Skeleton className="h-3.5 w-40" />
                  <Skeleton className="h-3 w-24" />
                </div>
              </div>
            ))
          : sessions.map((s, i) => (
              <div key={i} className="flex items-center gap-3 px-4 py-3">
                <div className="flex h-8 w-8 items-center justify-center rounded-md bg-[var(--color-muted)] text-[var(--color-muted-foreground)]">
                  <Monitor size={16} />
                </div>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium">Session</p>
                  <p className="text-xs text-[var(--color-muted-foreground)]">
                    Created {new Date(s.createdAt).toLocaleString()} ·
                    expires {new Date(s.expiresAt).toLocaleString()}
                  </p>
                </div>
              </div>
            ))}
      </div>
    </div>
  );
}
