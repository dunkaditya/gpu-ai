"use client";

import { usePathname } from "next/navigation";
import { UserButton } from "@clerk/nextjs";

const pathLabels: Record<string, string> = {
  "/cloud/instances": "Instances",
  "/cloud/gpu-availability": "GPU Availability",
  "/cloud/ssh-keys": "SSH Keys",
  "/cloud/billing": "Billing",
  "/cloud/settings": "Settings",
};

const hasClerk = !!process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY;

export function DashboardTopbar() {
  const pathname = usePathname();
  const pageLabel = pathLabels[pathname] ?? "Dashboard";

  return (
    <header className="flex items-center justify-between h-14 px-6 bg-bg border-b border-border shrink-0">
      {/* Breadcrumb */}
      <div className="flex items-center gap-2 type-ui-sm">
        <span className="text-text-dim">Cloud</span>
        <span className="text-text-dim">/</span>
        <span className="text-text font-medium">{pageLabel}</span>
      </div>

      {/* User section */}
      <div className="flex items-center gap-4">
        {hasClerk ? (
          <UserButton
            appearance={{
              elements: {
                avatarBox: "w-8 h-8",
              },
            }}
          />
        ) : (
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-full bg-accent/20 border border-accent/40 flex items-center justify-center">
              <span className="text-xs font-medium text-accent">D</span>
            </div>
            <span className="text-xs text-text-dim font-mono">Dev User</span>
          </div>
        )}
      </div>
    </header>
  );
}
