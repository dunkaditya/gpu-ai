"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import { ChipLogo } from "@/components/ui/ChipLogo";

/* ── Icon Components ── */

const InstancesIcon = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
    <rect x="1" y="2" width="14" height="4" rx="1" stroke="currentColor" strokeWidth="1.5" />
    <rect x="1" y="10" width="14" height="4" rx="1" stroke="currentColor" strokeWidth="1.5" />
    <circle cx="4" cy="4" r="1" fill="currentColor" />
    <circle cx="4" cy="12" r="1" fill="currentColor" />
  </svg>
);

const GPUIcon = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
    <rect x="1" y="8" width="3" height="6" rx="0.5" stroke="currentColor" strokeWidth="1.5" />
    <rect x="6.5" y="4" width="3" height="10" rx="0.5" stroke="currentColor" strokeWidth="1.5" />
    <rect x="12" y="2" width="3" height="12" rx="0.5" stroke="currentColor" strokeWidth="1.5" />
  </svg>
);

const SSHKeysIcon = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
    <circle cx="6" cy="7" r="3" stroke="currentColor" strokeWidth="1.5" />
    <path d="M8.5 9.5L14 15" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
    <path d="M12 13L14 11" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
  </svg>
);

const BillingIcon = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
    <rect x="1" y="3" width="14" height="10" rx="1.5" stroke="currentColor" strokeWidth="1.5" />
    <path d="M1 6.5H15" stroke="currentColor" strokeWidth="1.5" />
    <path d="M4 10H7" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
  </svg>
);

const APIKeysIcon = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M2 4h12M2 8h8M2 12h10" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
    <circle cx="13" cy="8" r="2" stroke="currentColor" strokeWidth="1.5" />
  </svg>
);

const TeamIcon = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
    <circle cx="6" cy="5" r="2.5" stroke="currentColor" strokeWidth="1.5" />
    <path d="M1.5 14c0-2.5 2-4.5 4.5-4.5s4.5 2 4.5 4.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
    <circle cx="11.5" cy="5.5" r="1.5" stroke="currentColor" strokeWidth="1.2" />
    <path d="M11.5 9c1.5 0 3 1 3 3" stroke="currentColor" strokeWidth="1.2" strokeLinecap="round" />
  </svg>
);

const SettingsIcon = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
    <circle cx="8" cy="8" r="2" stroke="currentColor" strokeWidth="1.5" />
    <path
      d="M8 1V3M8 13V15M1 8H3M13 8H15M2.93 2.93L4.34 4.34M11.66 11.66L13.07 13.07M13.07 2.93L11.66 4.34M4.34 11.66L2.93 13.07"
      stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"
    />
  </svg>
);

/* ── Navigation Structure ── */

interface NavItem {
  label: string;
  href: string;
  icon: React.ReactNode;
  comingSoon?: boolean;
}

const primaryNav: NavItem[] = [
  { label: "Instances", href: "/cloud/instances", icon: <InstancesIcon /> },
  { label: "GPU Availability", href: "/cloud/gpu-availability", icon: <GPUIcon /> },
];

const managementNav: NavItem[] = [
  { label: "SSH Keys", href: "/cloud/ssh-keys", icon: <SSHKeysIcon /> },
  { label: "Billing", href: "/cloud/billing", icon: <BillingIcon /> },
  { label: "API Keys", href: "/cloud/api-keys", icon: <APIKeysIcon />, comingSoon: true },
  { label: "Team", href: "/cloud/team", icon: <TeamIcon />, comingSoon: true },
];

const bottomNav: NavItem[] = [
  { label: "Settings", href: "/cloud/settings", icon: <SettingsIcon /> },
];

/* ── NavLink Component ── */

function NavLink({ item, pathname, onClick }: { item: NavItem; pathname: string; onClick?: () => void }) {
  const isActive = pathname === item.href || pathname.startsWith(item.href + "/");

  return (
    <Link
      href={item.href}
      onClick={onClick}
      className={cn(
        "flex items-center gap-3 px-3 py-1.5 type-ui-sm transition-colors",
        isActive
          ? "border-l-2 border-purple text-text"
          : "border-l-2 border-transparent text-text-muted hover:text-text"
      )}
    >
      <span className="shrink-0 opacity-70">{item.icon}</span>
      <span className="flex-1">{item.label}</span>
      {item.comingSoon && (
        <span className="type-ui-2xs bg-bg-card-hover text-text-dim rounded-full px-1.5 py-0.5">
          Soon
        </span>
      )}
    </Link>
  );
}

/* ── Sidebar Content (shared between desktop and mobile) ── */

function SidebarContent({ pathname, onNavClick }: { pathname: string; onNavClick?: () => void }) {
  return (
    <>
      {/* Logo */}
      <div className="flex items-center h-14 px-5 border-b border-border">
        <Link href="/" className="flex items-center gap-0.5">
          <ChipLogo size={28} />
          <span className="font-sans text-[18px] font-bold tracking-[-0.5px]">
            <span className="text-white">gpu</span>
            <span className="gradient-text">.ai</span>
          </span>
        </Link>
      </div>

      {/* Navigation */}
      <nav className="flex-1 flex flex-col px-3 py-4">
        {/* Primary: Instances, GPU Availability */}
        <div className="space-y-0.5">
          {primaryNav.map((item) => (
            <NavLink key={item.href} item={item} pathname={pathname} onClick={onNavClick} />
          ))}
        </div>

        {/* Spacer between nav groups */}
        <div className="my-4" />

        {/* Management: SSH Keys, Billing, API Keys, Team */}
        <div className="space-y-0.5">
          {managementNav.map((item) => (
            <NavLink key={item.href} item={item} pathname={pathname} onClick={onNavClick} />
          ))}
        </div>

        {/* Spacer */}
        <div className="flex-1" />

        {/* Bottom: Settings */}
        <div className="space-y-0.5">
          {bottomNav.map((item) => (
            <NavLink key={item.href} item={item} pathname={pathname} onClick={onNavClick} />
          ))}
        </div>
      </nav>

      {/* Footer */}
      <div className="px-5 py-4 border-t border-border">
        <p className="type-ui-2xs text-text-dim/50">v1.0.0</p>
      </div>
    </>
  );
}

/* ── Desktop Sidebar ── */

export function DashboardSidebar({ isMobileOpen, onMobileClose }: {
  isMobileOpen?: boolean;
  onMobileClose?: () => void;
}) {
  const pathname = usePathname();

  return (
    <>
      {/* Desktop sidebar */}
      <aside className="hidden md:flex w-60 flex-col bg-bg-alt border-r border-border h-screen shrink-0">
        <SidebarContent pathname={pathname} />
      </aside>

      {/* Mobile sidebar overlay */}
      {isMobileOpen && (
        <div className="fixed inset-0 z-50 md:hidden">
          {/* Backdrop */}
          <div
            className="absolute inset-0 bg-bg/80 backdrop-blur-sm"
            onClick={onMobileClose}
          />

          {/* Slide-in sidebar */}
          <aside className="relative w-60 h-full flex flex-col bg-bg-alt border-r border-border shadow-2xl animate-fade-up"
            style={{ animationDuration: "0.2s" }}
          >
            <SidebarContent pathname={pathname} onNavClick={onMobileClose} />
          </aside>
        </div>
      )}
    </>
  );
}
