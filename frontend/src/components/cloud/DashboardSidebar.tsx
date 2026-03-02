"use client";

import Link from "next/link";
import { usePathname, useSearchParams } from "next/navigation";
import { cn } from "@/lib/utils";

const navItems = [
  {
    label: "Instances",
    href: "/instances",
    icon: (
      <svg
        width="16"
        height="16"
        viewBox="0 0 16 16"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <rect
          x="1"
          y="2"
          width="14"
          height="4"
          rx="1"
          stroke="currentColor"
          strokeWidth="1.5"
        />
        <rect
          x="1"
          y="10"
          width="14"
          height="4"
          rx="1"
          stroke="currentColor"
          strokeWidth="1.5"
        />
        <circle cx="4" cy="4" r="1" fill="currentColor" />
        <circle cx="4" cy="12" r="1" fill="currentColor" />
      </svg>
    ),
  },
  {
    label: "GPU Availability",
    href: "/gpu-availability",
    icon: (
      <svg
        width="16"
        height="16"
        viewBox="0 0 16 16"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <rect
          x="1"
          y="8"
          width="3"
          height="6"
          rx="0.5"
          stroke="currentColor"
          strokeWidth="1.5"
        />
        <rect
          x="6.5"
          y="4"
          width="3"
          height="10"
          rx="0.5"
          stroke="currentColor"
          strokeWidth="1.5"
        />
        <rect
          x="12"
          y="2"
          width="3"
          height="12"
          rx="0.5"
          stroke="currentColor"
          strokeWidth="1.5"
        />
      </svg>
    ),
  },
  {
    label: "SSH Keys",
    href: "/ssh-keys",
    icon: (
      <svg
        width="16"
        height="16"
        viewBox="0 0 16 16"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <circle cx="6" cy="7" r="3" stroke="currentColor" strokeWidth="1.5" />
        <path
          d="M8.5 9.5L14 15"
          stroke="currentColor"
          strokeWidth="1.5"
          strokeLinecap="round"
        />
        <path
          d="M12 13L14 11"
          stroke="currentColor"
          strokeWidth="1.5"
          strokeLinecap="round"
        />
      </svg>
    ),
  },
  {
    label: "Billing",
    href: "/billing",
    icon: (
      <svg
        width="16"
        height="16"
        viewBox="0 0 16 16"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <rect
          x="1"
          y="3"
          width="14"
          height="10"
          rx="1.5"
          stroke="currentColor"
          strokeWidth="1.5"
        />
        <path
          d="M1 6.5H15"
          stroke="currentColor"
          strokeWidth="1.5"
        />
        <path
          d="M4 10H7"
          stroke="currentColor"
          strokeWidth="1.5"
          strokeLinecap="round"
        />
      </svg>
    ),
  },
  {
    label: "Settings",
    href: "/settings",
    icon: (
      <svg
        width="16"
        height="16"
        viewBox="0 0 16 16"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <circle cx="8" cy="8" r="2" stroke="currentColor" strokeWidth="1.5" />
        <path
          d="M8 1V3M8 13V15M1 8H3M13 8H15M2.93 2.93L4.34 4.34M11.66 11.66L13.07 13.07M13.07 2.93L11.66 4.34M4.34 11.66L2.93 13.07"
          stroke="currentColor"
          strokeWidth="1.5"
          strokeLinecap="round"
        />
      </svg>
    ),
  },
];

export function DashboardSidebar() {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const siteParam = searchParams.get("site");

  function buildHref(path: string) {
    if (siteParam === "cloud") {
      return `${path}?site=cloud`;
    }
    return path;
  }

  return (
    <aside className="hidden md:flex w-60 flex-col bg-bg-alt border-r border-border h-screen shrink-0">
      {/* Logo */}
      <div className="flex items-center h-14 px-5 border-b border-border">
        <Link
          href={siteParam === "cloud" ? "/?site=marketing" : "/"}
          className="font-sans text-lg font-bold tracking-tight text-text hover:text-purple transition-colors"
        >
          GPU.ai
        </Link>
      </div>

      {/* Navigation */}
      <nav className="flex-1 px-3 py-4 space-y-0.5">
        {navItems.map((item) => {
          const isActive = pathname === item.href || pathname.startsWith(item.href + "/");

          return (
            <Link
              key={item.href}
              href={buildHref(item.href)}
              className={cn(
                "flex items-center gap-3 px-3 py-2 rounded-md type-ui-sm transition-colors",
                isActive
                  ? "bg-bg-card-hover text-text"
                  : "text-text-muted hover:text-text hover:bg-bg-card"
              )}
            >
              <span className="shrink-0 opacity-70">{item.icon}</span>
              {item.label}
            </Link>
          );
        })}
      </nav>

      {/* Footer */}
      <div className="px-5 py-4 border-t border-border">
        <p className="type-ui-2xs text-text-dim">v1.0.0</p>
      </div>
    </aside>
  );
}
