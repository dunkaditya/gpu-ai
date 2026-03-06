import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "SSH Keys",
};

export default function SSHKeysLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return <>{children}</>;
}
