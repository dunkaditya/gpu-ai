// ── Navigation ──

export const NAV_LINKS = [
  { label: "Products", href: "#" },
  { label: "Solutions", href: "#" },
  { label: "Docs", href: "#" },
  { label: "Pricing", href: "#" },
  { label: "Enterprise", href: "#" },
] as const;

// ── Hero ──

export const HERO_CONTENT = {
  headline: "Your Infrastructure for GPU Compute",
  subtitle:
    "GPU.ai aggregates inventory from every major GPU cloud so you get the best price, the fastest deploy, and a single API for all of it.",
  primaryCTA: "Launch GPU Instance",
  secondaryCTA: "Talk to our team",
  metrics: [
    { value: "12+", label: "Providers" },
    { value: "30%", label: "Cheaper" },
    { value: "< 60s", label: "Deploy Time" },
    { value: "99.9%", label: "Uptime" },
  ],
} as const;

// ── Use Case Tabs ──

export const USE_CASE_TABS = [
  {
    id: "ml-training",
    label: "ML Training",
    title: "Train models on the cheapest available GPUs",
    description:
      "Distribute training jobs across providers automatically. GPU.ai finds the lowest-cost H100s and A100s in real time so your training runs finish faster and cheaper.",
    features: [
      "Multi-node distributed training across providers",
      "Automatic spot instance failover",
      "Real-time cost optimization per epoch",
      "Pre-configured PyTorch and JAX environments",
    ],
  },
  {
    id: "inference",
    label: "Inference",
    title: "Deploy inference endpoints with zero cold starts",
    description:
      "Serve models at scale with per-second billing and automatic load balancing. GPU.ai routes traffic to the cheapest available GPUs across all providers.",
    features: [
      "Auto-scaling from zero to thousands of GPUs",
      "Per-second billing with no minimum commitment",
      "Global edge routing for low latency",
      "Support for vLLM, TGI, and custom runtimes",
    ],
  },
  {
    id: "fine-tuning",
    label: "Fine-tuning",
    title: "Fine-tune foundation models without the infrastructure headache",
    description:
      "Launch fine-tuning jobs on optimal hardware with a single command. GPU.ai handles provider selection, data transfer, and checkpoint management.",
    features: [
      "LoRA and full fine-tuning support",
      "Automatic checkpoint saving to your cloud storage",
      "Cost-optimized GPU selection per model size",
      "Private WireGuard networking for data security",
    ],
  },
  {
    id: "rendering",
    label: "Rendering",
    title: "Burst rendering capacity on demand",
    description:
      "Scale to hundreds of GPUs for rendering workloads and release them when done. Pay only for the seconds you use with no reserved capacity required.",
    features: [
      "Burst to 100+ GPUs in under 60 seconds",
      "Blender, Unreal, and custom pipeline support",
      "Automatic job distribution and frame assembly",
      "Spot pricing for non-urgent render queues",
    ],
  },
  {
    id: "research",
    label: "Research",
    title: "Run experiments without managing infrastructure",
    description:
      "Focus on research, not cloud consoles. GPU.ai gives you SSH access to any GPU type across every provider through one CLI and one bill.",
    features: [
      "SSH into any GPU in under 60 seconds",
      "Jupyter notebook environments pre-installed",
      "Spend limits and alerts per project",
      "One invoice across all providers",
    ],
  },
] as const;

// ── Feature Pillars ──

export const FEATURE_PILLARS = [
  {
    id: "source",
    title: "Source",
    subtitle: "The best GPU at the best price",
    description:
      "GPU.ai continuously scans 12+ cloud providers to find available GPUs at the lowest price. You specify what you need. We find where to get it.",
    features: [
      "12+ providers aggregated in real time",
      "Automatic best-price selection",
      "Spot and on-demand inventory",
      "H100, A100, L40S, RTX 4090, and more",
    ],
  },
  {
    id: "deploy",
    title: "Deploy",
    subtitle: "From zero to GPU in 60 seconds",
    description:
      "One command provisions your instance, configures networking, and gives you SSH access. Every instance gets a private WireGuard tunnel for security.",
    features: [
      "60-second launch time",
      "Private WireGuard networking",
      "SSH ready on boot",
      "Pre-configured ML environments",
    ],
  },
  {
    id: "scale",
    title: "Scale",
    subtitle: "Elastic GPU capacity, per-second billing",
    description:
      "Scale from one GPU to hundreds across providers. Set spending limits, use spot instances for cost savings, and pay only for the seconds you use.",
    features: [
      "Per-second billing",
      "Spot instance support",
      "Spending limits and alerts",
      "Multi-provider failover",
    ],
  },
] as const;

// ── CLI Demo ──

export const CLI_COMMANDS = [
  {
    prompt: "$",
    command: "gpuctl launch --gpu h100 --count 4 --region us-east",
    output: [
      "Searching 12 providers for 4x H100 in us-east...",
      "Best price: $2.49/hr/gpu via provider-7",
      "Provisioning instance gpu-abc123...",
      "WireGuard tunnel established",
      "Instance gpu-abc123 ready (52s)",
    ],
  },
  {
    prompt: "$",
    command: "gpuctl status gpu-abc123",
    output: [
      "Instance:  gpu-abc123",
      "Status:    running",
      "GPUs:      4x NVIDIA H100 80GB",
      "Region:    us-east",
      "Cost:      $9.96/hr ($2.49/gpu)",
      "Uptime:    3h 24m",
    ],
  },
  {
    prompt: "$",
    command: "gpuctl ssh gpu-abc123",
    output: [
      "Connecting via WireGuard tunnel...",
      "Welcome to gpu-abc123 (Ubuntu 22.04)",
      "4x NVIDIA H100 80GB | CUDA 12.4 | PyTorch 2.3",
      "root@gpu-abc123:~#",
    ],
  },
] as const;

// ── Footer ──

export const FOOTER_COLUMNS = [
  {
    title: "Product",
    links: [
      { label: "Features", href: "#" },
      { label: "Pricing", href: "#" },
      { label: "Docs", href: "#" },
      { label: "CLI", href: "#" },
      { label: "API", href: "#" },
    ],
  },
  {
    title: "Company",
    links: [
      { label: "About", href: "#" },
      { label: "Blog", href: "#" },
      { label: "Careers", href: "#" },
      { label: "Contact", href: "#" },
    ],
  },
  {
    title: "Resources",
    links: [
      { label: "Documentation", href: "#" },
      { label: "API Reference", href: "#" },
      { label: "Status", href: "#" },
      { label: "Changelog", href: "#" },
    ],
  },
  {
    title: "Legal",
    links: [
      { label: "Privacy", href: "#" },
      { label: "Terms", href: "#" },
      { label: "Security", href: "#" },
    ],
  },
] as const;
