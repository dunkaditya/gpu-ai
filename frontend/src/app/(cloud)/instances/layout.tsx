import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Instances",
};

export default function InstancesLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return <>{children}</>;
}
