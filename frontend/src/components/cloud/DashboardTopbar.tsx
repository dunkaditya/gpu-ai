"use client";

import { usePathname } from "next/navigation";
import { UserButton } from "@clerk/nextjs";

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
      <div className="flex items-center gap-3">
        {/* Hamburger menu - mobile only */}
        <button
          onClick={onMenuToggle}
          className="md:hidden flex items-center justify-center w-8 h-8 rounded-md text-text-muted hover:text-text hover:bg-bg-card transition-colors"
          aria-label="Toggle navigation menu"
        >
          <svg width="18" height="18" viewBox="0 0 18 18" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M2 4.5h14M2 9h14M2 13.5h14" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
          </svg>
        </button>

        {/* Breadcrumb */}
        <div className="flex items-center gap-2 type-ui-sm">
          <span className="text-text-dim">Cloud</span>
          {breadcrumb.map((crumb, idx) => (
            <span key={idx} className="flex items-center gap-2">
              <span className="text-text-dim/40">&gt;</span>
              <span className={crumb.isLast ? "text-text" : "text-text-dim"}>
                {crumb.label}
              </span>
            </span>
          ))}
        </div>
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
