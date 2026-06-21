"use client";

import { FormEvent, useEffect, useState } from "react";
import { UserPlus, Trash2, Power, KeyRound, ChevronLeft, ChevronRight } from "lucide-react";
import { useAdminClient } from "@/hooks/useAdminClient";
import { formatDateTime } from "@/lib/utils";

interface UserRow { id: string; username: string; role: string; enabled: boolean; createdAt: string; }

const PAGE = 50;

export default function UsersPage() {
  const client = useAdminClient();
  const [users, setUsers] = useState<UserRow[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(false);
  const [form, setForm] = useState({ username: "", password: "", role: "viewer" });

  async function load() {
    if (!client) return;
    setLoading(true);
    const { data } = await client.GET("/api/v1/admin/users", { params: { query: { limit: PAGE, offset } } });
    if (data) {
      setUsers(data.users.map((u) => ({ id: u.id, username: u.username, role: u.role, enabled: u.enabled, createdAt: u.createdAt })));
      setTotal(data.pagination.total);
    }
    setLoading(false);
  }

  useEffect(() => { load(); }, [client, offset]); // eslint-disable-line react-hooks/exhaustive-deps

  async function create(e: FormEvent) {
    e.preventDefault();
    if (!client) return;
    await client.POST("/api/v1/admin/users", { body: { username: form.username, password: form.password, role: form.role as "viewer" | "admin" } });
    setForm({ username: "", password: "", role: "viewer" });
    await load();
  }

  async function toggle(u: UserRow) {
    if (!client) return;
    const path = u.enabled ? "/api/v1/admin/users/{id}/disable" : "/api/v1/admin/users/{id}/enable";
    await client.POST(path, { params: { path: { id: u.id } } });
    await load();
  }

  async function del(id: string) {
    if (!client || !window.confirm("Delete user?")) return;
    await client.DELETE("/api/v1/admin/users/{id}", { params: { path: { id } } });
    await load();
  }

  async function forcePwd(u: UserRow) {
    const pwd = window.prompt(`New password for ${u.username}`);
    if (!pwd || !client) return;
    await client.POST("/api/v1/admin/users/{id}/change-password", { params: { path: { id: u.id } }, body: { newPassword: pwd } });
  }

  async function patchRole(u: UserRow, role: "viewer" | "admin") {
    if (!client) return;
    await client.PATCH("/api/v1/admin/users/{id}", { params: { path: { id: u.id } }, body: { role } });
    await load();
  }

  const totalPages = Math.ceil(total / PAGE);
  const page = Math.floor(offset / PAGE) + 1;

  return (
    <div className="space-y-6">
      <h1 className="font-display text-xl font-bold tracking-wider text-[var(--color-primary)]">USERS</h1>

      {/* Create form */}
      <form onSubmit={create} className="grid gap-3 rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)] p-4 sm:grid-cols-[1fr_1fr_140px_auto]">
        {[
          { key: "username", placeholder: "Username", type: "text" },
          { key: "password", placeholder: "Password", type: "password" },
        ].map(({ key, placeholder, type }) => (
          <input key={key} type={type} placeholder={placeholder} required value={form[key as "username" | "password"]}
            onChange={(e) => setForm((f) => ({ ...f, [key]: e.target.value }))}
            className="rounded-md border border-[var(--color-border)] bg-[var(--color-void)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)] placeholder:text-[var(--color-text-muted)]"
          />
        ))}
        <select value={form.role} onChange={(e) => setForm((f) => ({ ...f, role: e.target.value }))}
          className="rounded-md border border-[var(--color-border)] bg-[var(--color-void)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
        >
          <option value="viewer">viewer</option>
          <option value="admin">admin</option>
        </select>
        <button className="flex items-center justify-center gap-1.5 rounded-md bg-[var(--color-primary)] px-3 py-2 text-sm font-semibold text-[var(--color-primary-fg)] hover:opacity-90">
          <UserPlus size={14} /> Create
        </button>
      </form>

      {/* Table */}
      <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)] overflow-hidden">
        <div className="grid grid-cols-[1fr_100px_80px_80px_auto] gap-0 border-b border-[var(--color-border)] px-4 py-2 text-xs font-semibold uppercase tracking-wider text-[var(--color-text-muted)]">
          <span>User</span><span>Role</span><span>Status</span><span>Created</span><span></span>
        </div>
        {loading ? (
          <div className="px-4 py-8 text-center text-sm text-[var(--color-text-muted)]">Loading…</div>
        ) : users.map((u) => (
          <div key={u.id} className="grid grid-cols-[1fr_100px_80px_80px_auto] items-center gap-0 border-b border-[var(--color-border)] px-4 py-3 last:border-0 hover:bg-[var(--color-surface-raised)] transition-colors">
            <div>
              <p className="text-sm font-medium text-[var(--color-text)]">{u.username}</p>
              <p className="font-mono text-xs text-[var(--color-text-muted)]">{u.id}</p>
            </div>
            <select value={u.role} onChange={(e) => patchRole(u, e.target.value as "viewer" | "admin")}
              className="w-24 rounded border border-[var(--color-border)] bg-[var(--color-void)] px-2 py-1 text-xs text-[var(--color-text)]"
            >
              <option value="viewer">viewer</option>
              <option value="admin">admin</option>
            </select>
            <span className={u.enabled ? "text-xs text-[var(--color-success)]" : "text-xs text-[var(--color-text-muted)]"}>
              {u.enabled ? "active" : "disabled"}
            </span>
            <span className="text-xs text-[var(--color-text-muted)]">{new Date(u.createdAt).toLocaleDateString()}</span>
            <div className="flex items-center gap-1">
              <Btn onClick={() => toggle(u)} title={u.enabled ? "Disable" : "Enable"}><Power size={13} /></Btn>
              <Btn onClick={() => forcePwd(u)} title="Force password"><KeyRound size={13} /></Btn>
              <Btn onClick={() => del(u.id)} title="Delete" danger><Trash2 size={13} /></Btn>
            </div>
          </div>
        ))}
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between text-sm text-[var(--color-text-muted)]">
          <span>{total} users</span>
          <div className="flex items-center gap-2">
            <button onClick={() => setOffset(Math.max(0, offset - PAGE))} disabled={page <= 1} className="rounded p-1 hover:bg-[var(--color-surface-raised)] disabled:opacity-30"><ChevronLeft size={16} /></button>
            <span>{page} / {totalPages}</span>
            <button onClick={() => setOffset(offset + PAGE)} disabled={page >= totalPages} className="rounded p-1 hover:bg-[var(--color-surface-raised)] disabled:opacity-30"><ChevronRight size={16} /></button>
          </div>
        </div>
      )}
    </div>
  );
}

function Btn({ children, onClick, title, danger }: { children: React.ReactNode; onClick: () => void; title?: string; danger?: boolean }) {
  return (
    <button onClick={onClick} title={title}
      className={`rounded p-1.5 transition-colors ${danger ? "text-[var(--color-text-muted)] hover:text-[var(--color-danger)]" : "text-[var(--color-text-muted)] hover:text-[var(--color-text)]"}`}
    >{children}</button>
  );
}
