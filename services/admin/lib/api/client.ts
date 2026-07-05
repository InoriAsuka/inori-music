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

/**
 * Uploads lyrics (+ optional translation) as multipart/form-data. openapi-typescript
 * types `format: binary` request fields as `string`, so the typed client can't accept
 * File objects directly — this supplies a bodySerializer that builds real FormData from
 * the File values, per openapi-fetch's documented file-upload recipe.
 */
export function uploadTrackLyrics(
  client: ReturnType<typeof adminClient>,
  trackId: string,
  file: File,
  translation?: File,
) {
  return client.POST("/api/v1/catalog/tracks/{id}/lyrics", {
    params: { path: { id: trackId } },
    body: { file, translation } as unknown as { file: string; translation?: string },
    bodySerializer(body) {
      const { file, translation } = body as unknown as { file: File; translation?: File };
      const form = new FormData();
      form.append("file", file);
      if (translation) form.append("translation", translation);
      return form;
    },
  });
}
