"use client";

import { useMemo } from "react";
import { bearerAdminApi } from "@/lib/api/admin";
import { useAuthStore } from "@/store/auth";
import { useAdminStore } from "@/store/admin";

/**
 * Returns an admin API client when possible.
 * Prefer admin user's JWT; fall back to static bootstrap token from AdminTokenPanel.
 */
export function useAdminApi() {
  const viewerToken = useAuthStore((s) => s.token);
  const user = useAuthStore((s) => s.user);
  const bootstrapToken = useAdminStore((s) => s.adminToken);
  const token = user?.role === "admin" ? viewerToken : bootstrapToken;

  return useMemo(() => {
    if (!token) return null;
    return bearerAdminApi(token);
  }, [token]);
}

export function useHasAdminAccess() {
  const viewerToken = useAuthStore((s) => s.token);
  const user = useAuthStore((s) => s.user);
  const bootstrapToken = useAdminStore((s) => s.adminToken);
  return (user?.role === "admin" && !!viewerToken) || !!bootstrapToken;
}
