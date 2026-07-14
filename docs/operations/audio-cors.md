# Web ReplayGain / WebAudio CORS Requirements

## Scope

v5.2.0 adds a WebAudio gain pipeline to the web client (`services/web/lib/audio/audioGraph.ts`)
so ReplayGain and crossfade can process audio samples in-browser. This only
concerns the **web** client — the mobile client uses `just_audio`'s native
gain control and is unaffected.

## Why CORS matters here

`MediaElementAudioSourceNode` (the WebAudio node that taps an `<audio>`
element's decoded samples) requires the element's `crossOrigin` attribute to
be set and the audio response to carry a matching
`Access-Control-Allow-Origin` header. If either is missing, the browser
"taints" the media element: the `<audio>` tag itself plays normally, but
WebAudio can no longer read its samples — `createMediaElementSource` may
throw, or the resulting graph may produce silence, depending on the browser.

The web client sets `audio.crossOrigin = "anonymous"` on every playback
element. It does **not** send credentials (cookies) with the audio fetch, so
the server does not need `Access-Control-Allow-Credentials`.

## Requirements by backend

- **Same-origin playback** (`streamUrl`, proxied through nginx / the Next.js
  rewrite in dev): no CORS headers are needed — same-origin requests aren't
  subject to CORS at all. This is the default for `local`/`nfs`/`smb`
  backends and is already CORS-clean today.
- **S3-compatible presigned URLs** (`presignedUrl`, see
  `docs/architecture/storage-backends.md`): the bucket must have a CORS
  policy that allows `GET` requests from the web client's origin(s) and
  exposes the response without requiring credentials:

  ```json
  [
    {
      "AllowedOrigins": ["https://your-inori-web-origin"],
      "AllowedMethods": ["GET"],
      "AllowedHeaders": ["Range"],
      "ExposeHeaders": ["Content-Length", "Content-Range", "Accept-Ranges"],
      "MaxAgeSeconds": 3600
    }
  ]
  ```

  Include every origin the web client is actually served from (e.g. both a
  production domain and a local dev `http://localhost:3000` if used against
  a shared bucket). Wildcard `AllowedOrigins: ["*"]` also satisfies the
  WebAudio requirement but is broader than necessary — prefer an explicit
  origin list for a production bucket holding user media.

## Graceful degradation

If the bucket CORS policy is missing or misconfigured, `createAudioGraph`
catches the resulting failure and falls back to direct playback: the
`<audio>` element still plays through its normal (non-WebAudio) path, so
**users never lose audio** — they simply don't get ReplayGain normalization
or crossfade until CORS is fixed. No action is required to keep playback
working; this doc exists so ReplayGain/crossfade can be diagnosed and fixed
when they silently don't apply.

## Verifying

1. Open DevTools → Console while playing a track with ReplayGain enabled
   (`/settings/audio`). A CORS failure surfaces as a `SecurityError` or
   "already connected" warning near `createMediaElementSource` the first
   time playback starts; the app itself doesn't log anything on the
   fallback path since it's intentionally silent for the user.
2. Confirm the bucket CORS policy is deployed for presigned-URL backends
   per the requirements above, then reload — the error should disappear and
   the ReplayGain toggle should produce an audible loudness difference.
