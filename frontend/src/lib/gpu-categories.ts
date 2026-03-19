export interface GPUCategoryDef {
  label: string;
  description: string;
  models: string[]; // lowercase identifiers to match against gpu_model
}

export const GPU_CATEGORIES: GPUCategoryDef[] = [
  {
    label: "Blackwell",
    description: "Latest generation",
    models: [
      "b200",
      "b300",
      "rtx pro 6000",
      "rtx pro 4500",
      "rtx 5090",
      "rtx 5080",
    ],
  },
  {
    label: "Hopper",
    description: "Data center H-series",
    models: ["h200", "h100"],
  },
  {
    label: "Ada Lovelace",
    description: "L-series & RTX 40-series",
    models: [
      "l40s",
      "l40",
      "l4",
      "rtx 6000 ada",
      "rtx 5000 ada",
      "rtx 4000 ada",
      "rtx 2000 ada",
      "rtx 4090",
      "rtx 4080",
    ],
  },
  {
    label: "Ampere",
    description: "A-series & RTX 30-series",
    models: [
      "a100",
      "a40",
      "a30",
      "a10",
      "rtx a6000",
      "rtx a5000",
      "rtx a4500",
      "rtx a4000",
      "rtx 3090",
      "rtx 3080",
    ],
  },
  {
    label: "Legacy",
    description: "Previous generation",
    models: ["v100"],
  },
];

export function classifyGPU(gpuModel: string): string {
  const lower = gpuModel.toLowerCase().replace(/_/g, " ");
  for (const cat of GPU_CATEGORIES) {
    if (cat.models.some((m) => lower.includes(m))) return cat.label;
  }
  return "Other";
}

export const GPU_DISPLAY_NAMES: Record<string, string> = {
  // Blackwell
  "b200": "B200",
  "b300": "B300",
  "rtx pro 6000": "RTX PRO 6000",
  "rtx pro 4500": "RTX PRO 4500",
  "rtx 5090": "RTX 5090",
  "rtx 5080": "RTX 5080",
  // Hopper (exact canonical names from API)
  "h200 sxm": "H200 SXM",
  "h200 nvl": "H200 NVL",
  "h200": "H200 SXM",
  "h100 sxm": "H100 SXM",
  "h100 nvl": "H100 NVL",
  "h100 pcie": "H100 PCIe",
  "h100": "H100 SXM",
  // Ada Lovelace
  "l40s": "L40S",
  "l40": "L40",
  "l4": "L4",
  "rtx 6000 ada": "RTX 6000 Ada",
  "rtx 5000 ada": "RTX 5000 Ada",
  "rtx 4000 ada": "RTX 4000 Ada",
  "rtx 2000 ada": "RTX 2000 Ada",
  "rtx 4090": "RTX 4090",
  "rtx 4080": "RTX 4080",
  // Ampere (exact canonical names from API)
  "a100 80gb": "A100",
  "a100 40gb": "A100 40GB",
  "a100": "A100",
  "a40": "A40",
  "a30": "A30",
  "a10": "A10",
  "rtx a6000": "RTX A6000",
  "rtx a5000": "RTX A5000",
  "rtx a4500": "RTX A4500",
  "rtx a4000": "RTX A4000",
  "rtx 3090": "RTX 3090",
  "rtx 3080": "RTX 3080",
  // Legacy
  "v100": "V100",
};

export function getDisplayName(gpuModel: string): string {
  const normalized = gpuModel.toLowerCase().replace(/_/g, " ");
  const name = GPU_DISPLAY_NAMES[normalized] ?? normalized;
  return name.toUpperCase();
}
