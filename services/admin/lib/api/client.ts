import createClient from "openapi-fetch";
import type { paths } from "@/types/api.gen";

const baseUrl = typeof window === "undefined"
  ? (process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080")
  : "";

export const api = createClient<paths>({ baseUrl });

/** Admin client using Bearer token (admin JWT or bootstrap token). */
export function adminClient(token: string) {
  return createClient<paths>({
    baseUrl,
    headers: { Authorization: `Bearer ${token}` },
  });
}
