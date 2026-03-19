"use client";

import { useEffect, useRef, useState } from "react";
import { Button } from "@/components/ui/Button";
import { ChipLogo } from "@/components/ui/ChipLogo";
import { NAV_LINKS, COMPANY_LINKS } from "@/lib/constants";

function CompanyDropdown() {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, []);

  return (
    <div ref={ref} className="relative">
      <button
        onClick={() => setOpen(!open)}
        className="flex items-center gap-1.5 type-ui font-medium uppercase tracking-[0.08em] text-text-muted transition-colors hover:text-white"
      >
        Company
        <svg
          className={`h-3 w-3 transition-transform duration-200 ${open ? "rotate-180" : ""}`}
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={2.5}
        >
          <path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {/* Dropdown panel */}
      <div
        className={`absolute right-1/2 translate-x-1/2 top-full pt-3 transition-all duration-200 ${
          open
            ? "opacity-100 translate-y-0 pointer-events-auto"
            : "opacity-0 -translate-y-1 pointer-events-none"
        }`}
      >
        <div
          className="rounded-xl border border-border-light/40 p-1"
          style={{
            background:
              "linear-gradient(165deg, rgba(20, 18, 36, 0.92) 0%, rgba(10, 10, 18, 0.96) 100%)",
            backdropFilter: "blur(24px) saturate(1.4)",
            WebkitBackdropFilter: "blur(24px) saturate(1.4)",
            boxShadow:
              "0 20px 50px -10px rgba(0,0,0,0.6), inset 0 1px 0 rgba(255,255,255,0.06)",
          }}
        >
          <div className="flex min-w-[380px]">
            {COMPANY_LINKS.map((link) => (
              <a
                key={link.label}
                href={link.href}
                onClick={() => setOpen(false)}
                className="group flex-1 rounded-lg px-5 py-4 transition-colors hover:bg-white/[0.04]"
              >
                <span className="flex items-center gap-1.5 text-[14px] font-semibold text-white">
                  {link.label}
                  <svg
                    className="h-3 w-3 text-text-dim opacity-0 transition-all duration-200 group-hover:opacity-100 group-hover:translate-x-0.5"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    strokeWidth={2}
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      d="M7 17L17 7M17 7H7M17 7v10"
                    />
                  </svg>
                </span>
              </a>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

export function Navbar() {
  const [scrolled, setScrolled] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 40);
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  return (
    <nav
      className={`fixed top-0 z-50 w-full border-b border-border transition-all duration-300 ${
        scrolled ? "glass" : "bg-bg/80 backdrop-blur-sm"
      }`}
    >
      <div className="mx-auto flex h-[88px] max-w-[1250px] items-center justify-between px-6">
        {/* Logo */}
        <a href="/" className="flex items-center gap-0.5">
          <ChipLogo size={38} />
          <span className="font-sans text-[22px] font-bold tracking-[-0.5px]">
            <span className="text-white">gpu</span>
            <span className="gradient-text">.ai</span>
          </span>
        </a>

        {/* Desktop Nav — centered, uppercase, tracked */}
        <div className="hidden items-center gap-16 lg:flex">
          {NAV_LINKS.map((link) => (
            <a
              key={link.label}
              href={link.href}
              className="type-ui font-medium uppercase tracking-[0.08em] text-text-muted transition-colors hover:text-white"
            >
              {link.label}
            </a>
          ))}
          <CompanyDropdown />
        </div>

        {/* Desktop CTA */}
        <div className="hidden items-center gap-6 lg:flex">
          <a
            href="/sign-in"
            className="type-ui font-medium uppercase tracking-[0.08em] text-text-muted transition-colors hover:text-white"
          >
            Log in
          </a>
          <Button size="sm" href="/free-trial" className="gradient-btn">
            $100 FREE TRIAL
          </Button>
        </div>

        {/* Mobile Hamburger */}
        <button
          className="flex flex-col gap-[5px] lg:hidden"
          onClick={() => setMobileOpen(!mobileOpen)}
          aria-label="Toggle menu"
        >
          <span
            className={`h-[2px] w-5 bg-white transition-all duration-200 ${mobileOpen ? "translate-y-[7px] rotate-45" : ""}`}
          />
          <span
            className={`h-[2px] w-5 bg-white transition-all duration-200 ${mobileOpen ? "opacity-0" : ""}`}
          />
          <span
            className={`h-[2px] w-5 bg-white transition-all duration-200 ${mobileOpen ? "-translate-y-[7px] -rotate-45" : ""}`}
          />
        </button>
      </div>

      {/* Mobile Menu */}
      {mobileOpen && (
        <div className="border-t border-border bg-bg/95 backdrop-blur-md lg:hidden">
          <div className="mx-auto flex max-w-[1200px] flex-col gap-4 px-6 py-6">
            {NAV_LINKS.map((link) => (
              <a
                key={link.label}
                href={link.href}
                className="type-ui-sm uppercase tracking-[0.06em] text-text-muted transition-colors hover:text-white"
                onClick={() => setMobileOpen(false)}
              >
                {link.label}
              </a>
            ))}
            <span className="type-ui-sm uppercase tracking-[0.06em] text-text-dim">
              Company
            </span>
            {COMPANY_LINKS.map((link) => (
              <a
                key={link.label}
                href={link.href}
                className="type-ui-sm pl-3 text-text-muted transition-colors hover:text-white"
                onClick={() => setMobileOpen(false)}
              >
                {link.label}
              </a>
            ))}
            <hr className="border-border" />
            <a href="/sign-in" className="type-ui-sm uppercase tracking-[0.06em] text-text-muted">
              Log in
            </a>
            <Button size="sm" href="/free-trial" className="gradient-btn">
              $100 FREE TRIAL
            </Button>
          </div>
        </div>
      )}
    </nav>
  );
}
