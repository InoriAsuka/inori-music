"use client";

import { useEffect } from "react";
import { useAuthStore } from "@/store/auth";

/** Syncs the admin session cookie hint for middleware. */
export function AdminAuthProvider({ children }: { children: React.ReactNode }) {
  const { token, refreshUser } = useAuthStore();

  useEffect(() => {
    if (!token) {
      document.cookie = "inori_admin_session=; path=/; max-age=0";
      return;
    }
    document.cookie = `inori_admin_session=1; path=/; max-age=86400; SameSite=Lax`;
    refreshUser().then((valid) => {
      if (!valid) document.cookie = "inori_admin_session=; path=/; max-age=0";
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

  useEffect(() => {
    const unsub = useAuthStore.subscribe((state) => {
      if (state.token || state.bootstrapToken) {
        document.cookie = `inori_admin_session=1; path=/; max-age=86400; SameSite=Lax`;
      } else {
        document.cookie = "inori_admin_session=; path=/; max-age=0";
      }
    });
    return unsub;
  }, []);

  return <>{children}</>;
}
