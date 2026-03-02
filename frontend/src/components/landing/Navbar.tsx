"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/Button";
import { ChipLogo } from "@/components/ui/ChipLogo";
import { NAV_LINKS } from "@/lib/constants";

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
        <a href="/" className="flex items-center gap-2.5">
          <ChipLogo size={38} />
          <span className="font-sans text-[22px] font-bold tracking-[-0.5px]">
            <span className="text-white">gpu</span>
            <span className="gradient-text">.ai</span>
          </span>
        </a>

        {/* Desktop Nav — centered, uppercase, tracked */}
        <div className="hidden items-center gap-20 lg:flex">
          {NAV_LINKS.map((link) => (
            <a
              key={link.label}
              href={link.href}
              className="type-ui font-medium uppercase tracking-[0.08em] text-text-muted transition-colors hover:text-white"
            >
              {link.label}
            </a>
          ))}
        </div>

        {/* Desktop CTA */}
        <div className="hidden items-center gap-6 lg:flex">
          <a
            href="/sign-in"
            className="type-ui font-medium uppercase tracking-[0.08em] text-text-muted transition-colors hover:text-white"
          >
            Log in
          </a>
          <Button size="sm" href="/sign-up" className="gradient-btn">
            GET STARTED
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
            <hr className="border-border" />
            <a href="/sign-in" className="type-ui-sm uppercase tracking-[0.06em] text-text-muted">
              Log in
            </a>
            <Button size="sm" href="/sign-up" className="gradient-btn">
              GET STARTED
            </Button>
          </div>
        </div>
      )}
    </nav>
  );
}
