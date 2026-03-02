import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "GPU.ai -- GPU Cloud at Unbeatable Prices",
  description:
    "We aggregate GPU inventory from 12+ providers to find you the best price. Up to 30% cheaper than hyperscalers. Deploy in seconds.",
  openGraph: {
    title: "GPU.ai -- GPU Cloud at Unbeatable Prices",
    description:
      "Aggregate GPU cloud. Up to 30% cheaper. Deploy in seconds.",
    type: "website",
  },
};

export default function MarketingLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return <>{children}</>;
}
