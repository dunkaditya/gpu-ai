import type { Metadata } from "next";
import { ClerkProvider } from "@clerk/nextjs";
import localFont from "next/font/local";
import "./globals.css";

const vremenaGrotesk = localFont({
  src: "../../public/fonts/vremena-grotesk.woff2",
  variable: "--font-vremena-grotesk",
  display: "swap",
});

const nectoMono = localFont({
  src: "../../public/fonts/necto-mono.woff2",
  variable: "--font-necto-mono",
  display: "swap",
});

export const metadata: Metadata = {
  title: { template: "%s | GPU.ai", default: "GPU.ai" },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <ClerkProvider>
      <html lang="en" className="dark">
        <body
          className={`${vremenaGrotesk.variable} ${nectoMono.variable} antialiased`}
        >
          {children}
        </body>
      </html>
    </ClerkProvider>
  );
}
