"use client";

import { useRef, useState } from "react";
import { Container } from "@/components/ui/Container";
import { ChipLogo } from "@/components/ui/ChipLogo";
import Link from "next/link";

const GPU_MODELS = [
  "H100 80GB",
  "A100 80GB",
  "RTX 4090",
  "L40S",
  "RTX A6000",
  "Multiple Models",
  "Not Sure Yet",
];

const WORKLOADS = [
  "AI/ML Training",
  "Inference",
  "Fine-tuning",
  "HPC / Simulation",
  "Data Analytics",
  "Rendering",
  "Other",
];

const MONTHLY_SPEND = [
  "Under $1,000",
  "$1,000 – $5,000",
  "$5,000 – $25,000",
  "$25,000 – $100,000",
  "$100,000+",
];

const TIMELINES = [
  "Immediately",
  "Within 2 weeks",
  "Within 1 month",
  "Within 3 months",
  "Just exploring",
];

const TRIAL_PERKS = [
  { label: "Up to $100 free credits", desc: "No credit card required" },
  { label: "Priority onboarding", desc: "Dedicated setup assistance" },
  { label: "All GPU models", desc: "H100, A100, RTX 4090 & more" },
  { label: "Pay-as-you-go", desc: "No minimums or commitments" },
];

function SelectField({
  name,
  label,
  options,
  showError,
}: {
  name: string;
  label: string;
  options: string[];
  showError: boolean;
}) {
  return (
    <div className="group">
      <label className="mb-2 block font-mono text-[11px] font-semibold uppercase tracking-[0.12em] text-text-dim">
        {label}
      </label>
      <div className="relative">
        <select
          name={name}
          defaultValue=""
          className={`w-full appearance-none border-b bg-transparent pb-2.5 pt-1 font-mono text-[14px] font-medium text-white outline-none transition-colors ${
            showError
              ? "border-red-500/60"
              : "border-border focus:border-purple"
          }`}
        >
          <option value="" disabled className="bg-bg text-text-dim">
            Select...
          </option>
          {options.map((opt) => (
            <option key={opt} value={opt} className="bg-bg text-white">
              {opt}
            </option>
          ))}
        </select>
        <svg
          className="pointer-events-none absolute right-0 top-1/2 h-4 w-4 -translate-y-1/2 text-text-dim"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={2}
        >
          <path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" />
        </svg>
      </div>
    </div>
  );
}

function InputField({
  name,
  label,
  type = "text",
  required,
  showError,
}: {
  name: string;
  label: string;
  type?: string;
  required?: boolean;
  showError: boolean;
}) {
  return (
    <div className="group">
      <label className="mb-2 block font-mono text-[11px] font-semibold uppercase tracking-[0.12em] text-text-dim">
        {label}
        {required && <span className="ml-0.5 text-purple">*</span>}
      </label>
      <input
        name={name}
        type={type}
        required={required}
        className={`w-full border-b bg-transparent pb-2.5 pt-1 font-mono text-[14px] font-medium text-white outline-none transition-colors placeholder:text-text-dim/40 ${
          showError
            ? "border-red-500/60 invalid:border-red-500/60"
            : "border-border focus:border-purple"
        }`}
      />
    </div>
  );
}

export default function FreeTrialPage() {
  const [submitted, setSubmitted] = useState(false);
  const [showError, setShowError] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const formRef = useRef<HTMLFormElement>(null);

  if (submitted) {
    return (
      <div className="film-grain min-h-screen">
        <div className="stripe-lines" />
        <Container className="relative z-10 flex min-h-screen flex-col items-center justify-center">
          <div
            className="animate-fade-up text-center"
            style={{ animationDelay: "0.05s" }}
          >
            {/* Animated check */}
            <div className="mx-auto mb-8 flex h-20 w-20 items-center justify-center rounded-full border border-green/20 bg-green-dim">
              <svg
                className="h-10 w-10 text-green"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2.5}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M5 13l4 4L19 7"
                />
              </svg>
            </div>

            <h1 className="type-h2 font-sans font-bold text-white">
              Request received.
            </h1>
            <p className="type-body-lg mt-4 font-normal text-text-muted">
              We&apos;ll reach out within one business day to get you set up.
            </p>
            <Link
              href="/"
              className="mt-10 inline-flex items-center gap-2 font-mono text-[13px] font-medium text-text-muted transition-colors hover:text-white"
            >
              <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
              </svg>
              Back to home
            </Link>
          </div>
        </Container>
      </div>
    );
  }

  return (
    <div className="film-grain min-h-screen">
      <div className="stripe-lines" />

      <Container className="relative z-10 pb-24 pt-12">
        {/* Back link */}
        <Link
          href="/"
          className="inline-flex items-center gap-2 font-mono text-[13px] font-medium text-text-muted transition-colors hover:text-white"
        >
          <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
          Back to home
        </Link>

        {/* Two-column layout */}
        <div className="mt-12 grid grid-cols-1 gap-16 lg:grid-cols-[1fr_520px] lg:gap-24">
          {/* Left — value prop */}
          <div className="animate-fade-up" style={{ animationDelay: "0.1s" }}>
            <div className="mb-6 flex items-center gap-3">
              <ChipLogo size={28} />
              <span className="font-mono text-[11px] font-semibold uppercase tracking-[0.14em] text-purple-light">
                Free Trial
              </span>
            </div>

            <h1 className="type-h1 font-sans font-bold text-white">
              Get compute.
              <br />
              <span className="gradient-text">Skip the markup.</span>
            </h1>

            <p className="mt-6 max-w-[480px] font-mono text-[15px] font-normal leading-relaxed text-text-muted">
              Spin up H100s, A100s, and RTX 4090s on demand — billed per second with no long-term contracts. One API, one dashboard, every major GPU cloud behind it.
            </p>

            {/* Perks grid */}
            <div className="mt-12 grid grid-cols-2 gap-6">
              {TRIAL_PERKS.map((perk, i) => (
                <div
                  key={perk.label}
                  className="animate-fade-up"
                  style={{ animationDelay: `${0.2 + i * 0.06}s` }}
                >
                  <div className="mb-1.5 flex items-center gap-2">
                    <div className="h-1.5 w-1.5 rounded-full bg-purple" />
                    <span className="font-mono text-[13px] font-semibold text-white">
                      {perk.label}
                    </span>
                  </div>
                  <p className="pl-[14px] font-mono text-[12px] font-normal text-text-dim">
                    {perk.desc}
                  </p>
                </div>
              ))}
            </div>

            {/* Trust line */}
            <div className="mt-16 border-t border-border pt-6">
              <p className="font-mono text-[12px] font-normal leading-relaxed text-text-dim">
                Trusted by teams running training, inference, and fine-tuning
                workloads on NVIDIA hardware. SOC 2 compliant infrastructure.
              </p>
            </div>
          </div>

          {/* Right — form */}
          <div
            className="animate-fade-up"
            style={{ animationDelay: "0.15s" }}
          >
            <div className="rounded-[12px] border border-border bg-bg-card/60 p-8 backdrop-blur-sm lg:p-10">
              <h2 className="mb-1 font-sans text-[22px] font-bold tracking-[-0.02em] text-white">
                Request free trial
              </h2>
              <p className="mb-8 font-mono text-[13px] font-normal text-text-dim">
                Fill out the form and we&apos;ll get you started.
              </p>

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
                  setSubmitting(true);
                  const fd = new FormData(formRef.current);
                  try {
                    await fetch("/api/trial", {
                      method: "POST",
                      headers: { "Content-Type": "application/json" },
                      body: JSON.stringify({
                        firstName: fd.get("firstName"),
                        lastName: fd.get("lastName"),
                        email: fd.get("email"),
                        phone: fd.get("phone"),
                        company: fd.get("company"),
                        jobTitle: fd.get("jobTitle"),
                        gpuModel: fd.get("gpuModel"),
                        workload: fd.get("workload"),
                        monthlySpend: fd.get("monthlySpend"),
                        timeline: fd.get("timeline"),
                      }),
                    });
                  } catch {
                    // still show success — email may have sent server-side
                  }
                  setSubmitting(false);
                  setSubmitted(true);
                }}
                className="space-y-6"
                data-show-errors={showError || undefined}
              >
                {/* Name row */}
                <div className="grid grid-cols-2 gap-4">
                  <InputField
                    name="firstName"
                    label="First Name"
                    required
                    showError={showError}
                  />
                  <InputField
                    name="lastName"
                    label="Last Name"
                    required
                    showError={showError}
                  />
                </div>

                {/* Email + Phone */}
                <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                  <InputField
                    name="email"
                    label="Work Email"
                    type="email"
                    required
                    showError={showError}
                  />
                  <InputField
                    name="phone"
                    label="Phone"
                    type="tel"
                    required
                    showError={showError}
                  />
                </div>

                {/* Company + Title */}
                <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                  <InputField
                    name="company"
                    label="Company"
                    required
                    showError={showError}
                  />
                  <InputField
                    name="jobTitle"
                    label="Job Title"
                    required
                    showError={showError}
                  />
                </div>

                {/* Divider */}
                <div className="border-t border-border/50" />

                {/* GPU Model + Workload */}
                <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                  <SelectField
                    name="gpuModel"
                    label="GPU Model"
                    options={GPU_MODELS}
                    showError={false}
                  />
                  <SelectField
                    name="workload"
                    label="Workload Type"
                    options={WORKLOADS}
                    showError={false}
                  />
                </div>

                {/* Spend + Timeline */}
                <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                  <SelectField
                    name="monthlySpend"
                    label="Est. Monthly Spend"
                    options={MONTHLY_SPEND}
                    showError={false}
                  />
                  <SelectField
                    name="timeline"
                    label="Timeline"
                    options={TIMELINES}
                    showError={false}
                  />
                </div>

                {/* Error */}
                {showError && (
                  <p className="font-mono text-[12px] font-medium text-red-400">
                    Please complete all required fields.
                  </p>
                )}

                {/* Submit */}
                <button
                  type="submit"
                  disabled={submitting}
                  className="gradient-btn w-full rounded-lg px-6 py-3.5 font-mono text-[14px] font-semibold text-white transition-all duration-200 hover:shadow-[0_0_28px_rgba(124,107,240,0.35)] disabled:opacity-50"
                >
                  {submitting ? "Submitting..." : "Request Free Trial"}
                </button>

                <p className="text-center font-mono text-[11px] font-normal leading-relaxed text-text-dim">
                  By submitting, you agree to let GPU.ai store and process this
                  information. We&apos;ll never share your data with third parties.
                </p>
              </form>
            </div>
          </div>
        </div>
      </Container>
    </div>
  );
}
