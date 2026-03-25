"use client";

import { useState } from "react";
import { DashboardSidebar } from "@/components/cloud/DashboardSidebar";
import { DashboardTopbar } from "@/components/cloud/DashboardTopbar";

export default function CloudLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const [isMobileOpen, setIsMobileOpen] = useState(false);

  return (
    <div className="flex h-screen">
      <DashboardSidebar
        isMobileOpen={isMobileOpen}
        onMobileClose={() => setIsMobileOpen(false)}
      />
      <div className="flex flex-1 flex-col overflow-hidden">
        <DashboardTopbar onMenuToggle={() => setIsMobileOpen((prev) => !prev)} />
        <main className="flex-1 overflow-y-auto p-4 md:p-6">{children}</main>
      </div>
    </div>
  );
}
