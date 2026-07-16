import { afterEach, describe, expect, it, vi } from "vitest";
import {
  appendUserPlaylistTrack,
  removeFirstUserPlaylistTrack,
  replaceUserPlaylistTracks,
  listUserPlaylists,
} from "./me-playlists";

/**
 * Wrapper contract tests for the empty-body track-mutation endpoints.
 *
 * The OpenAPI contract types `POST/DELETE .../tracks` with a `UserPlaylist`
 * body, but the Go server actually answers with an EMPTY 2xx (200 for append,
 * 204 for remove-first). openapi-fetch turns an empty/`Content-Length: 0` 2xx
 * into `{ data: undefined }`, so a `!data` guard would throw on every success
 * and (in the detail page) trigger an optimistic rollback. These tests pin the
 * wrappers to "success = no error", reproducing the real server behaviour via a
 * stubbed global fetch.
 */

function stubFetch(response: Response) {
  const spy = vi.fn(async () => response);
  vi.stubGlobal("fetch", spy);
  return spy;
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("appendUserPlaylistTrack (empty 200)", () => {
  it("resolves when the server returns an empty 200 body", async () => {
    // Content-Length: 0 is what net/http sends for w.WriteHeader(200) + no body.
    stubFetch(new Response(null, { status: 200, headers: { "Content-Length": "0" } }));
    await expect(appendUserPlaylistTrack("tok", "pl1", "trk1")).resolves.toBeUndefined();
  });

  it("throws when the server returns an error status", async () => {
    stubFetch(
      new Response(JSON.stringify({ error: { code: "not_found" } }), {
        status: 404,
        headers: { "Content-Type": "application/json" },
      })
    );
    await expect(appendUserPlaylistTrack("tok", "pl1", "trk1")).rejects.toBeDefined();
  });
});

describe("removeFirstUserPlaylistTrack (empty 204)", () => {
  it("resolves when the server returns 204 No Content", async () => {
    stubFetch(new Response(null, { status: 204 }));
    await expect(removeFirstUserPlaylistTrack("tok", "pl1", "trk1")).resolves.toBeUndefined();
  });

  it("throws when the server returns an error status", async () => {
    stubFetch(
      new Response(JSON.stringify({ error: { code: "forbidden" } }), {
        status: 403,
        headers: { "Content-Type": "application/json" },
      })
    );
    await expect(removeFirstUserPlaylistTrack("tok", "pl1", "trk1")).rejects.toBeDefined();
  });
});

describe("replaceUserPlaylistTracks (PUT returns a body)", () => {
  it("returns the updated playlist the server sends back", async () => {
    const body = {
      id: "pl1",
      userId: "u1",
      name: "Mix",
      description: "",
      trackIds: ["b", "a"],
      createdAt: "2026-01-01T00:00:00Z",
      updatedAt: "2026-01-02T00:00:00Z",
    };
    stubFetch(
      new Response(JSON.stringify(body), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      })
    );
    const updated = await replaceUserPlaylistTracks("tok", "pl1", ["b", "a"]);
    expect(updated.trackIds).toEqual(["b", "a"]);
  });
});

describe("listUserPlaylists", () => {
  it("returns an empty array when the server omits the playlists field", async () => {
    stubFetch(
      new Response(JSON.stringify({}), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      })
    );
    await expect(listUserPlaylists("tok")).resolves.toEqual([]);
  });
});
