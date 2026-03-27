import { Navbar } from "@/components/landing/Navbar";
import { Footer } from "@/components/landing/Footer";
import { Container } from "@/components/ui/Container";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Changelog",
};

const ENTRIES = [
  {
    date: "March 2026",
    items: [
      "Custom GPU buildouts service launched",
      "Mobile responsive redesign across all pages",
      "On-demand GPU product page with live pricing",
    ],
  },
  {
    date: "February 2026",
    items: [
      "FRP tunneling replaces WireGuard for instance connectivity",
      "Real-time competitor pricing comparison",
      "Credit system and free trial support",
    ],
  },
  {
    date: "January 2026",
    items: [
      "GPU.ai platform launched",
      "Initial provider integrations",
      "Per-second billing engine",
      "SSH key management dashboard",
    ],
  },
];

export default function ChangelogPage() {
  return (
    <div className="film-grain">
      <div className="stripe-lines" />
      <Navbar />

      <section className="relative pt-[88px]">
        <div className="absolute inset-0 -z-10">
          <div className="radial-glow absolute left-1/2 top-0 h-[600px] w-full -translate-x-1/2" />
        </div>
        <Container className="pb-16 pt-20 md:pb-24 md:pt-28">
          <p
            className="animate-fade-up type-ui-sm mb-4 font-medium uppercase tracking-[0.12em] text-purple-light"
            style={{ animationDelay: "0.1s" }}
          >
            Changelog
          </p>
          <h1
            className="animate-fade-up type-h1 max-w-[600px] font-bold text-white"
            style={{ animationDelay: "0.17s" }}
          >
            What&apos;s new
          </h1>
        </Container>
      </section>

      <section className="border-t border-border py-16 md:py-24">
        <Container>
          <div className="mx-auto max-w-[700px] space-y-12">
            {ENTRIES.map((entry) => (
              <div key={entry.date}>
                <h2 className="type-ui-sm mb-4 font-semibold uppercase tracking-[0.08em] text-purple-light">
                  {entry.date}
                </h2>
                <ul className="space-y-3">
                  {entry.items.map((item) => (
                    <li
                      key={item}
                      className="type-body flex items-start gap-3 text-text-muted"
                    >
                      <span className="mt-2.5 h-1 w-1 shrink-0 rounded-full bg-text-dim" />
                      {item}
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
        </Container>
      </section>

      <Footer />
    </div>
  );
}
