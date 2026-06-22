import { NextRequest, NextResponse } from "next/server";

const PUBLIC_PATHS = ["/login", "/_next", "/favicon.ico", "/api/v1"];

function stripBasePath(pathname: string): string {
  if (pathname === "/admin") return "/";
  if (pathname.startsWith("/admin/")) return pathname.slice("/admin".length);
  return pathname;
}

export function middleware(req: NextRequest) {
  const rawPathname = req.nextUrl.pathname;
  const pathname = stripBasePath(rawPathname);

  if (PUBLIC_PATHS.some((p) => pathname.startsWith(p))) return NextResponse.next();
  if (rawPathname.startsWith("/admin/_next")) return NextResponse.next();

  const hasSession = req.cookies.has("inori_admin_session");
  if (!hasSession) {
    const loginUrl = req.nextUrl.clone();
    loginUrl.pathname = "/admin/login";
    loginUrl.searchParams.set("from", rawPathname);
    return NextResponse.redirect(loginUrl);
  }
  return NextResponse.next();
}

export const config = {
  matcher: ["/admin/:path*"],
};
