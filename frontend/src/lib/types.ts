// Matches internal/api/handlers.go
export interface ConnectionInfo {
  hostname: string
  port: number
  ssh_command: string
}

export interface InstanceResponse {
  id: string
  name?: string
  status: 'starting' | 'running' | 'stopping' | 'terminated' | 'error'
  gpu_type: string
  gpu_count: number
  tier: 'spot' | 'on_demand'
  region: string
  price_per_hour: number
  connection?: ConnectionInfo
  error_reason?: string
  created_at: string
  ready_at?: string
  terminated_at?: string
}

export interface CreateInstanceRequest {
  gpu_type: string
  gpu_count: number
  region: string
  tier: 'spot' | 'on_demand'
  ssh_key_ids?: string[]
}

// Matches internal/availability/types.go
export interface AvailableOffering {
  gpu_model: string
  vram_gb: number
  cpu_cores: number
  ram_gb: number
  storage_gb: number
  price_per_hour: number
  region: string
  tier: 'spot' | 'on_demand'
  available_count: number
  avg_uptime_pct: number
}

// Matches internal/api/handlers_ssh_keys.go
export interface SSHKeyResponse {
  id: string
  name: string
  fingerprint: string
  created_at: string
}

// Matches internal/api/handlers_billing.go
export interface BillingSessionResponse {
  id: string
  instance_id: string
  gpu_type: string
  gpu_count: number
  price_per_hour: number
  started_at: string
  ended_at?: string
  duration_seconds?: number
  total_cost?: number
  estimated_cost?: number
  is_active: boolean
}

export interface UsageResponse {
  sessions: BillingSessionResponse[]
  total_cost: number
  currency: string
}

export interface PaginatedResponse<T> {
  next_cursor?: string
  items: T[]
}
