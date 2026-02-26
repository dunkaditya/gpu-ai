"use client";

import { useState, useEffect } from "react";

interface PriceBarProps {
  name: string;
  price: number;
  maxP: number;
  isSelf: boolean;
  delay: number;
}

function PriceBar({ name, price, maxP, isSelf, delay }: PriceBarProps) {
  const [w, setW] = useState(0);

  useEffect(() => {
    const t = setTimeout(() => setW((price / maxP) * 100), 300 + delay);
    return () => clearTimeout(t);
  }, [price, maxP, delay]);

  return (
    <div className="mb-2.5">
      <div className="mb-1 flex justify-between type-ui-xs">
        <span className={isSelf ? "font-semibold text-green" : "text-text-muted"}>
          {name}
        </span>
        <span
          className={`type-ui-xs font-medium ${
            isSelf ? "text-green" : "text-text-dim"
          }`}
        >
          ${price.toFixed(2)}
        </span>
      </div>
      <div className="h-[5px] overflow-hidden rounded-[3px] bg-border/70">
        <div
          className={`h-full rounded-[3px] transition-[width] duration-[2700ms] ${
            isSelf ? "bg-green" : "bg-border-light"
          }`}
          style={{
            width: `${w}%`,
            transitionTimingFunction: "cubic-bezier(0.16, 1, 0.3, 1)",
          }}
        />
      </div>
    </div>
  );
}

const GPUS = [
  { name: "H200 SXM", vram: "141 GB", price: 1.89, avail: 48 },
  { name: "H100 SXM", vram: "80 GB", price: 1.29, avail: 124 },
  { name: "B200 SXM", vram: "192 GB", price: 3.19, avail: 16 },
  { name: "A100 SXM", vram: "80 GB", price: 0.79, avail: 256 },
] as const;

const COMPS: Record<string, Record<string, number>> = {
  "H200 SXM": { Lambda: 3.99, RunPod: 3.49, AWS: 6.98 },
  "H100 SXM": { Lambda: 2.99, RunPod: 2.49, AWS: 3.9 },
  "B200 SXM": { Lambda: 5.74, CoreWeave: 5.5 },
  "A100 SXM": { Lambda: 1.48, RunPod: 1.64, AWS: 3.4 },
};

export function PricingWidget() {
  const [selGpu, setSelGpu] = useState(0);

  const gpu = GPUS[selGpu];
  const comps = COMPS[gpu.name];
  const compArr = Object.entries(comps);
  const maxP = Math.max(gpu.price, ...compArr.map(([, v]) => v));
  const savePct = Math.round(
    (1 - gpu.price / (compArr.reduce((a, [, v]) => a + v, 0) / compArr.length)) * 100,
  );

  return (
    <div className="relative w-full max-w-[400px]">
      {/* Gradient border */}
      <div
        className="absolute -inset-[1px] rounded-2xl"
        style={{
          background: "linear-gradient(135deg, rgba(124, 107, 240, 0.4) 0%, rgba(124, 107, 240, 0.15) 50%, rgba(124, 107, 240, 0.4) 100%)",
        }}
      />
      <div className="relative overflow-hidden rounded-2xl border border-transparent bg-bg-card/90 backdrop-blur-2xl">
        {/* GPU Tabs */}
        <div className="flex border-b border-border">
          {GPUS.map((g, i) => (
            <button
              key={g.name}
              onClick={() => setSelGpu(i)}
              className={`flex-1 border-b-2 px-2 py-3 type-ui-2xs transition-all ${
                selGpu === i
                  ? "border-purple bg-border/50 font-semibold text-text"
                  : "border-transparent text-text-dim hover:bg-border/30"
              }`}
            >
              {g.name.split(" ")[0]}
            </button>
          ))}
        </div>

        {/* Content */}
        <div className="px-6 py-5">
          <div className="mb-1.5 flex items-baseline justify-between">
            <span className="text-[16px] font-semibold text-white">{gpu.name}</span>
            <span className="type-ui-2xs text-text-dim">{gpu.vram}</span>
          </div>

          <div className="mb-5 flex items-baseline gap-0.5">
            <span className="text-[32px] font-extrabold tracking-tight text-green">
              ${gpu.price.toFixed(2)}
            </span>
            <span className="type-ui-sm text-text-dim">/gpu/hr</span>
            <span className="ml-auto inline-flex items-center rounded-[5px] border border-green/20 bg-green-dim px-2 py-0.5 type-ui-2xs font-semibold text-green">
              -{savePct}%
            </span>
          </div>

          <div className="mb-2.5 text-[10px] font-semibold uppercase tracking-[1px] text-text-dim">
            vs. Competitors
          </div>

          <PriceBar name="GPU.ai" price={gpu.price} maxP={maxP} isSelf delay={0} />
          {compArr.map(([name, p], i) => (
            <PriceBar
              key={name}
              name={name}
              price={p}
              maxP={maxP}
              isSelf={false}
              delay={(i + 1) * 60}
            />
          ))}

          <div className="mt-5 border-t border-border pt-3.5">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-1.5">
                <div
                  className={`h-[7px] w-[7px] rounded-full ${
                    gpu.avail > 30
                      ? "bg-green shadow-[0_0_8px_rgba(34,197,94,0.27)]"
                      : "bg-amber-400 shadow-[0_0_8px_rgba(251,191,36,0.27)]"
                  }`}
                />
                <span className="type-ui-xs text-text-muted">{gpu.avail} available</span>
              </div>
              <span className="type-ui-2xs text-text-dim">Spot · India</span>
            </div>
            <a
              href="/sign-up"
              className="gradient-btn mt-3.5 flex w-full items-center justify-center gap-2 rounded-lg py-3 text-[15px] font-semibold text-white transition-all"
            >
              LAUNCH INSTANCE →
            </a>
          </div>
        </div>
      </div>
    </div>
  );
}
