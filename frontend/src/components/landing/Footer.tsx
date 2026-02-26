import { FOOTER_COLUMNS } from "@/lib/constants";
import { Container } from "@/components/ui";

export function Footer() {
  return (
    <footer className="border-t border-white/[0.08] py-16">
      <Container>
        {/* Top section: columns */}
        <div className="grid grid-cols-2 md:grid-cols-5 gap-8">
          {/* Branding column */}
          <div>
            <span className="text-lg font-semibold text-white">GPU.ai</span>
            <p className="mt-2 text-sm text-text-dim">GPU compute, simplified.</p>
          </div>

          {/* Link columns */}
          {FOOTER_COLUMNS.map((column) => (
            <div key={column.title}>
              <span className="text-sm font-medium text-white">{column.title}</span>
              <div className="mt-4 space-y-3">
                {column.links.map((link) => (
                  <a
                    key={link.label}
                    href={link.href}
                    className="block text-sm text-text-muted hover:text-white transition-colors"
                  >
                    {link.label}
                  </a>
                ))}
              </div>
            </div>
          ))}
        </div>

        {/* Bottom section */}
        <div className="mt-12 flex items-center justify-between border-t border-white/[0.08] pt-8">
          <span className="text-sm text-text-dim">
            &copy; 2026 GPU.ai. All rights reserved.
          </span>
        </div>
      </Container>
    </footer>
  );
}
