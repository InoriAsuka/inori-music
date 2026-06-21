"use client";

import { useMemo } from "react";
import { useAuthStore } from "@/store/auth";
import { adminClient } from "@/lib/api/client";

/** Returns an admin API client when a valid token is available. */
export function useAdminClient() {
  const token = useAuthStore((s) => s.effectiveToken());
  return useMemo(() => (token ? adminClient(token) : null), [token]);
}
