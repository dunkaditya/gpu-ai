"use client";

import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";

const pathLabels: Record<string, string> = {
  "/instances": "Instances",
  "/gpu-availability": "GPU Availability",
  "/ssh-keys": "SSH Keys",
  "/billing": "Billing",
  "/settings": "Settings",
};

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
        <button
          className={cn(
            "type-ui-xs text-text-muted hover:text-text transition-colors"
          )}
        >
          Sign Out
        </button>
        <div
          className="flex items-center justify-center w-8 h-8 rounded-full bg-purple-dim text-purple type-ui-xs font-bold"
          aria-label="User avatar"
        >
          AR
        </div>
      </div>
    </header>
  );
}
