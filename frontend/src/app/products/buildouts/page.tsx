"use client";

import { useRef, useState } from "react";
import { Navbar } from "@/components/landing/Navbar";
import { Footer } from "@/components/landing/Footer";
import { Container } from "@/components/ui/Container";
import { Button } from "@/components/ui/Button";
import { Section, SectionHeader } from "@/components/ui/Section";
import { Card } from "@/components/ui/Card";
import Image from "next/image";
import Link from "next/link";

/* ── Form Options ── */

const COMPUTE_NEEDS = [
  "Dedicated Servers",
  "On Demand Instances",
  "GPU Procurement",
  "Superclusters",
  "AI Infra Consulting",
  "Other",
];

const CONTRACT_LENGTHS = [
  "Month-to-month",
  "3 months",
  "6 months",
  "1 year",
  "2+ years",
];

const TIMELINES = [
  "Immediately",
  "Within 2 weeks",
  "Within 1 month",
  "Within 3 months",
  "Just exploring",
];

const WORKLOADS = [
  "AI/ML Training",
  "Inference",
  "Fine-tuning",
  "HPC / Simulation",
  "Data Analytics",
  "Other",
];

/* ── Services ── */

const SERVICES = [
  {
    title: "GPU Procurement",
    description:
      "Access to NVIDIA B300, B200, H200, and H100 GPUs with volume pricing and vendor-negotiated discounts.",
    details: [
      "End-to-end logistics from sourcing to rack installation",
      "Hardware lifecycle management and upgrades",
    ],
  },
  {
    title: "Dedicated Servers",
    description:
      "Single-tenant bare-metal servers with full root access, custom hardware configurations, and guaranteed resources.",
    details: [
      "Custom hardware tailored to your workload",
      "24/7 monitoring and on-site support",
    ],
  },
  {
    title: "Deployment Support",
    description:
      "Infrastructure provisioning, OS/driver installation, performance benchmarking, and continuous monitoring.",
    details: [
      "Framework installation and tuning",
      "Proactive alerting and optimization",
    ],
  },
  {
    title: "AI Infra Consulting",
    description:
      "Architecture review, workload profiling, cost modeling, and migration planning from cloud to bare metal.",
    details: [
      "GPU selection guidance and TCO analysis",
      "Scalability roadmap and planning",
    ],
  },
];

/* ── Supercluster Tiers ── */

const TIERS = [
  {
    name: "GB300 NVL72",
    gpu: "NVIDIA GB300 NVL72",
    description:
      "Blackwell Ultra rack-scale architecture with 72 interconnected GPUs for unprecedented training throughput.",
    specs: [
      { label: "GPU Memory", value: "288 GB HBM3e per GPU" },
      { label: "Rack Scale", value: "72 GPUs per rack" },
      { label: "Interconnect", value: "5th-gen NVLink + NVSwitch" },
    ],
  },
  {
    name: "HGX H200",
    gpu: "NVIDIA HGX H200",
    description:
      "Enhanced Hopper architecture with HBM3e memory for large-scale training and high-throughput inference.",
    specs: [
      { label: "GPU Memory", value: "141 GB HBM3e per GPU" },
      { label: "Cluster Sizes", value: "128 \u2013 4,096 GPUs" },
      { label: "Interconnect", value: "NVLink + InfiniBand NDR" },
    ],
  },
  {
    name: "Custom",
    gpu: "Mixed / Multi-generation",
    description:
      "Tailored supercluster topologies designed around your workloads, performance targets, and budget.",
    specs: [
      { label: "GPU Options", value: "B300, B200, H200, H100+" },
      { label: "Cluster Sizes", value: "Custom to your needs" },
      { label: "Support", value: "Dedicated engineering team" },
    ],
  },
];

/* ── Partners ── */

const PARTNERS = [
  { name: "NVIDIA", logo: "/logos/nvidia.png", width: 120, height: 48 },
  { name: "Supermicro", logo: "/logos/supermicro.png", width: 140, height: 48 },
  { name: "NovaCore", logo: "/logos/novacore.png", width: 120, height: 48 },
];

/* ── Form Field ── */

function FormField({
  label,
  required,
  children,
}: {
  label: string;
  required?: boolean;
  children: React.ReactNode;
}) {
  return (
    <div>
      <label className="type-ui-2xs mb-2 block font-medium uppercase tracking-[0.08em] text-text-dim">
        {label}
        {required && <span className="text-purple-light"> *</span>}
      </label>
      {children}
    </div>
  );
}

/* ── Page ── */

export default function BuildoutsPage() {
  const [submitted, setSubmitted] = useState(false);
  const [showError, setShowError] = useState(false);
  const formRef = useRef<HTMLFormElement>(null);

  const inputClasses =
    "type-ui w-full rounded-lg border border-border bg-bg-alt px-4 py-3 text-white outline-none transition-colors placeholder:text-text-dim focus:border-purple/60 focus:ring-1 focus:ring-purple/30";
  const selectClasses =
    "type-ui w-full appearance-none rounded-lg border border-border bg-bg-alt px-4 py-3 text-white outline-none transition-colors focus:border-purple/60 focus:ring-1 focus:ring-purple/30 cursor-pointer";

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
            Custom Buildouts
          </p>
          <h1
            className="animate-fade-up type-h1 mx-auto max-w-[800px] font-bold text-white"
            style={{ animationDelay: "0.17s" }}
          >
            Enterprise GPU infrastructure, built for you
          </h1>
          <p
            className="animate-fade-up type-body-lg mx-auto mt-5 max-w-[620px] text-text-muted"
            style={{ animationDelay: "0.24s" }}
          >
            From hardware procurement to deployment and ongoing support.
            Partner with NVIDIA, Supermicro, and NovaCore through GPU.ai.
          </p>
          <div
            className="animate-fade-up mt-8 flex flex-wrap items-center justify-center gap-4"
            style={{ animationDelay: "0.31s" }}
          >
            <Button href="#contact" className="gradient-btn">
              Get in Touch
            </Button>
            <Button variant="secondary" href="/products/on-demand">
              On-Demand GPUs
            </Button>
          </div>
        </Container>
      </section>

      {/* Partners */}
      <section className="border-t border-border py-12">
        <Container>
          <p className="type-ui-2xs mb-8 text-center font-medium uppercase tracking-[0.1em] text-text-dim">
            In partnership with
          </p>
          <div className="flex flex-wrap items-center justify-center gap-12 md:gap-20 -translate-x-4">
            {PARTNERS.map((p) => (
              <Image
                key={p.name}
                src={p.logo}
                alt={p.name}
                width={p.width}
                height={p.height}
                className="h-8 w-auto opacity-60 brightness-0 invert transition-opacity hover:opacity-100 md:h-10"
              />
            ))}
          </div>
        </Container>
      </section>

      {/* Services */}
      <Section id="services">
        <SectionHeader
          label="What We Offer"
          title="End-to-end GPU infrastructure"
          description="Hardware, deployment, and ongoing support under one roof."
        />
        <div className="grid gap-6 md:grid-cols-2">
          {SERVICES.map((s) => (
            <Card key={s.title}>
              <h3 className="type-h5 mb-2 font-semibold text-white">
                {s.title}
              </h3>
              <p className="type-body mb-4 leading-[1.7] text-text-muted">
                {s.description}
              </p>
              <ul className="space-y-1.5">
                {s.details.map((d) => (
                  <li
                    key={d}
                    className="type-ui-sm flex items-start gap-2 text-text-dim"
                  >
                    <span className="mt-1.5 h-1 w-1 shrink-0 rounded-full bg-purple-light/60" />
                    {d}
                  </li>
                ))}
              </ul>
            </Card>
          ))}
        </div>
      </Section>

      {/* Supercluster Tiers */}
      <Section id="configurations">
        <SectionHeader
          label="Configurations"
          title="Supercluster options"
          description="Pre-designed architectures or fully custom topologies."
        />
        <div className="grid gap-6 md:grid-cols-3">
          {TIERS.map((t) => (
            <Link key={t.name} href="#contact" className="block">
              <Card className="h-full cursor-pointer">
                <p className="type-ui-2xs mb-1 font-medium uppercase tracking-[0.08em] text-purple-light">
                  {t.gpu}
                </p>
                <h3 className="type-h5 mb-3 font-semibold text-white">
                  {t.name}
                </h3>
                <p className="type-body mb-4 text-text-muted">{t.description}</p>
                <div className="space-y-3 border-t border-border pt-4">
                  {t.specs.map((sp) => (
                    <div key={sp.label}>
                      <span className="type-ui-2xs block font-medium uppercase tracking-[0.08em] text-text-dim">{sp.label}</span>
                      <span className="type-ui-sm font-medium text-text-muted">
                        {sp.value}
                      </span>
                    </div>
                  ))}
                </div>
              </Card>
            </Link>
          ))}
        </div>
      </Section>

      {/* Contact Form */}
      <Section id="contact">
        <SectionHeader
          label="Get Started"
          title="Tell us about your project"
          description="Our team typically responds within one business day."
        />

        {submitted ? (
          <div className="mx-auto max-w-[600px] text-center">
            <div className="mb-4 inline-flex h-14 w-14 items-center justify-center rounded-full bg-green-dim">
              <svg
                className="h-7 w-7 text-green"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2}
              >
                <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <h3 className="type-h4 mb-2 font-semibold text-white">
              We&apos;ll be in touch
            </h3>
            <p className="type-body text-text-muted">
              Our team typically responds within one business day.
            </p>
            <Link
              href="/"
              className="type-ui-sm mt-6 inline-block text-purple-light transition-colors hover:text-white"
            >
              &larr; Back to home
            </Link>
          </div>
        ) : (
          <form
            ref={formRef}
            noValidate
            onSubmit={async (e) => {
              e.preventDefault();
              if (!formRef.current?.checkValidity()) {
                setShowError(true);
                return;
              }
              setShowError(false);
              const fd = new FormData(formRef.current);
              try {
                await fetch("/api/buildouts", {
                  method: "POST",
                  headers: { "Content-Type": "application/json" },
                  body: JSON.stringify({
                    firstName: fd.get("firstName"),
                    lastName: fd.get("lastName"),
                    email: fd.get("email"),
                    phone: fd.get("phone"),
                    company: fd.get("company"),
                    jobTitle: fd.get("jobTitle"),
                    computeNeeds: fd.get("computeNeeds"),
                    contractLength: fd.get("contractLength"),
                    timeline: fd.get("timeline"),
                    workload: fd.get("workload"),
                  }),
                });
              } catch {
                // still show success — email may have sent server-side
              }
              setSubmitted(true);
            }}
            className="mx-auto max-w-[800px] space-y-6"
          >
            {/* Row 1: Name */}
            <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
              <FormField label="First Name" required>
                <input
                  name="firstName"
                  type="text"
                  required
                  className={inputClasses}
                />
              </FormField>
              <FormField label="Last Name" required>
                <input
                  name="lastName"
                  type="text"
                  required
                  className={inputClasses}
                />
              </FormField>
            </div>

            {/* Row 2: Email + Phone */}
            <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
              <FormField label="Business Email" required>
                <input
                  name="email"
                  type="email"
                  required
                  className={inputClasses}
                />
              </FormField>
              <FormField label="Phone Number" required>
                <input
                  name="phone"
                  type="tel"
                  required
                  className={inputClasses}
                />
              </FormField>
            </div>

            {/* Row 3: Company + Job Title */}
            <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
              <FormField label="Company Name" required>
                <input
                  name="company"
                  type="text"
                  required
                  className={inputClasses}
                />
              </FormField>
              <FormField label="Job Title" required>
                <input
                  name="jobTitle"
                  type="text"
                  required
                  className={inputClasses}
                />
              </FormField>
            </div>

            {/* Row 4: Compute Needs + Contract */}
            <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
              <FormField label="What are your compute needs?" required>
                <select
                  name="computeNeeds"
                  required
                  defaultValue=""
                  className={selectClasses}
                >
                  <option value="" disabled className="bg-bg">
                    Please select
                  </option>
                  {COMPUTE_NEEDS.map((opt) => (
                    <option key={opt} value={opt} className="bg-bg">
                      {opt}
                    </option>
                  ))}
                </select>
              </FormField>
              <FormField label="Desired contract length?" required>
                <select
                  name="contractLength"
                  required
                  defaultValue=""
                  className={selectClasses}
                >
                  <option value="" disabled className="bg-bg">
                    Please select
                  </option>
                  {CONTRACT_LENGTHS.map((opt) => (
                    <option key={opt} value={opt} className="bg-bg">
                      {opt}
                    </option>
                  ))}
                </select>
              </FormField>
            </div>

            {/* Row 5: Timeline + Workload */}
            <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
              <FormField label="How soon do you need this?" required>
                <select
                  name="timeline"
                  required
                  defaultValue=""
                  className={selectClasses}
                >
                  <option value="" disabled className="bg-bg">
                    Please select
                  </option>
                  {TIMELINES.map((opt) => (
                    <option key={opt} value={opt} className="bg-bg">
                      {opt}
                    </option>
                  ))}
                </select>
              </FormField>
              <FormField label="What is your main workload?" required>
                <select
                  name="workload"
                  required
                  defaultValue=""
                  className={selectClasses}
                >
                  <option value="" disabled className="bg-bg">
                    Please select
                  </option>
                  {WORKLOADS.map((opt) => (
                    <option key={opt} value={opt} className="bg-bg">
                      {opt}
                    </option>
                  ))}
                </select>
              </FormField>
            </div>

            {/* Consent */}
            <div className="space-y-3 pt-2">
              <p className="type-ui-sm leading-relaxed text-text-dim">
                By clicking submit, you consent to allow GPU.ai to store and
                process the personal information submitted above to provide you
                the content requested. See our{" "}
                <Link
                  href="/privacy"
                  className="text-purple-light underline transition-colors hover:text-white"
                >
                  privacy policy
                </Link>
                .
              </p>
              <label className="type-ui-sm flex cursor-pointer items-center gap-3">
                <input
                  type="checkbox"
                  className="h-4 w-4 accent-purple rounded border border-border bg-transparent"
                />
                <span className="text-text-muted">
                  I agree to receive other communications from GPU.ai.
                </span>
              </label>
            </div>

            {/* Validation error */}
            {showError && (
              <p className="type-ui-sm font-medium text-red-400">
                Please complete all required fields.
              </p>
            )}

            {/* Submit */}
            <div className="pt-2">
              <Button type="submit" className="gradient-btn">
                Submit Inquiry
              </Button>
            </div>
          </form>
        )}
      </Section>

      <Footer />
    </div>
  );
}
