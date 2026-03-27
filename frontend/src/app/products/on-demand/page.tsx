"use client";

import { useEffect, useRef } from "react";
import { Navbar } from "@/components/landing/Navbar";
import { Footer } from "@/components/landing/Footer";
import { HowItWorks } from "@/components/landing/HowItWorks";
import { Container } from "@/components/ui/Container";
import { Button } from "@/components/ui/Button";
import { Section, SectionHeader } from "@/components/ui/Section";
import createGlobe from "cobe";

/* ── Stats ── */

const STATS = [
  { value: "30%", label: "Average savings vs. hyperscalers" },
  { value: "<60s", label: "From launch to SSH access" },
  { value: "24/7", label: "Availability monitoring" },
];

/* ── GPU Models ── */

const GPUS = [
  { model: "H200 SXM", vram: "141 GB HBM3e", from: "$3.81/hr", savings: "28%" },
  { model: "H100 SXM", vram: "80 GB HBM3e", from: "$2.64/hr", savings: "24%" },
  { model: "B200", vram: "192 GB HBM3e", from: "$5.29/hr", savings: "18%" },
  { model: "A100 80GB", vram: "80 GB HBM2e", from: "$2.00/hr", savings: "8%" },
];

/* ── Globe Component ── */

function Globe() {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    if (!canvasRef.current) return;

    let phi = 0;
    let animId: number;
    const canvas = canvasRef.current;
    const width = canvas.offsetWidth;

    const globe = createGlobe(canvas, {
      devicePixelRatio: 2,
      width: width * 2,
      height: width * 2,
      phi: 0,
      theta: 0.2,
      dark: 1,
      diffuse: 1.2,
      mapSamples: 16000,
      mapBrightness: 6,
      baseColor: [0.15, 0.12, 0.25],
      markerColor: [0.49, 0.42, 0.94],
      glowColor: [0.12, 0.10, 0.22],
      markers: [
        { location: [37.7749, -122.4194], size: 0.06 },
        { location: [40.7128, -74.006], size: 0.06 },
        { location: [51.5074, -0.1278], size: 0.05 },
        { location: [1.3521, 103.8198], size: 0.05 },
        { location: [35.6762, 139.6503], size: 0.04 },
        { location: [50.1109, 8.6821], size: 0.05 },
        { location: [19.076, 72.8777], size: 0.04 },
      ],
    });

    function tick() {
      phi += 0.003;
      globe.update({ phi });
      animId = requestAnimationFrame(tick);
    }
    animId = requestAnimationFrame(tick);

    // Fade in
    requestAnimationFrame(() => { canvas.style.opacity = "1"; });

    return () => {
      cancelAnimationFrame(animId);
      globe.destroy();
    };
  }, []);

  return (
    <div className="aspect-square w-full">
      <canvas
        ref={canvasRef}
        className="h-full w-full"
        style={{ contain: "layout paint size", opacity: 0, transition: "opacity 1s ease" }}
      />
    </div>
  );
}

/* ── Page ── */

export default function OnDemandPage() {
  return (
    <div className="film-grain">
      <div className="stripe-lines" />
      <Navbar />

      {/* Hero */}
      <section className="relative pt-[88px]">
        <div className="absolute inset-0 -z-10">
          <div className="radial-glow absolute left-1/2 top-0 h-[600px] w-full -translate-x-1/2" />
        </div>
        <Container className="pb-16 pt-20 text-center md:pb-24 md:pt-28">
          <p
            className="animate-fade-up type-ui-sm mb-4 font-medium uppercase tracking-[0.12em] text-purple-light"
            style={{ animationDelay: "0.1s" }}
          >
            On-Demand GPUs
          </p>
          <h1
            className="animate-fade-up type-h1 mx-auto max-w-[800px] font-bold text-white"
            style={{ animationDelay: "0.17s" }}
          >
            On-demand GPUs, best price guaranteed
          </h1>
          <p
            className="animate-fade-up type-body-lg mx-auto mt-5 max-w-[600px] text-text-muted"
            style={{ animationDelay: "0.24s" }}
          >
            We aggregate GPU inventory from providers worldwide to find you the
            best price. Deploy in seconds, pay by the second.
          </p>
          <div
            className="animate-fade-up mt-8 flex flex-wrap items-center justify-center gap-4"
            style={{ animationDelay: "0.31s" }}
          >
            <Button href="/free-trial" className="gradient-btn">
              $100 Free Trial
            </Button>
            <Button variant="secondary" href="/#pricing">
              See Pricing
            </Button>
          </div>
        </Container>
      </section>

      {/* Stats */}
      <section className="border-t border-border py-16 md:py-20">
        <Container>
          <div className="grid grid-cols-3 divide-x divide-border">
            {STATS.map((s) => (
              <div key={s.label} className="px-4 text-center md:px-8">
                <span className="block font-sans text-[36px] font-bold tracking-tight text-purple-light md:text-[48px]">
                  {s.value}
                </span>
                <span className="type-ui-sm mt-1 block text-text-dim">
                  {s.label}
                </span>
              </div>
            ))}
          </div>
        </Container>
      </section>

      {/* Global GPU Network */}
      <Section id="globe-section">
        <div className="mx-auto grid max-w-[1000px] items-center gap-12 lg:grid-cols-2">
          <div>
            <p className="type-ui-2xs mb-3 font-medium uppercase tracking-[0.1em] text-purple-light">
              Global Network
            </p>
            <h2 className="type-h2 mb-5 font-bold text-white">
              Worldwide GPU aggregation
            </h2>
            <p className="type-body mb-6 leading-[1.8] text-text-muted">
              GPU.ai sources compute from data centers across North America,
              Europe, and Asia. Our platform continuously scans availability and
              pricing, routing your workloads to the best option in real time.
            </p>
            <ul className="space-y-3">
              {[
                "Data centers in US, EU, and Asia-Pacific",
                "Real-time price and availability monitoring",
                "Automatic routing to lowest-cost provider",
                "No vendor lock-in — switch providers instantly",
              ].map((item) => (
                <li
                  key={item}
                  className="type-ui-sm flex items-start gap-2.5 text-text-dim"
                >
                  <svg
                    className="mt-0.5 h-4 w-4 shrink-0 text-purple-light"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    strokeWidth={2}
                  >
                    <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                  {item}
                </li>
              ))}
            </ul>
          </div>
          <div className="flex items-center justify-center">
            <Globe />
          </div>
        </div>
      </Section>

      {/* Available GPUs */}
      <Section id="gpus">
        <SectionHeader
          label="Available Models"
          title="Enterprise GPUs, on demand"
          description="Pricing updates in real time across all providers."
        />
        <div className="mx-auto max-w-[800px]">
          <div className="overflow-hidden rounded-xl border border-border">
            <table className="w-full border-collapse text-left">
              <thead>
                <tr className="border-b border-border bg-bg-alt">
                  <th className="type-ui-sm px-6 py-4 font-medium uppercase tracking-[0.05em] text-text-dim">
                    Model
                  </th>
                  <th className="type-ui-sm px-6 py-4 font-medium uppercase tracking-[0.05em] text-text-dim">
                    VRAM
                  </th>
                  <th className="type-ui-sm px-6 py-4 font-medium uppercase tracking-[0.05em] text-text-dim">
                    From
                  </th>
                  <th className="type-ui-sm px-6 py-4 font-medium uppercase tracking-[0.05em] text-text-dim">
                    Savings
                  </th>
                </tr>
              </thead>
              <tbody>
                {GPUS.map((g) => (
                  <tr
                    key={g.model}
                    className="border-b border-border last:border-0 transition-colors hover:bg-bg-card-hover"
                  >
                    <td className="px-6 py-5 text-[15px] font-medium text-white">
                      {g.model}
                    </td>
                    <td className="type-ui px-6 py-5 text-text-muted">
                      {g.vram}
                    </td>
                    <td className="px-6 py-5 text-[15px] font-semibold text-purple-light">
                      {g.from}
                    </td>
                    <td className="px-6 py-5">
                      <span className="type-ui-sm inline-flex items-center rounded-full bg-green-dim px-3 py-1 font-semibold text-green">
                        {g.savings}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </Section>

      {/* How it works — reuse animated component from landing page */}
      <HowItWorks />

      {/* CTA */}
      <section className="relative py-24 md:py-32">
        <div className="absolute inset-0 -z-10">
          <div className="radial-glow absolute left-1/2 top-1/2 h-[500px] w-full -translate-x-1/2 -translate-y-1/2" />
        </div>
        <Container className="text-center">
          <h2 className="type-h2 mx-auto max-w-[600px] font-bold text-white">
            Start building with
            <br />
            on-demand GPUs today
          </h2>
          <div className="mt-8 flex flex-wrap items-center justify-center gap-4">
            <Button href="/free-trial" className="gradient-btn">
              $100 Free Trial
            </Button>
            <Button variant="secondary" href="/cloud/gpu-availability">
              Launch Instance
            </Button>
          </div>
        </Container>
      </section>

      <Footer />
    </div>
  );
}
