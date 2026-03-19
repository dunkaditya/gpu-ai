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

// GPU card data grouped by model (client-side aggregation)
export interface GPUCardData {
  gpu_model: string
  vram_gb: number
  cpu_cores: number
  ram_gb: number
  storage_gb: number
  regions: string[]
  spot_price?: number
  on_demand_price?: number
  total_available: number
  offerings: AvailableOffering[]
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

export interface SpendingLimitResponse {
  monthly_limit_cents: number
  monthly_limit_dollars: number
  current_month_spend_cents: number
  current_month_spend_dollars: number
  percent_used: number
  billing_cycle_start: string
  limit_reached_at?: string
}

export interface PaginatedResponse<T> {
  next_cursor?: string
  items: T[]
}

// Matches internal/pricing/types.go
export interface CompetitorEntry {
  name: string
  price: number | null
}

export interface GPUComparison {
  gpu_model: string
  display_name: string
  vram_gb: number
  gpuai_price: number | null
  available_count: number
  competitors: CompetitorEntry[]
  savings_pct: number | null
}

export interface PricingComparisonResponse {
  featured_models: string[]
  gpus: GPUComparison[]
  competitor_names: string[]
  updated_at: string
}

// Credit Balance types
export interface BalanceResponse {
  balance_cents: number
  balance_dollars: number
  auto_pay_enabled: boolean
  auto_pay_threshold_cents: number
  auto_pay_amount_cents: number
}

export interface TransactionResponse {
  id: string
  type: 'credit_purchase' | 'auto_pay' | 'credit_code' | 'usage_deduction' | 'adjustment'
  amount_cents: number
  amount_dollars: number
  balance_after_cents: number
  description: string
  reference_id?: string
  created_at: string
}

export interface TransactionsListResponse {
  transactions: TransactionResponse[]
  has_more: boolean
  next_cursor?: string
}

export interface PurchaseCreditsResponse {
  checkout_url: string
  session_id: string
}

export interface RedeemCodeResponse {
  amount_cents: number
  amount_dollars: number
  new_balance_cents: number
}
