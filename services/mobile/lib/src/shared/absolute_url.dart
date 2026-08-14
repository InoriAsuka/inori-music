/// Makes a server-relative API URL absolute against the configured server.
///
/// Several endpoints hand back a **relative**, HMAC-signed path instead of a
/// fully-qualified URL — e.g. track streaming's
/// `/api/v1/catalog/tracks/<id>/stream?exp=…&sig=…`, or (as of v5.39.0) album
/// artwork's `/api/v1/catalog/albums/<id>/artwork/file?exp=…&sig=…`. That is
/// correct for the web client, which is same-origin and lets the browser
/// fill in the rest; it is unusable anywhere a URL has to be handed to
/// something that resolves it itself — a native player, or an image loader.
/// `Uri.parse('/api/v1/…')` yields a URI with no scheme and no host: handing
/// that to AVPlayer produced `NSURLErrorUnsupportedURL` (-1002, v5.37.1), and
/// an `Image.network`/`CachedNetworkImage` widget fails the same way, just
/// silently — no exception, just a request that can never leave the device.
///
/// [Uri.resolve] covers both shapes in one call — a presigned object-storage
/// URL is already absolute, possibly on an entirely different host than the
/// API, and passes through untouched; a relative path gains the base's
/// scheme and host.
///
/// The query string (`exp`/`sig`) is the only credential these signed
/// requests carry — neither the audio engine nor an image loader attaches an
/// Authorization header — so it must survive unchanged. [Uri.resolve]
/// preserves it because it operates on the whole reference, including the
/// query, not just the path component.
String absoluteApiUrl(String url, String base) =>
    Uri.parse(base).resolve(url).toString();
