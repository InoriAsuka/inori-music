/**
 * Admin auth store — admin user JWT or bootstrap token.
 * Separate from inori-web auth; uses key "inori-admin-auth".
 */
"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";
import createClient from "openapi-fetch";
import type { paths } from "@/types/api.gen";

const baseUrl = typeof window === "undefined" ? (process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080") : "";

export interface AdminUser {
  id: string;
  username: string;
  role: "viewer" | "admin";
  createdAt: string;
}

interface AuthState {
  token: string | null;
  user: AdminUser | null;
  /** Bootstrap token (static, no JWT) */
  bootstrapToken: string | null;
  setSession: (token: string, user: AdminUser) => void;
  setBootstrapToken: (token: string) => void;
  clearBootstrapToken: () => void;
  clearSession: () => void;
  refreshUser: () => Promise<boolean>;
  /** Returns the best available token (JWT > bootstrap) */
  effectiveToken: () => string | null;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      user: null,
      bootstrapToken: null,

      setSession(token, user) {
        set({ token, user });
      },
      setBootstrapToken(token) {
        set({ bootstrapToken: token.trim() || null });
      },
      clearBootstrapToken() {
        set({ bootstrapToken: null });
      },
      clearSession() {
        set({ token: null, user: null });
      },

      effectiveToken() {
        const { token, user, bootstrapToken } = get();
        if (token && user?.role === "admin") return token;
        return bootstrapToken ?? null;
      },

      async refreshUser() {
        const { token } = get();
        if (!token) return false;
        const client = createClient<paths>({
          baseUrl,
          headers: { Authorization: `Bearer ${token}` },
        });
        const { data, error } = await client.GET("/api/v1/me");
        if (error || !data) {
          set({ token: null, user: null });
          return false;
        }
        set({
          user: {
            id: data.id,
            username: data.username,
            role: data.role as "viewer" | "admin",
            createdAt: data.createdAt,
          },
        });
        return true;
      },
    }),
    {
      name: "inori-admin-auth",
      partialize: (s) => ({ token: s.token, user: s.user, bootstrapToken: s.bootstrapToken }),
    }
  )
);

export const useEffectiveToken = () => useAuthStore((s) => s.effectiveToken());
export const useIsAdminLoggedIn = () =>
  useAuthStore((s) => {
    const { token, user, bootstrapToken } = s;
    return (token !== null && user?.role === "admin") || bootstrapToken !== null;
  });
