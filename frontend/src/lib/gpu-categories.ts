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
  const lower = gpuModel.toLowerCase();
  for (const cat of GPU_CATEGORIES) {
    if (cat.models.some((m) => lower.includes(m))) return cat.label;
  }
  return "Other";
}
