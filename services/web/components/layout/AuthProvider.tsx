/**
 * AuthProvider — bootstraps auth state from localStorage on first render.
 *
 * Sets a lightweight "inori_session" cookie so the middleware can check
 * route eligibility without reading localStorage (which is not available
 * in the Edge runtime).
 */
"use client";

import { useEffect } from "react";
import { useAuthStore } from "@/store/auth";

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const { token, refreshUser } = useAuthStore();

  useEffect(() => {
    if (!token) {
      // Ensure stale cookie is cleared if store is empty.
      document.cookie = "inori_session=; path=/; max-age=0";
      return;
    }

    // Set the hint cookie (1 day) so middleware can read it.
    document.cookie = "inori_session=1; path=/; max-age=86400; SameSite=Lax";

    // Validate the stored token is still good.
    refreshUser().then((valid) => {
      if (!valid) {
        document.cookie = "inori_session=; path=/; max-age=0";
      }
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

  // Keep cookie in sync when the session changes.
  useEffect(() => {
    const unsub = useAuthStore.subscribe((state) => {
      if (state.token) {
        document.cookie = "inori_session=1; path=/; max-age=86400; SameSite=Lax";
      } else {
        document.cookie = "inori_session=; path=/; max-age=0";
      }
    });
    return unsub;
  }, []);

  return <>{children}</>;
}
