export interface MockInstance {
  id: string;
  name?: string;
  status: "starting" | "running" | "stopping" | "terminated" | "error";
  gpu_type: string;
  gpu_count: number;
  tier: "spot" | "on_demand";
  region: string;
  price_per_hour: number;
  connection?: { hostname: string; port: number; ssh_command: string };
  error_reason?: string;
  created_at: string;
  ready_at?: string;
  terminated_at?: string;
}

export const MOCK_INSTANCES: MockInstance[] = [
  {
    id: "inst_h100_abc1",
    name: "training-run-47",
    status: "running",
    gpu_type: "H100 SXM",
    gpu_count: 4,
    tier: "spot",
    region: "us-east",
    price_per_hour: 8.49,
    connection: {
      hostname: "h100-abc1.gpu.ai",
      port: 10042,
      ssh_command: "ssh root@h100-abc1.gpu.ai -p 10042",
    },
    created_at: "2026-03-01T14:30:00Z",
    ready_at: "2026-03-01T14:32:15Z",
  },
  {
    id: "inst_a100_def2",
    name: "inference-prod",
    status: "running",
    gpu_type: "A100 SXM",
    gpu_count: 2,
    tier: "on_demand",
    region: "eu-west",
    price_per_hour: 5.12,
    connection: {
      hostname: "a100-def2.gpu.ai",
      port: 10087,
      ssh_command: "ssh root@a100-def2.gpu.ai -p 10087",
    },
    created_at: "2026-02-28T09:15:00Z",
    ready_at: "2026-02-28T09:17:42Z",
  },
  {
    id: "inst_4090_ghi3",
    status: "starting",
    gpu_type: "RTX 4090",
    gpu_count: 1,
    tier: "spot",
    region: "us-east",
    price_per_hour: 0.74,
    created_at: "2026-03-02T21:00:00Z",
  },
  {
    id: "inst_l40s_jkl4",
    name: "finetune-llama",
    status: "error",
    gpu_type: "L40S",
    gpu_count: 1,
    tier: "on_demand",
    region: "ap-south",
    price_per_hour: 1.89,
    error_reason: "Provider capacity exhausted",
    created_at: "2026-03-02T18:45:00Z",
  },
  {
    id: "inst_a10g_mno5",
    name: "batch-embeddings",
    status: "terminated",
    gpu_type: "A10G",
    gpu_count: 2,
    tier: "spot",
    region: "us-west",
    price_per_hour: 1.21,
    created_at: "2026-02-27T06:00:00Z",
    ready_at: "2026-02-27T06:02:30Z",
    terminated_at: "2026-02-28T18:00:00Z",
  },
];
