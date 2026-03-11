import type { NextConfig } from "next";

const apiUrl = process.env.BACKEND_URL || "http://localhost:9090";

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      {
        source: "/api/v1/:path*",
        destination: `${apiUrl}/api/v1/:path*`,
      },
    ];
  },
};

export default nextConfig;
