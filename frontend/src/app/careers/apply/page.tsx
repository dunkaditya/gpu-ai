"use client";

import { Suspense, useRef, useState } from "react";
import { useSearchParams } from "next/navigation";
import { Container } from "@/components/ui/Container";
import { ChipLogo } from "@/components/ui/ChipLogo";
import Link from "next/link";

const ROLES = [
  "Backend Engineer",
  "Frontend Engineer",
  "Infrastructure / DevOps Engineer",
  "Other",
];

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

function SelectField({
  name,
  label,
  options,
  defaultValue,
}: {
  name: string;
  label: string;
  options: string[];
  defaultValue?: string;
}) {
  return (
    <div className="group">
      <label className="mb-2 block font-mono text-[11px] font-semibold uppercase tracking-[0.12em] text-text-dim">
        {label}
        <span className="ml-0.5 text-purple">*</span>
      </label>
      <div className="relative">
        <select
          name={name}
          defaultValue={defaultValue || ""}
          required
          className="w-full appearance-none border-b border-border bg-transparent pb-2.5 pt-1 font-mono text-[14px] font-medium text-white outline-none transition-colors focus:border-purple"
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

export default function ApplyPage() {
  return (
    <Suspense>
      <ApplyPageInner />
    </Suspense>
  );
}

function ApplyPageInner() {
  const searchParams = useSearchParams();
  const roleParam = searchParams.get("role") || "";
  const [submitted, setSubmitted] = useState(false);
  const [showError, setShowError] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [fileName, setFileName] = useState<string | null>(null);
  const formRef = useRef<HTMLFormElement>(null);
  const fileRef = useRef<HTMLInputElement>(null);

  if (submitted) {
    return (
      <div className="film-grain min-h-screen">
        <div className="stripe-lines" />
        <Container className="relative z-10 flex min-h-screen flex-col items-center justify-center">
          <div
            className="animate-fade-up text-center"
            style={{ animationDelay: "0.05s" }}
          >
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
              Application received.
            </h1>
            <p className="type-body-lg mt-4 font-normal text-text-muted">
              We&apos;ll review your application and get back to you soon.
            </p>
            <Link
              href="/careers"
              className="mt-10 inline-flex items-center gap-2 font-mono text-[13px] font-medium text-text-muted transition-colors hover:text-white"
            >
              <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
              </svg>
              Back to careers
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
        <Link
          href="/careers"
          className="inline-flex items-center gap-2 font-mono text-[13px] font-medium text-text-muted transition-colors hover:text-white"
        >
          <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
          Back to careers
        </Link>

        <div className="mt-12 grid grid-cols-1 gap-16 lg:grid-cols-[1fr_520px] lg:gap-24">
          {/* Left — value prop */}
          <div className="animate-fade-up" style={{ animationDelay: "0.1s" }}>
            <div className="mb-6 flex items-center gap-3">
              <ChipLogo size={28} />
              <span className="font-mono text-[11px] font-semibold uppercase tracking-[0.14em] text-purple-light">
                Careers
              </span>
            </div>

            <h1 className="type-h1 font-sans font-bold text-white">
              Join the team.
              <br />
              <span className="gradient-text">Ship real infra.</span>
            </h1>

            <p className="mt-6 max-w-[480px] font-mono text-[15px] font-normal leading-relaxed text-text-muted">
              We&apos;re building the aggregation layer for GPU compute. Small team, hard problems, massive market. Your work ships to production the same week.
            </p>

            <div className="mt-12 grid grid-cols-2 gap-6">
              {[
                { label: "Remote-first", desc: "Work from anywhere" },
                { label: "Cutting-edge stack", desc: "Go, React, NVIDIA hardware" },
                { label: "Small team", desc: "Your work ships fast" },
                { label: "Real infrastructure", desc: "Not wrappers or prototypes" },
              ].map((perk, i) => (
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
          </div>

          {/* Right — form */}
          <div
            className="animate-fade-up"
            style={{ animationDelay: "0.15s" }}
          >
            <div className="rounded-[12px] border border-border bg-bg-card/60 p-8 backdrop-blur-sm lg:p-10">
              <h2 className="mb-1 font-sans text-[22px] font-bold tracking-[-0.02em] text-white">
                Apply
              </h2>
              <p className="mb-8 font-mono text-[13px] font-normal text-text-dim">
                Tell us about yourself and we&apos;ll be in touch.
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
                    await fetch("/api/apply", {
                      method: "POST",
                      body: fd,
                    });
                  } catch {
                    // still show success
                  }
                  setSubmitting(false);
                  setSubmitted(true);
                }}
                className="space-y-6"
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
                    label="Email"
                    type="email"
                    required
                    showError={showError}
                  />
                  <InputField
                    name="phone"
                    label="Phone"
                    type="tel"
                    showError={showError}
                  />
                </div>

                {/* Role */}
                <SelectField
                  name="role"
                  label="Role"
                  options={ROLES}
                  defaultValue={roleParam}
                />

                {/* LinkedIn */}
                <InputField
                  name="linkedin"
                  label="LinkedIn URL"
                  showError={showError}
                />

                <div className="border-t border-border/50" />

                {/* Resume upload */}
                <div className="group">
                  <label className="mb-2 block font-mono text-[11px] font-semibold uppercase tracking-[0.12em] text-text-dim">
                    Resume
                  </label>
                  <input
                    ref={fileRef}
                    name="resume"
                    type="file"
                    accept=".pdf,.doc,.docx"
                    className="hidden"
                    onChange={(e) => {
                      const file = e.target.files?.[0];
                      setFileName(file ? file.name : null);
                    }}
                  />
                  <button
                    type="button"
                    onClick={() => fileRef.current?.click()}
                    className="flex w-full items-center gap-3 rounded-lg border border-dashed border-border px-4 py-3 font-mono text-[13px] text-text-dim transition-colors hover:border-purple/50 hover:text-text-muted"
                  >
                    <svg className="h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="M12 16.5V9.75m0 0l3 3m-3-3l-3 3M6.75 19.5a4.5 4.5 0 01-1.41-8.775 5.25 5.25 0 0110.233-2.33 3 3 0 013.758 3.848A3.752 3.752 0 0118 19.5H6.75z" />
                    </svg>
                    {fileName || "Upload PDF, DOC, or DOCX"}
                  </button>
                </div>

                {/* Message */}
                <div className="group">
                  <label className="mb-2 block font-mono text-[11px] font-semibold uppercase tracking-[0.12em] text-text-dim">
                    Message
                  </label>
                  <textarea
                    name="message"
                    rows={4}
                    className="w-full resize-none border-b border-border bg-transparent pb-2.5 pt-1 font-mono text-[14px] font-medium text-white outline-none transition-colors placeholder:text-text-dim/40 focus:border-purple"
                    placeholder="Tell us what you'd want to build..."
                  />
                </div>

                {showError && (
                  <p className="font-mono text-[12px] font-medium text-red-400">
                    Please complete all required fields.
                  </p>
                )}

                <button
                  type="submit"
                  disabled={submitting}
                  className="gradient-btn w-full rounded-lg px-6 py-3.5 font-mono text-[14px] font-semibold text-white transition-all duration-200 hover:shadow-[0_0_28px_rgba(124,107,240,0.35)] disabled:opacity-50"
                >
                  {submitting ? "Submitting..." : "Submit Application"}
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
