"use client";

import { useState, type FormEvent, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Shield, Loader2 } from "lucide-react";
import { api, adminClient } from "@/lib/api/client";
import { useAuthStore } from "@/store/auth";

function LoginForm() {
  const router = useRouter();
  const params = useSearchParams();
  const { setSession, setBootstrapToken } = useAuthStore();

  const [tab, setTab] = useState<"jwt" | "token">("jwt");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [bootstrapDraft, setBootstrapDraft] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleJwtLogin(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const { data } = await api.POST("/api/v1/auth/login", { body: { username, password } });
      if (!data) {
        setError("Invalid credentials.");
        return;
      }
      const { data: me } = await adminClient(data.token).GET("/api/v1/me");
      if (!me) {
        setError("Failed to load profile.");
        return;
      }
      if (me.role !== "admin") {
        setError("This account does not have admin access.");
        return;
      }
      setSession(data.token, { id: me.id, username: me.username, role: "admin", createdAt: me.createdAt });
      document.cookie = "inori_admin_session=1; path=/; max-age=86400; SameSite=Lax";
      router.replace(params.get("from") ?? "/dashboard");
    } catch {
      setError("Network error.");
    } finally {
      setLoading(false);
    }
  }

  function handleTokenLogin(e: FormEvent) {
    e.preventDefault();
    if (!bootstrapDraft.trim()) {
      setError("Token is required.");
      return;
    }
    setBootstrapToken(bootstrapDraft.trim());
    document.cookie = "inori_admin_session=1; path=/; max-age=86400; SameSite=Lax";
    router.replace(params.get("from") ?? "/dashboard");
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-[var(--color-void)] px-4 scanlines">
      <div className="w-full max-w-sm space-y-6">
        <div className="flex flex-col items-center gap-3">
          <div className="flex h-14 w-14 items-center justify-center rounded-xl border border-[var(--color-border-glow)] bg-[var(--color-primary-dim)] glow-primary">
            <Shield size={26} className="text-[var(--color-primary)]" />
          </div>
          <h1 className="font-display text-lg font-bold tracking-widest text-[var(--color-primary)]">INORI ADMIN</h1>
          <p className="text-xs text-[var(--color-text-muted)]">Management Console</p>
        </div>

        {/* Tab switcher */}
        <div className="flex gap-1 rounded-lg border border-[var(--color-border)] bg-[var(--color-surface)] p-1">
          {(["jwt", "token"] as const).map((t) => (
            <button
              type="button"
              key={t}
              onClick={() => {
                setTab(t);
                setError(null);
              }}
              className={
                tab === t
                  ? "flex-1 rounded-md bg-[var(--color-primary)] py-1.5 text-xs font-semibold text-[var(--color-primary-fg)]"
                  : "flex-1 rounded-md py-1.5 text-xs text-[var(--color-text-secondary)] hover:text-[var(--color-text)]"
              }
            >
              {t === "jwt" ? "Admin Account" : "Bootstrap Token"}
            </button>
          ))}
        </div>

        <form
          onSubmit={tab === "jwt" ? handleJwtLogin : handleTokenLogin}
          className="space-y-4 rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)] p-6"
        >
          {tab === "jwt" ? (
            <>
              <Field
                id="username"
                label="Username"
                value={username}
                onChange={setUsername}
                autoComplete="username"
                placeholder="admin"
              />
              <Field
                id="password"
                label="Password"
                value={password}
                onChange={setPassword}
                autoComplete="current-password"
                type="password"
                placeholder="••••••••"
              />
            </>
          ) : (
            <div className="space-y-1.5">
              <label htmlFor="bootstrap-token" className="text-sm font-medium text-[var(--color-text)]">
                Bootstrap Token
              </label>
              <input
                id="bootstrap-token"
                type="password"
                value={bootstrapDraft}
                onChange={(e) => setBootstrapDraft(e.target.value)}
                placeholder="Paste static admin token"
                required
                className="w-full rounded-md border border-[var(--color-border)] bg-[var(--color-void)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)] transition-colors placeholder:text-[var(--color-text-muted)]"
              />
              <p className="text-xs text-[var(--color-text-muted)]">
                The static token configured via INORI_ADMIN_TOKEN.
              </p>
            </div>
          )}

          {error && (
            <p className="rounded-md border border-[var(--color-danger)] bg-[var(--color-danger-dim)] px-3 py-2 text-sm text-[var(--color-danger)]">
              {error}
            </p>
          )}

          <button
            type="submit"
            disabled={loading}
            className="flex w-full items-center justify-center gap-2 rounded-md bg-[var(--color-primary)] px-4 py-2.5 text-sm font-semibold text-[var(--color-primary-fg)] hover:opacity-90 disabled:opacity-60 transition-opacity"
          >
            {loading && <Loader2 size={14} className="animate-spin" />}
            {tab === "jwt" ? "Sign in" : "Use token"}
          </button>
        </form>
      </div>
    </div>
  );
}

function Field({
  id,
  label,
  value,
  onChange,
  type = "text",
  autoComplete,
  placeholder,
}: {
  id: string;
  label: string;
  value: string;
  onChange: (v: string) => void;
  type?: string;
  autoComplete?: string;
  placeholder?: string;
}) {
  return (
    <div className="space-y-1.5">
      <label htmlFor={id} className="text-sm font-medium text-[var(--color-text)]">
        {label}
      </label>
      <input
        id={id}
        type={type}
        autoComplete={autoComplete}
        required
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="w-full rounded-md border border-[var(--color-border)] bg-[var(--color-void)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)] transition-colors placeholder:text-[var(--color-text-muted)]"
      />
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense>
      <LoginForm />
    </Suspense>
  );
}
