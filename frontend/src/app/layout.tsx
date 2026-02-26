import type { Metadata } from "next";
import { GeistSans } from "geist/font/sans";
import { GeistMono } from "geist/font/mono";
import "./globals.css";

export const metadata: Metadata = {
  title: "GPU.ai — Your Infrastructure for GPU Compute",
  description:
    "Aggregate GPU inventory from 12+ providers. Best price, fastest deploy, single API. Up to 30% cheaper than hyperscalers.",
  openGraph: {
    title: "GPU.ai — Your Infrastructure for GPU Compute",
    description:
      "Aggregate GPU inventory from 12+ providers. Best price, fastest deploy, single API.",
    type: "website",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body
        className={`${GeistSans.variable} ${GeistMono.variable} antialiased`}
      >
        {children}
      </body>
    </html>
  );
}
