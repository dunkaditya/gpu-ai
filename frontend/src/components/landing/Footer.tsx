import { Container } from "@/components/ui/Container";
import { ChipLogo } from "@/components/ui/ChipLogo";
import { Icon } from "@/components/ui/Icon";
import { FOOTER_LINKS } from "@/lib/constants";

export function Footer() {
  return (
    <footer className="border-t border-border bg-bg py-16 md:py-20">
      <Container>
        <div className="grid grid-cols-2 gap-10 lg:grid-cols-5">
          {/* Brand */}
          <div className="md:col-span-1">
            <a href="/" className="flex items-center gap-0.5">
              <ChipLogo size={28} />
              <span className="font-sans text-[18px] font-bold tracking-[-0.5px]">
                <span className="text-white">gpu</span>
                <span className="gradient-text">.ai</span>
              </span>
            </a>
            <p className="type-ui-sm mt-2 leading-[1.6] text-text-dim">
              The GPU cloud built to save you money.
            </p>
          </div>

          {/* Link columns */}
          {Object.entries(FOOTER_LINKS).map(([category, links]) => (
            <div key={category}>
              <h4 className="type-ui-2xs mb-4 font-semibold uppercase tracking-[0.08em] text-text-muted">
                {category}
              </h4>
              <ul className="flex flex-col gap-2.5">
                {links.map((link) => (
                  <li key={link.label}>
                    <a
                      href={link.href}
                      className="type-ui-sm text-text-dim transition-colors hover:text-text"
                    >
                      {link.label}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        {/* Bottom bar */}
        <div className="mt-14 flex flex-col items-center justify-between gap-4 border-t border-border pt-8 md:flex-row">
          <span className="type-ui-sm text-text-dim">
            &copy; {new Date().getFullYear()} GPU.ai. All rights reserved.
          </span>
          <div className="flex items-center gap-5">
            <a href="https://twitter.com" aria-label="Twitter">
              <Icon
                name="twitter"
                size={18}
                className="text-text-dim transition-colors hover:text-text-muted"
              />
            </a>
            <a href="https://github.com" aria-label="GitHub">
              <Icon
                name="github"
                size={18}
                className="text-text-dim transition-colors hover:text-text-muted"
              />
            </a>
            <a href="https://discord.gg" aria-label="Discord">
              <Icon
                name="discord"
                size={18}
                className="text-text-dim transition-colors hover:text-text-muted"
              />
            </a>
          </div>
        </div>
      </Container>
    </footer>
  );
}
