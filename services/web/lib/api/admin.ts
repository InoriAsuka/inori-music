import createClient from "openapi-fetch";
import type { paths } from "@/types/api.gen";

const baseUrl = typeof window === "undefined"
  ? (process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080")
  : "";

/** Admin API client using Bearer token (admin JWT or static bootstrap token). */
export function bearerAdminApi(token: string) {
  return createClient<paths>({
    baseUrl,
    headers: { Authorization: `Bearer ${token}` },
  });
}
