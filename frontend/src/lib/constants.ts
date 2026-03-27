// ── Navigation ──

export const NAV_LINKS = [
  { label: "Features", href: "/#features" },
  { label: "Pricing", href: "/#pricing" },
] as const;

export const PRODUCTS_LINKS = [
  { label: "On-Demand", href: "/products/on-demand" },
  { label: "Buildouts", href: "/products/buildouts" },
] as const;

export const COMPANY_LINKS = [
  { label: "About", href: "/about" },
  { label: "Careers", href: "/careers" },
] as const;

// ── Hero Stats ──

export const HERO_STATS = [
  { value: 10, suffix: "+", label: "GPU Models" },
  { value: 30, suffix: "%", label: "Avg Savings" },
  { value: 60, prefix: "<", suffix: "s", label: "Deploy Time" },
] as const;

// ── Features ──

export const FEATURES = [
  {
    icon: "globe",
    title: "Real-Time Price Comparison",
    description:
      "GPU.ai continuously compares prices across cloud providers so you always get the lowest rate. No need to check five dashboards — we surface the best deal automatically.",
  },
  {
    icon: "zap",
    title: "Deploy in Under 60 Seconds",
    description:
      "A single CLI command provisions your instance, configures networking, and gives you SSH access. No cloud consoles, no Terraform, no waiting.",
  },
  {
    icon: "shield",
    title: "Per-Second Billing, No Minimums",
    description:
      "Pay only for what you use, billed by the second. No reserved contracts, no hourly rounding, no minimum commitments. Shut down and stop paying instantly.",
  },
] as const;

// ── How It Works ──

export const HOW_IT_WORKS = [
  {
    step: 1,
    title: "Pick your GPU",
    description:
      "Select a model and quantity — we find the lowest price across all providers.",
  },
  {
    step: 2,
    title: "Deploy instantly",
    description:
      "One command gives you SSH access to a ready ML environment in under 60 seconds.",
  },
  {
    step: 3,
    title: "Pay per second",
    description:
      "No contracts or minimums. Billed per second, cheapest provider auto-selected.",
  },
] as const;

// ── Code Example ──

export const CODE_EXAMPLE = [
  'from gpuai import GPU',
  '',
  '# Launch 4x H100 at the best available price',
  'instance = GPU.launch(',
  '    gpu="h100",',
  '    count=4,',
  '    region="us-east",',
  '    image="pytorch:2.3-cuda12.4"',
  ')',
  '',
  'print(f"Instance {instance.id} ready")',
  'print(f"SSH: ssh root@{instance.ip}")',
  "print(f\"Cost: \\${instance.price_hr}/hr per GPU\")",
  '',
  '# Run your training job',
  'instance.exec("torchrun --nproc_per_node=4 train.py")',
].join("\n");

// ── Pricing Table ──

import type { PricingComparisonResponse } from "./types";

export const COMPETITOR_NAMES = ["Lambda", "CoreWeave", "AWS"] as const;

// Fallback pricing data — shown immediately while API loads.
// Matches the structure of PricingComparisonResponse for SWR fallback.
export const PRICING_FALLBACK: PricingComparisonResponse = {
  featured_models: ["h200_sxm", "h100_sxm", "b200", "a100_80gb"],
  competitor_names: ["Lambda", "CoreWeave", "AWS"],
  updated_at: "",
  gpus: [
    {
      gpu_model: "h200_sxm",
      display_name: "H200 SXM",
      vram_gb: 141,
      gpuai_price: 3.81,
      available_count: 48,
      competitors: [
        { name: "Lambda", price: 4.99 },
        { name: "CoreWeave", price: 6.31 },
        { name: "AWS", price: 4.97 },
      ],
      savings_pct: 28,
    },
    {
      gpu_model: "h100_sxm",
      display_name: "H100 SXM",
      vram_gb: 80,
      gpuai_price: 2.64,
      available_count: 124,
      competitors: [
        { name: "Lambda", price: 3.29 },
        { name: "CoreWeave", price: 6.16 },
        { name: "AWS", price: 3.93 },
      ],
      savings_pct: 24,
    },
    {
      gpu_model: "b200",
      display_name: "B200",
      vram_gb: 192,
      gpuai_price: 5.29,
      available_count: 16,
      competitors: [
        { name: "Lambda", price: 6.08 },
        { name: "CoreWeave", price: 8.60 },
        { name: "AWS", price: 14.24 },
      ],
      savings_pct: 18,
    },
    {
      gpu_model: "a100_80gb",
      display_name: "A100",
      vram_gb: 80,
      gpuai_price: 2.00,
      available_count: 256,
      competitors: [
        { name: "Lambda", price: 2.06 },
        { name: "CoreWeave", price: 2.70 },
        { name: "AWS", price: 5.12 },
      ],
      savings_pct: 8,
    },
  ],
};

// Legacy flat format for any components still using it.
export const PRICING_DATA = PRICING_FALLBACK.gpus.map((g) => ({
  gpu: g.display_name,
  vram: `${g.vram_gb} GB`,
  gpuai: g.gpuai_price ?? 0,
  competitors: g.competitors
    .filter((c): c is { name: string; price: number } => c.price !== null)
    .map((c) => ({ name: c.name, price: c.price })),
  savings: g.savings_pct ?? 0,
}));

// ── Footer Links ──

export const FOOTER_LINKS = {
  Product: [
    { label: "Features", href: "/#features" },
    { label: "On-Demand GPUs", href: "/products/on-demand" },
    { label: "Custom Buildouts", href: "/products/buildouts" },
    { label: "Pricing", href: "/#pricing" },
  ],
  Company: [
    { label: "About", href: "/about" },
    { label: "Careers", href: "/careers" },
  ],
  Resources: [
    { label: "Documentation", href: "/docs/ssh-keys" },
    { label: "Changelog", href: "/changelog" },
  ],
  Legal: [
    { label: "Privacy", href: "/privacy" },
    { label: "Terms", href: "/terms" },
  ],
} as const;
