import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "GPU Availability",
};

export default function GPUAvailabilityLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return <>{children}</>;
}
