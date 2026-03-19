"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import {
  motion,
  useInView,
  useReducedMotion,
} from "motion/react";
import { Section, SectionHeader } from "@/components/ui/Section";
import { HOW_IT_WORKS } from "@/lib/constants";

const CYCLE_MS = 2500;
const STEP_COUNT = HOW_IT_WORKS.length;

export function HowItWorks() {
  const [activeStep, setActiveStep] = useState(0);
  const [isMobile, setIsMobile] = useState(false);
  const sectionRef = useRef<HTMLDivElement>(null);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const pausedRef = useRef(false);

  const isInView = useInView(sectionRef, { amount: 0.15 });
  const prefersReduced = useReducedMotion();

  // Mobile detection
  useEffect(() => {
    const mq = window.matchMedia("(max-width: 767px)");
    setIsMobile(mq.matches);
    const handler = (e: MediaQueryListEvent) => setIsMobile(e.matches);
    mq.addEventListener("change", handler);
    return () => mq.removeEventListener("change", handler);
  }, []);

  // Auto-cycle
  const startCycle = useCallback(() => {
    if (intervalRef.current) clearInterval(intervalRef.current);
    intervalRef.current = setInterval(() => {
      if (!pausedRef.current) {
        setActiveStep((s) => (s + 1) % STEP_COUNT);
      }
    }, CYCLE_MS);
  }, []);

  useEffect(() => {
    if (isInView && !prefersReduced) {
      startCycle();
    } else {
      if (intervalRef.current) clearInterval(intervalRef.current);
    }
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [isInView, prefersReduced, startCycle]);

  const handleDotClick = (idx: number) => {
    setActiveStep(idx);
    startCycle(); // reset timer
  };

  const progressWidth = `${(activeStep / (STEP_COUNT - 1)) * 100}%`;

  // Reduced-motion: show all steps statically
  if (prefersReduced) {
    return (
      <Section id="how-it-works" className="pt-20 md:pt-24">
        <SectionHeader
          label="How it works"
          title="Three steps. Under a minute."
        />
        <div className="mx-auto max-w-[800px]">
          {/* Static rail */}
          <div className="relative mb-12 h-[2px]">
            <div className="absolute inset-0 bg-border" />
            <div className="absolute inset-y-0 left-0 bg-purple" style={{ width: "100%" }} />
            {HOW_IT_WORKS.map((_, i) => (
              <div
                key={i}
                className="absolute top-1/2 -translate-y-1/2"
                style={{ left: `${(i / (STEP_COUNT - 1)) * 100}%` }}
              >
                <div className="h-[14px] w-[14px] -translate-x-1/2 rounded-full bg-purple shadow-[0_0_12px_rgba(124,107,240,0.5)]" />
              </div>
            ))}
          </div>
          <div className="grid gap-8 md:grid-cols-3">
            {HOW_IT_WORKS.map((item) => (
              <div key={item.step} className="text-center">
                <span className="type-caption mb-2 block font-mono font-semibold tracking-widest text-purple-light uppercase">
                  Step 0{item.step}
                </span>
                <h3 className="type-h5 mb-2 font-semibold text-white">
                  {item.title}
                </h3>
                <p className="type-body-sm leading-[1.7] text-text-muted">
                  {item.description}
                </p>
              </div>
            ))}
          </div>
        </div>
      </Section>
    );
  }

  return (
    <Section id="how-it-works" className="pt-20 md:pt-24">
      <SectionHeader
        label="How it works"
        title="Three steps. Under a minute."
      />

      <div
        ref={sectionRef}
        onMouseEnter={() => { pausedRef.current = true; }}
        onMouseLeave={() => { pausedRef.current = false; }}
        className="mx-auto max-w-[800px]"
      >
        {/* ── Timeline Rail ── */}
        <div className="relative mb-12 hidden h-[2px] md:block">
          {/* Background track */}
          <div className="absolute inset-0 rounded-full bg-border" />

          {/* Progress fill with snake pulse */}
          <motion.div
            className="timeline-progress absolute inset-y-0 left-0 rounded-full bg-purple"
            animate={{ width: progressWidth }}
            transition={{ duration: 0.5, ease: [0.32, 0.72, 0, 1] }}
          />

          {/* Dots */}
          {HOW_IT_WORKS.map((_, i) => {
            const isActive = i === activeStep;
            const isPast = i < activeStep;
            const pct = (i / (STEP_COUNT - 1)) * 100;

            return (
              <motion.button
                key={i}
                onClick={() => handleDotClick(i)}
                className="absolute top-1/2 z-10 flex h-[44px] w-[44px] -translate-x-1/2 -translate-y-1/2 cursor-pointer items-center justify-center"
                style={{ left: `${pct}%` }}
                aria-label={`Step ${i + 1}: ${HOW_IT_WORKS[i].title}`}
              >
                {/* Pulse ring (active only) */}
                {isActive && (
                  <motion.span
                    className="absolute rounded-full border border-purple"
                    initial={{ width: 10, height: 10, opacity: 0.6 }}
                    animate={{
                      width: [10, 28],
                      height: [10, 28],
                      opacity: [0.6, 0],
                    }}
                    transition={{
                      duration: 1.5,
                      repeat: Infinity,
                      ease: "easeOut",
                    }}
                  />
                )}

                {/* Glow backdrop (active only) */}
                {isActive && (
                  <span className="absolute h-[24px] w-[24px] rounded-full bg-purple/30 blur-[6px]" />
                )}

                {/* Core dot */}
                <motion.span
                  className="relative rounded-full"
                  animate={{
                    width: isActive ? 14 : 10,
                    height: isActive ? 14 : 10,
                    backgroundColor: isActive || isPast
                      ? "var(--color-purple)"
                      : "var(--color-border-light)",
                    boxShadow: isActive
                      ? "0 0 12px rgba(124, 107, 240, 0.5)"
                      : "none",
                  }}
                  transition={{ duration: 0.3 }}
                />
              </motion.button>
            );
          })}
        </div>

        {/* ── Mobile Rail (simplified) ── */}
        <div className="relative mb-8 block h-[2px] md:hidden">
          <div className="absolute inset-0 rounded-full bg-border" />
          <motion.div
            className="timeline-progress absolute inset-y-0 left-0 rounded-full bg-purple"
            animate={{ width: progressWidth }}
            transition={{ duration: 0.5, ease: [0.32, 0.72, 0, 1] }}
          />
          {HOW_IT_WORKS.map((_, i) => {
            const isActive = i === activeStep;
            const isPast = i < activeStep;
            const pct = (i / (STEP_COUNT - 1)) * 100;
            return (
              <motion.button
                key={i}
                onClick={() => handleDotClick(i)}
                className="absolute top-1/2 z-10 flex h-[44px] w-[44px] -translate-x-1/2 -translate-y-1/2 cursor-pointer items-center justify-center"
                style={{ left: `${pct}%` }}
                aria-label={`Step ${i + 1}: ${HOW_IT_WORKS[i].title}`}
              >
                {isActive && (
                  <motion.span
                    className="absolute rounded-full border border-purple"
                    initial={{ width: 10, height: 10, opacity: 0.6 }}
                    animate={{
                      width: [10, 24],
                      height: [10, 24],
                      opacity: [0.6, 0],
                    }}
                    transition={{
                      duration: 1.5,
                      repeat: Infinity,
                      ease: "easeOut",
                    }}
                  />
                )}
                <motion.span
                  className="relative rounded-full"
                  animate={{
                    width: isActive ? 12 : 8,
                    height: isActive ? 12 : 8,
                    backgroundColor: isActive || isPast
                      ? "var(--color-purple)"
                      : "var(--color-border-light)",
                    boxShadow: isActive
                      ? "0 0 10px rgba(124, 107, 240, 0.5)"
                      : "none",
                  }}
                  transition={{ duration: 0.3 }}
                />
              </motion.button>
            );
          })}
        </div>

        {/* ── Step Content ── */}
        {/* Desktop: 3-column grid */}
        <div className="hidden gap-6 md:grid md:grid-cols-3">
          {HOW_IT_WORKS.map((item, i) => {
            const isActive = i === activeStep;
            return (
              <motion.div
                key={item.step}
                className="text-center"
                animate={{
                  opacity: isActive ? 1 : 0.35,
                  y: isActive ? 0 : 4,
                }}
                transition={{ duration: 0.4, ease: "easeOut" }}
              >
                <span className="type-caption mb-2 block font-mono font-semibold tracking-widest text-purple-light uppercase">
                  Step 0{item.step}
                </span>
                <h3 className="type-h5 mb-2 font-semibold text-white">
                  {item.title}
                </h3>
                <p className="type-body-sm leading-[1.7] text-text-muted">
                  {item.description}
                </p>
              </motion.div>
            );
          })}
        </div>

        {/* Mobile: stacked, all visible, active has purple accent */}
        <div className="flex flex-col gap-6 md:hidden">
          {HOW_IT_WORKS.map((item, i) => {
            const isActive = i === activeStep;
            return (
              <motion.div
                key={item.step}
                className={`rounded-lg border-l-2 pl-4 transition-colors ${
                  isActive
                    ? "border-purple"
                    : "border-transparent"
                }`}
                animate={{
                  opacity: isActive ? 1 : 0.5,
                }}
                transition={{ duration: 0.3 }}
              >
                <span className="type-caption mb-1 block font-mono font-semibold tracking-widest text-purple-light uppercase">
                  Step 0{item.step}
                </span>
                <h3 className="type-body mb-1 font-semibold text-white">
                  {item.title}
                </h3>
                <p className="type-body-sm leading-[1.6] text-text-muted">
                  {item.description}
                </p>
              </motion.div>
            );
          })}
        </div>
      </div>
    </Section>
  );
}
