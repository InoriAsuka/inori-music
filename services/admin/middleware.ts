import { NextRequest, NextResponse } from "next/server";

const PUBLIC_PATHS = ["/login", "/_next", "/favicon.ico", "/api/v1"];

export function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl;
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
  matcher: ["/((?!_next/static|_next/image|.*\\..*|api/v1).*)"],
};
