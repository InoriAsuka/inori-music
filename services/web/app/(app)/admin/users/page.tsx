/**
 * Admin users — /admin/users
 */
"use client";

import { FormEvent, useEffect, useState } from "react";
import { UserPlus, Trash2, Power, KeyRound } from "lucide-react";
import { AdminTokenPanel } from "@/components/admin/AdminTokenPanel";
import { useAdminApi, useHasAdminAccess } from "@/hooks/useAdminApi";
import { PaginationBar, type OffsetPagination, offsetFromPage } from "@/components/ui/PaginationBar";
import { Skeleton } from "@/components/ui/Skeleton";
import { EmptyState } from "@/components/ui/EmptyState";

interface UserRow {
  id: string;
  username: string;
  role: "viewer" | "admin";
  enabled: boolean;
  createdAt: string;
}

const PAGE_SIZE = 50;

export default function AdminUsersPage() {
  const admin = useAdminApi();
  const hasAccess = useHasAdminAccess();
  const [users, setUsers] = useState<UserRow[]>([]);
  const [pagination, setPagination] = useState<OffsetPagination | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [form, setForm] = useState({ username: "", password: "", role: "viewer" as "viewer" | "admin" });

  async function load() {
    if (!admin) { setLoading(false); return; }
    setLoading(true);
    const { data } = await admin.GET("/api/v1/admin/users", {
      params: { query: { limit: PAGE_SIZE, offset: offsetFromPage(page, PAGE_SIZE) } },
    });
    if (data) {
      setUsers(data.users.map((u) => ({
        id: u.id,
        username: u.username,
        role: u.role as "viewer" | "admin",
        enabled: u.enabled,
        createdAt: u.createdAt,
      })));
      setPagination(data.pagination);
    }
    setLoading(false);
  }

  useEffect(() => { load(); }, [admin, page]); // eslint-disable-line react-hooks/exhaustive-deps

  async function createUser(e: FormEvent) {
    e.preventDefault();
    if (!admin) return;
    await admin.POST("/api/v1/admin/users", { body: form });
    setForm({ username: "", password: "", role: "viewer" });
    await load();
  }

  async function toggleEnabled(user: UserRow) {
    if (!admin) return;
    const path = user.enabled ? "/api/v1/admin/users/{id}/disable" : "/api/v1/admin/users/{id}/enable";
    await admin.POST(path, { params: { path: { id: user.id } } });
    await load();
  }

  async function deleteUser(id: string) {
    if (!admin || !window.confirm("Delete this user?")) return;
    await admin.DELETE("/api/v1/admin/users/{id}", { params: { path: { id } } });
    await load();
  }

  async function changeRole(user: UserRow, role: "viewer" | "admin") {
    if (!admin) return;
    await admin.PATCH("/api/v1/admin/users/{id}", { params: { path: { id: user.id } }, body: { role } });
    await load();
  }

  async function forcePassword(user: UserRow) {
    if (!admin) return;
    const password = window.prompt(`New password for ${user.username}`);
    if (!password) return;
    await admin.POST("/api/v1/admin/users/{id}/change-password", {
      params: { path: { id: user.id } },
      body: { newPassword: password },
    });
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Users</h1>
      <AdminTokenPanel />

      {!hasAccess && <EmptyState title="Admin access required" description="Sign in as an admin or paste a bootstrap token." />}

      {hasAccess && (
        <>
          <form onSubmit={createUser} className="grid gap-3 rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] p-4 sm:grid-cols-[1fr_1fr_140px_auto]">
            <input value={form.username} onChange={(e) => setForm((f) => ({ ...f, username: e.target.value }))} required placeholder="Username" className="rounded-md border border-[var(--color-border)] bg-[var(--color-background)] px-3 py-2 text-sm outline-none focus:border-[var(--color-primary)]" />
            <input value={form.password} onChange={(e) => setForm((f) => ({ ...f, password: e.target.value }))} required type="password" placeholder="Password" className="rounded-md border border-[var(--color-border)] bg-[var(--color-background)] px-3 py-2 text-sm outline-none focus:border-[var(--color-primary)]" />
            <select value={form.role} onChange={(e) => setForm((f) => ({ ...f, role: e.target.value as "viewer" | "admin" }))} className="rounded-md border border-[var(--color-border)] bg-[var(--color-background)] px-3 py-2 text-sm outline-none focus:border-[var(--color-primary)]">
              <option value="viewer">viewer</option>
              <option value="admin">admin</option>
            </select>
            <button className="flex items-center justify-center gap-1.5 rounded-md bg-[var(--color-primary)] px-3 py-2 text-sm font-semibold text-[var(--color-primary-foreground)] hover:opacity-90"><UserPlus size={14} /> Create</button>
          </form>

          <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)]">
            {loading ? Array.from({ length: 8 }).map((_, i) => <div key={i} className="border-b border-[var(--color-border)] px-4 py-3 last:border-0"><Skeleton className="h-6 w-full" /></div>) : users.map((u) => (
              <div key={u.id} className="flex items-center gap-3 border-b border-[var(--color-border)] px-4 py-3 last:border-0">
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium">{u.username} <span className="text-xs text-[var(--color-muted-foreground)]">({u.id})</span></p>
                  <p className="text-xs text-[var(--color-muted-foreground)]">Created {new Date(u.createdAt).toLocaleString()}</p>
                </div>
                <select value={u.role} onChange={(e) => changeRole(u, e.target.value as "viewer" | "admin")} className="rounded-md border border-[var(--color-border)] bg-[var(--color-background)] px-2 py-1 text-xs">
                  <option value="viewer">viewer</option>
                  <option value="admin">admin</option>
                </select>
                <span className={u.enabled ? "text-xs text-green-600" : "text-xs text-[var(--color-destructive)]"}>{u.enabled ? "enabled" : "disabled"}</span>
                <button onClick={() => toggleEnabled(u)} className="rounded p-1.5 text-[var(--color-muted-foreground)] hover:text-[var(--color-foreground)]" title={u.enabled ? "Disable" : "Enable"}><Power size={15} /></button>
                <button onClick={() => forcePassword(u)} className="rounded p-1.5 text-[var(--color-muted-foreground)] hover:text-[var(--color-foreground)]" title="Force password"><KeyRound size={15} /></button>
                <button onClick={() => deleteUser(u.id)} className="rounded p-1.5 text-[var(--color-muted-foreground)] hover:text-[var(--color-destructive)]" title="Delete"><Trash2 size={15} /></button>
              </div>
            ))}
          </div>

          {pagination && <PaginationBar pagination={pagination} onPageChange={setPage} />}
        </>
      )}
    </div>
  );
}
