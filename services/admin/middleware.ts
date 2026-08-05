import { type NextRequest, NextResponse } from "next/server";

const PUBLIC_PATHS = ["/login", "/_next", "/favicon.ico", "/api/v1"];

/**
 * With `basePath: "/admin"` set in next.config.ts, Next.js already strips
 * the basePath from `req.nextUrl.pathname` (and auto-prepends it back to
 * both `config.matcher` patterns and any pathname assigned on a cloned
 * NextURL) — this handler must not add or remove "/admin" itself, or the
 * matcher stops matching real requests and redirects double up the prefix.
 */
export function middleware(req: NextRequest) {
  const pathname = req.nextUrl.pathname;

  if (PUBLIC_PATHS.some((p) => pathname.startsWith(p))) return NextResponse.next();

  const hasSession = req.cookies.has("inori_admin_session");
  if (!hasSession) {
    const loginUrl = req.nextUrl.clone();
    loginUrl.pathname = "/login";
    loginUrl.searchParams.set("from", pathname);
    return NextResponse.redirect(loginUrl);
  }
  return NextResponse.next();
}

export const config = {
  matcher: ["/:path*"],
};
