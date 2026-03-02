import { clerkMiddleware, createRouteMatcher } from "@clerk/nextjs/server";
import { NextResponse } from "next/server";

const isPublicRoute = createRouteMatcher([
  "/",
  "/sign-in(.*)",
  "/sign-up(.*)",
]);

export const proxy = clerkMiddleware(async (auth, req) => {
  const hostname = req.headers.get("host")?.split(":")[0] ?? "";
  const { searchParams, pathname } = req.nextUrl;

  const isCloud =
    hostname === "cloud.gpu.ai" || searchParams.get("site") === "cloud";

  // Protect cloud routes (require authentication)
  if (isCloud && !isPublicRoute(req)) {
    await auth.protect();
  }

  // Rewrite to appropriate route group
  const url = req.nextUrl.clone();
  url.pathname = isCloud ? `/(cloud)${pathname}` : `/(marketing)${pathname}`;
  return NextResponse.rewrite(url);
});

export const config = {
  matcher: [
    "/((?!api|_next/static|_next/image|favicon.ico|fonts|.*\\.(?:svg|png|jpg|jpeg|gif|webp|woff2?|ico)$).*)",
  ],
};
