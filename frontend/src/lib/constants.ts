// ── Navigation ──

export const NAV_LINKS = [
  { label: "Features", href: "#features" },
  { label: "Pricing", href: "#pricing" },
  { label: "Docs", href: "/docs" },
  { label: "Enterprise", href: "/enterprise" },
] as const;

// ── Hero Stats ──

export const HERO_STATS = [
  { value: 12, suffix: "+", label: "GPU Providers" },
  { value: 30, suffix: "%", label: "Avg Savings" },
  { value: 60, prefix: "<", suffix: "s", label: "Deploy Time" },
  { value: 99, suffix: ".9%", label: "Uptime SLA" },
] as const;

// ── Features ──

export const FEATURES = [
  {
    icon: "globe",
    title: "12+ Providers, One API",
    description:
      "GPU.ai aggregates inventory from RunPod, Lambda, CoreWeave, and more. You pick the GPU — we find the cheapest available instance in real time.",
  },
  {
    icon: "zap",
    title: "Deploy in Under 60 Seconds",
    description:
      "A single CLI command provisions your instance, configures networking, and gives you SSH access. No cloud consoles, no Terraform, no waiting.",
  },
  {
    icon: "shield",
    title: "Private WireGuard Networking",
    description:
      "Every instance gets a private WireGuard tunnel out of the box. Your data stays encrypted in transit between your machines and your GPUs.",
  },
] as const;

// ── How It Works ──

export const HOW_IT_WORKS = [
  {
    step: 1,
    title: "Pick your GPU",
    description:
      "Choose the model, VRAM, and quantity you need. GPU.ai scans every provider to find the lowest available price across on-demand and spot inventory.",
  },
  {
    step: 2,
    title: "Deploy instantly",
    description:
      "One command spins up your instance with a private WireGuard tunnel, SSH access, and a pre-configured ML environment — all in under 60 seconds.",
  },
  {
    step: 3,
    title: "Pay per second",
    description:
      "No reserved contracts or minimums. You're billed per second of actual usage, and GPU.ai automatically selects the cheapest provider for your workload.",
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

export const COMPETITOR_NAMES = ["AWS", "GCP", "Lambda"] as const;

export const PRICING_DATA = [
  {
    gpu: "H100 SXM",
    vram: "80 GB",
    gpuai: 2.49,
    competitors: [
      { name: "AWS", price: 4.15 },
      { name: "GCP", price: 3.98 },
      { name: "Lambda", price: 2.99 },
    ],
    savings: 40,
  },
  {
    gpu: "A100 SXM",
    vram: "80 GB",
    gpuai: 1.29,
    competitors: [
      { name: "AWS", price: 2.48 },
      { name: "GCP", price: 2.21 },
      { name: "Lambda", price: 1.69 },
    ],
    savings: 42,
  },
  {
    gpu: "L40S",
    vram: "48 GB",
    gpuai: 0.89,
    competitors: [
      { name: "AWS", price: 1.52 },
      { name: "GCP", price: 1.40 },
      { name: "Lambda", price: 1.10 },
    ],
    savings: 41,
  },
  {
    gpu: "RTX 4090",
    vram: "24 GB",
    gpuai: 0.44,
    competitors: [
      { name: "AWS", price: 0.89 },
      { name: "GCP", price: 0.79 },
      { name: "Lambda", price: 0.59 },
    ],
    savings: 51,
  },
  {
    gpu: "A10G",
    vram: "24 GB",
    gpuai: 0.39,
    competitors: [
      { name: "AWS", price: 0.75 },
      { name: "GCP", price: 0.65 },
      { name: "Lambda", price: 0.55 },
    ],
    savings: 48,
  },
] as const;

// ── Footer Links ──

export const FOOTER_LINKS = {
  Product: [
    { label: "Features", href: "#features" },
    { label: "Pricing", href: "#pricing" },
    { label: "CLI", href: "/docs/cli" },
    { label: "API", href: "/docs/api" },
  ],
  Company: [
    { label: "About", href: "/about" },
    { label: "Blog", href: "/blog" },
    { label: "Careers", href: "/careers" },
    { label: "Contact", href: "/contact" },
  ],
  Resources: [
    { label: "Documentation", href: "/docs" },
    { label: "Status", href: "/status" },
    { label: "Changelog", href: "/changelog" },
  ],
  Legal: [
    { label: "Privacy", href: "/privacy" },
    { label: "Terms", href: "/terms" },
    { label: "Security", href: "/security" },
  ],
} as const;
