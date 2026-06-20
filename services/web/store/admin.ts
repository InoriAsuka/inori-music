/**
 * Admin store — optional static bootstrap admin token.
 *
 * Admin routes can be accessed with either an admin user's viewer JWT or the
 * static bootstrap token. This store lets operators paste the static token
 * into the UI during bootstrap.
 */
"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";

interface AdminState {
  adminToken: string | null;
  setAdminToken: (token: string) => void;
  clearAdminToken: () => void;
}

export const useAdminStore = create<AdminState>()(
  persist(
    (set) => ({
      adminToken: null,
      setAdminToken(token) { set({ adminToken: token.trim() || null }); },
      clearAdminToken() { set({ adminToken: null }); },
    }),
    { name: "inori-admin" }
  )
);
