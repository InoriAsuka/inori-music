/**
 * Auth store — viewer session state.
 *
 * Persists the JWT token to localStorage so the user stays logged in
 * across page refreshes without a server-side session.
 */
"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";
import { authedApi } from "@/lib/api/client";

export interface AuthUser {
  id: string;
  username: string;
  role: "viewer" | "admin";
  createdAt: string;
}

interface AuthState {
  token: string | null;
  user: AuthUser | null;
  /** Set after a successful login. */
  setSession: (token: string, user: AuthUser) => void;
  /** Clear the session (logout). */
  clearSession: () => void;
  /** Refresh /me to get the latest user info. Returns false if token is invalid. */
  refreshUser: () => Promise<boolean>;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      user: null,

      setSession(token, user) {
        set({ token, user });
      },

      clearSession() {
        set({ token: null, user: null });
      },

      async refreshUser() {
        const { token } = get();
        if (!token) return false;
        const client = authedApi(token);
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
      name: "inori-auth",
      // Only persist token and user; methods are not serializable.
      partialize: (s) => ({ token: s.token, user: s.user }),
    }
  )
);

/** Convenience selector — true when a viewer session is active. */
export const useIsLoggedIn = () => useAuthStore((s) => s.token !== null);
/** Convenience selector — the current viewer token. */
export const useToken = () => useAuthStore((s) => s.token);
