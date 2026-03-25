"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { UserButton } from "@clerk/nextjs";
import useSWR from "swr";
import { cn } from "@/lib/utils";
import { gracefulFetcher } from "@/lib/api";
import type { BalanceResponse } from "@/lib/types";

const hasClerk = !!process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY;

/* ── Segment labels for breadcrumb ── */
const segmentLabels: Record<string, string> = {
  instances: "Instances",
  "gpu-availability": "GPU Availability",
  "ssh-keys": "SSH Keys",
  billing: "Billing",
  settings: "Settings",
  "api-keys": "API Keys",
  team: "Team",
};

/* ── UUID pattern to detect dynamic route segments ── */
const UUID_REGEX = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

function buildBreadcrumb(pathname: string): { label: string; isLast: boolean }[] {
  // Remove /cloud prefix and split
  const segments = pathname.replace(/^\/cloud\/?/, "").split("/").filter(Boolean);

  if (segments.length === 0) {
    return [{ label: "Dashboard", isLast: true }];
  }

  return segments.map((seg, idx) => {
    const isLast = idx === segments.length - 1;

    // Known segment label
    if (segmentLabels[seg]) {
      return { label: segmentLabels[seg], isLast };
    }

    // UUID-like segment (instance detail)
    if (UUID_REGEX.test(seg)) {
      return { label: "Instance Detail", isLast };
    }

    // Fallback: capitalize
    return { label: seg.charAt(0).toUpperCase() + seg.slice(1), isLast };
  });
}

export function DashboardTopbar({ onMenuToggle }: { onMenuToggle?: () => void }) {
  const pathname = usePathname();
  const breadcrumb = buildBreadcrumb(pathname);

  return (
    <header className="flex items-center justify-between h-14 px-6 bg-bg border-b border-border shrink-0">
      <div className="flex items-center gap-3 min-w-0">
        {/* Hamburger menu - mobile only */}
        <button
          onClick={onMenuToggle}
          className="md:hidden flex items-center justify-center w-8 h-8 rounded-md text-text-muted hover:text-text hover:bg-bg-card transition-colors shrink-0"
          aria-label="Toggle navigation menu"
        >
          <svg width="18" height="18" viewBox="0 0 18 18" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M2 4.5h14M2 9h14M2 13.5h14" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
          </svg>
        </button>

        {/* Breadcrumb */}
        <div className="flex items-center gap-2 type-ui-sm min-w-0 overflow-hidden">
          <span className="text-text-dim whitespace-nowrap">Cloud</span>
          {breadcrumb.map((crumb, idx) => (
            <span key={idx} className="flex items-center gap-2 min-w-0">
              <span className="text-text-dim/40 shrink-0">&gt;</span>
              <span className={cn(crumb.isLast ? "text-text truncate" : "text-text-dim", "whitespace-nowrap")}>
                {crumb.label}
              </span>
            </span>
          ))}
        </div>
      </div>

      {/* Balance + User section */}
      <div className="flex items-center gap-3">
        <BalancePill />
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
            <div className="w-8 h-8 rounded-full bg-bg-card-hover border border-border flex items-center justify-center">
              <span className="text-xs font-medium text-text-dim">D</span>
            </div>
            <span className="text-xs text-text-dim font-mono">Dev User</span>
          </div>
        )}
      </div>
    </header>
  );
}

function BalancePill() {
  const { data } = useSWR<BalanceResponse>(
    "/api/v1/billing/balance",
    gracefulFetcher,
    { refreshInterval: 15000 }
  );

  if (!data) return null;

  const dollars = (data.balance_cents / 100).toFixed(2);

  return (
    <Link
      href="/cloud/billing"
      className="flex items-center gap-0 rounded-md border border-border bg-bg-card hover:bg-bg-card-hover transition-colors overflow-hidden"
    >
      <span className="px-2.5 py-1 type-ui-xs font-mono text-text">
        ${dollars}
      </span>
      <span className="px-1.5 py-1 bg-green text-bg font-bold type-ui-xs">
        +
      </span>
    </Link>
  );
}
