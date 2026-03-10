import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Instance Detail",
};

export default function InstanceDetailLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return <>{children}</>;
}
