/**
 * Login page — /login
 *
 * POST /api/v1/auth/login → LoginResponse { token, userId, expiresAt }
 * After login, fetches /me to get the full user object.
 */
"use client";

import { useState, type FormEvent, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Music2, Loader2 } from "lucide-react";
import { api, authedApi } from "@/lib/api/client";
import { useAuthStore } from "@/store/auth";
import type { AuthUser } from "@/store/auth";

function LoginForm() {
  const router = useRouter();
  const params = useSearchParams();
  const setSession = useAuthStore((s) => s.setSession);

  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      const { data, error: apiError } = await api.POST("/api/v1/auth/login", {
        body: { username, password },
      });

      if (apiError || !data) {
        setError("Invalid username or password.");
        return;
      }

      // LoginResponse only has { token, userId, expiresAt } — fetch /me for full profile.
      const { data: me } = await authedApi(data.token).GET("/api/v1/me");
      if (!me) {
        setError("Failed to load user profile.");
        return;
      }

      const user: AuthUser = {
        id: me.id,
        username: me.username,
        role: me.role as "viewer" | "admin",
        createdAt: me.createdAt,
      };

      setSession(data.token, user);

      const from = params.get("from") ?? "/";
      router.replace(from);
    } catch {
      setError("Network error. Please try again.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="aurora-veil flex min-h-dvh flex-col items-center justify-center bg-[var(--color-void)] px-4">
      <div className="relative z-10 w-full max-w-sm space-y-6">
        {/* Logo */}
        <div className="flex flex-col items-center gap-3">
          <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-[var(--color-primary)] glow-primary">
            <Music2 size={26} className="text-[var(--color-primary-ink)]" />
          </div>
          <h1 className="font-display text-2xl font-bold tracking-tight">Sign in to Inori Music</h1>
        </div>

        {/* Form */}
        <form
          onSubmit={handleSubmit}
          className="space-y-4 rounded-2xl border border-[var(--color-border)] bg-[var(--color-surface)] p-6"
        >
          <div className="space-y-1.5">
            <label htmlFor="username" className="text-sm font-medium">
              Username
            </label>
            <input
              id="username"
              type="text"
              autoComplete="username"
              required
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="w-full rounded-lg border border-[var(--color-border-strong)] bg-[var(--color-void)] px-3 py-2.5 text-sm outline-none focus:border-[var(--color-primary)] transition-colors"
              placeholder="your-username"
            />
          </div>

          <div className="space-y-1.5">
            <label htmlFor="password" className="text-sm font-medium">
              Password
            </label>
            <input
              id="password"
              type="password"
              autoComplete="current-password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full rounded-lg border border-[var(--color-border-strong)] bg-[var(--color-void)] px-3 py-2.5 text-sm outline-none focus:border-[var(--color-primary)] transition-colors"
              placeholder="••••••••"
            />
          </div>

          {error && (
            <p
              role="alert"
              className="rounded-lg border border-[var(--color-danger)]/40 bg-[var(--color-danger-dim)] px-3 py-2 text-sm text-[var(--color-danger)]"
            >
              {error}
            </p>
          )}

          <button
            type="submit"
            disabled={loading}
            className="flex w-full items-center justify-center gap-2 rounded-lg bg-[var(--color-primary)] px-4 py-2.5 text-sm font-semibold text-[var(--color-primary-ink)] transition-colors hover:bg-[var(--color-primary-hover)] disabled:opacity-60"
          >
            {loading && <Loader2 size={14} className="animate-spin" />}
            Sign in
          </button>
        </form>
      </div>
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
