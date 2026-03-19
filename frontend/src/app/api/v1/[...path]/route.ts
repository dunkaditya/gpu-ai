import { auth } from "@clerk/nextjs/server";
import { NextRequest, NextResponse } from "next/server";

const BACKEND_URL = process.env.BACKEND_URL || "http://localhost:9090";

async function proxyRequest(req: NextRequest) {
  const { getToken } = await auth();
  const token = await getToken();

  const url = new URL(req.nextUrl.pathname + req.nextUrl.search, BACKEND_URL);

  const headers = new Headers(req.headers);
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }
  headers.delete("host");

  const res = await fetch(url, {
    method: req.method,
    headers,
    body: req.method !== "GET" && req.method !== "HEAD" ? await req.text() : undefined,
  });

  return new NextResponse(res.body, {
    status: res.status,
    headers: res.headers,
  });
}

export const GET = proxyRequest;
export const POST = proxyRequest;
export const PUT = proxyRequest;
export const PATCH = proxyRequest;
export const DELETE = proxyRequest;
