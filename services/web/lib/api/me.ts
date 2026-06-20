import { authedApi } from "@/lib/api/client";
import type { AuthUser } from "@/store/auth";

/** Fetch /me with the given token and map to our AuthUser shape. */
export async function fetchMe(token: string): Promise<AuthUser | null> {
  const { data } = await authedApi(token).GET("/api/v1/me");
  if (!data) return null;
  return {
    id: data.id,
    username: data.username,
    role: data.role as "viewer" | "admin",
    createdAt: data.createdAt,
  };
}
