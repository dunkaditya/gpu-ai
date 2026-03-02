import type { Metadata } from "next";
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
  title: "GPU.ai — GPU Cloud at Unbeatable Prices",
  description:
    "We aggregate GPU inventory from 12+ providers to find you the best price. Up to 30% cheaper than hyperscalers. Deploy in seconds.",
  openGraph: {
    title: "GPU.ai — GPU Cloud at Unbeatable Prices",
    description:
      "Aggregate GPU cloud. Up to 30% cheaper. Deploy in seconds.",
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
        className={`${vremenaGrotesk.variable} ${nectoMono.variable} antialiased`}
      >
        {children}
      </body>
    </html>
  );
}
