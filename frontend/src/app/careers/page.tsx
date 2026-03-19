import { Navbar } from "@/components/landing/Navbar";
import { Footer } from "@/components/landing/Footer";
import { Container } from "@/components/ui/Container";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Careers",
};

const PERKS = [
  { title: "Remote-first", desc: "Work from anywhere. We coordinate async across time zones." },
{ title: "Cutting-edge stack", desc: "Go, React, NVIDIA hardware — production infra, not prototypes." },
  { title: "Small team, big leverage", desc: "Your work ships to users the same week. No layers of approval." },
];

const ROLES = [
  {
    title: "Backend Engineer",
    type: "Full-time · Remote",
    description:
      "Build the core Go services — availability polling, provisioning orchestration, billing, and the public API. You'll work directly with GPU cloud provider APIs and shape the system architecture.",
  },
  {
    title: "Frontend Engineer",
    type: "Full-time · Remote",
    description:
      "Own the Next.js dashboard and landing experience. Build real-time GPU availability views, instance management, and the pricing comparison tools that help customers make decisions.",
  },
  {
    title: "Infrastructure / DevOps Engineer",
    type: "Full-time · Remote",
    description:
      "Manage multi-cloud networking, tunneling, deployment pipelines, and monitoring. You'll ensure sub-60-second deploys and 99.9% uptime across provider integrations.",
  },
];

export default function CareersPage() {
  return (
    <div className="film-grain min-h-screen">
      <div className="stripe-lines" />
      <Navbar />

      {/* Hero */}
      <section className="relative pt-[88px]">
        <div className="radial-glow pointer-events-none absolute inset-0" />
        <Container className="relative z-10 pb-20 pt-24 md:pt-32">
          <p className="animate-fade-up font-mono text-[11px] font-semibold uppercase tracking-[0.14em] text-purple-light">
            Careers at GPU.ai
          </p>
          <h1
            className="type-display animate-fade-up mt-4 max-w-[700px] font-bold text-white"
            style={{ animationDelay: "0.08s" }}
          >
            Build the infra layer
            <br />
            <span className="gradient-text">AI runs on.</span>
          </h1>
          <p
            className="animate-fade-up mt-6 max-w-[540px] font-mono text-[15px] font-normal leading-relaxed text-text-muted"
            style={{ animationDelay: "0.14s" }}
          >
            We&apos;re a small founding team building the aggregation layer for
            GPU compute. If you want to ship fast, own large surfaces, and work
            on real infrastructure — not wrappers — we want to talk.
          </p>
        </Container>
      </section>

      {/* Why GPU.ai */}
      <section className="border-t border-border">
        <Container className="py-20 md:py-28">
          <p className="font-mono text-[11px] font-semibold uppercase tracking-[0.14em] text-text-dim">
            Why GPU.ai
          </p>
          <h2 className="type-h3 mt-4 max-w-[500px] font-bold text-white">
            Early-stage, real revenue, hard problems.
          </h2>
          <div className="mt-12 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {PERKS.map((perk) => (
              <div
                key={perk.title}
                className="rounded-[10px] border border-border bg-bg-card p-5"
              >
                <div className="mb-3 flex h-8 w-8 items-center justify-center rounded-md bg-purple-dim">
                  <div className="h-2 w-2 rounded-full bg-purple" />
                </div>
                <h3 className="text-[14px] font-semibold text-white">
                  {perk.title}
                </h3>
                <p className="mt-2 font-mono text-[12px] font-normal leading-relaxed text-text-dim">
                  {perk.desc}
                </p>
              </div>
            ))}
          </div>
        </Container>
      </section>

      {/* Open Roles */}
      <section className="border-t border-border">
        <Container className="py-20 md:py-28">
          <p className="font-mono text-[11px] font-semibold uppercase tracking-[0.14em] text-text-dim">
            Open Roles
          </p>
          <h2 className="type-h3 mt-4 font-bold text-white">
            Current openings
          </h2>
          <div className="mt-12 flex flex-col gap-4">
            {ROLES.map((role) => (
              <a
                key={role.title}
                href={`/careers/apply?role=${encodeURIComponent(role.title)}`}
                className="group rounded-[10px] border border-border bg-bg-card p-6 transition-all hover:border-border-light hover:bg-bg-card-hover"
              >
                <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                  <div className="flex-1">
                    <h3 className="text-[16px] font-semibold text-white">
                      {role.title}
                    </h3>
                    <p className="mt-0.5 font-mono text-[12px] font-normal text-purple-light">
                      {role.type}
                    </p>
                    <p className="mt-3 max-w-[600px] font-mono text-[13px] font-normal leading-relaxed text-text-dim">
                      {role.description}
                    </p>
                  </div>
                  <span className="flex shrink-0 items-center gap-1.5 font-mono text-[12px] font-medium text-text-dim transition-colors group-hover:text-purple-light sm:mt-1">
                    Apply
                    <svg
                      className="h-3 w-3 transition-transform group-hover:translate-x-0.5"
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
                </div>
              </a>
            ))}
          </div>
        </Container>
      </section>

      {/* General interest */}
      <section className="relative border-t border-border">
        <div className="radial-glow pointer-events-none absolute inset-0" />
        <Container className="relative z-10 py-20 text-center md:py-28">
          <h2 className="type-h2 font-bold text-white">
            Don&apos;t see your role?
          </h2>
          <p className="mx-auto mt-3 max-w-[440px] font-mono text-[14px] font-normal text-text-muted">
            We&apos;re always looking for exceptional people. Send us a note
            with what you&apos;d want to build.
          </p>
          <div className="mt-8">
            <a
              href="/careers/apply"
              className="gradient-btn inline-flex items-center justify-center rounded-lg px-6 py-3 text-[15px] font-semibold text-white"
            >
              Get in Touch
            </a>
          </div>
        </Container>
      </section>

      <Footer />
    </div>
  );
}
