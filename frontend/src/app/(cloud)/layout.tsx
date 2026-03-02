import type { Metadata } from "next";
import { DashboardSidebar } from "@/components/cloud/DashboardSidebar";
import { DashboardTopbar } from "@/components/cloud/DashboardTopbar";

export const metadata: Metadata = {
  title: "Dashboard",
};

export default function CloudLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <div className="flex h-screen">
      <DashboardSidebar />
      <div className="flex flex-1 flex-col overflow-hidden">
        <DashboardTopbar />
        <main className="flex-1 overflow-y-auto p-6">{children}</main>
      </div>
    </div>
  );
}
