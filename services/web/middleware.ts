/**
 * Next.js middleware — client-side route guard.
 *
 * Redirects unauthenticated visitors away from protected routes.
 * The /login page and public assets are exempt.
 *
 * NOTE: middleware runs on the Edge runtime and cannot access localStorage.
 * We use a "inori_auth" cookie (written by the AuthProvider on login) as a
 * lightweight hint that a session is active. The actual token is validated
 * by the API on every request; the cookie is only used here for routing.
 */
import { NextRequest, NextResponse } from "next/server";

const PUBLIC_PATHS = ["/login", "/_next", "/favicon.ico", "/api/v1"];

export function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl;

  // Allow public paths through without checking auth.
  if (PUBLIC_PATHS.some((p) => pathname.startsWith(p))) {
    return NextResponse.next();
  }

  const hasSession = req.cookies.has("inori_session");
  if (!hasSession) {
    const loginUrl = req.nextUrl.clone();
    loginUrl.pathname = "/login";
    loginUrl.searchParams.set("from", pathname);
    return NextResponse.redirect(loginUrl);
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    /*
     * Match all request paths EXCEPT static files, images, and API routes.
     * The pattern below excludes _next/static, _next/image, and any file
     * with an extension (e.g. .ico, .png).
     */
    "/((?!_next/static|_next/image|.*\\..*|api/v1).*)",
  ],
};
