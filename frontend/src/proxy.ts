import { NextResponse, type NextRequest } from "next/server";

export function proxy(request: NextRequest) {
  const hostname = request.headers.get("host")?.split(":")[0] ?? "";
  const { searchParams, pathname } = request.nextUrl;

  const isCloud =
    hostname === "cloud.gpu.ai" || searchParams.get("site") === "cloud";

  const url = request.nextUrl.clone();
  url.pathname = isCloud ? `/(cloud)${pathname}` : `/(marketing)${pathname}`;
  return NextResponse.rewrite(url);
}

export const config = {
  matcher: [
    "/((?!api|_next/static|_next/image|favicon.ico|fonts|.*\\.(?:svg|png|jpg|jpeg|gif|webp|woff2?|ico)$).*)",
  ],
};
