import type {
  InstanceResponse,
  AvailableOffering,
  SSHKeyResponse,
  UsageResponse,
  CreateInstanceRequest,
  SpendingLimitResponse,
  PricingComparisonResponse,
  BalanceResponse,
  PurchaseCreditsResponse,
  RedeemCodeResponse,
  TransactionsListResponse,
} from './types'

const API_BASE = '/api/v1'

// Generic fetcher for SWR
export const fetcher = (url: string) =>
  fetch(url).then((r) => {
    if (!r.ok) throw new Error(`API error: ${r.status}`)
    return r.json()
  })

// Graceful fetcher — returns null on auth/server errors instead of throwing.
// Allows dashboard pages to render empty states when backend is unavailable.
export const gracefulFetcher = (url: string) =>
  fetch(url).then((r) => {
    if (r.status === 401 || r.status === 403) return null
    if (!r.ok) throw new Error(`API error: ${r.status}`)
    return r.json()
  })

// Instances
export async function fetchInstances(): Promise<{ instances: InstanceResponse[] }> {
  const res = await fetch(`${API_BASE}/instances`)
  if (!res.ok) throw new Error('Failed to fetch instances')
  return res.json()
}

export async function createInstance(req: CreateInstanceRequest): Promise<InstanceResponse> {
  const res = await fetch(`${API_BASE}/instances`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.detail || 'Failed to create instance')
  }
  return res.json()
}

export async function renameInstance(id: string, name: string): Promise<InstanceResponse> {
  const res = await fetch(`${API_BASE}/instances/${id}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.detail || 'Failed to rename instance')
  }
  return res.json()
}

export async function terminateInstance(id: string): Promise<void> {
  const res = await fetch(`${API_BASE}/instances/${id}`, { method: 'DELETE' })
  if (!res.ok && res.status !== 204) throw new Error('Failed to terminate instance')
}

// GPU Availability
export async function fetchGPUAvailability(filters?: {
  gpu_model?: string
  region?: string
  tier?: string
}): Promise<{ available: AvailableOffering[] }> {
  const params = new URLSearchParams()
  if (filters?.gpu_model) params.set('gpu_model', filters.gpu_model)
  if (filters?.region) params.set('region', filters.region)
  if (filters?.tier) params.set('tier', filters.tier)
  const qs = params.toString()
  const res = await fetch(`${API_BASE}/gpu/available${qs ? `?${qs}` : ''}`)
  if (!res.ok) throw new Error('Failed to fetch GPU availability')
  return res.json()
}

// SSH Keys
export async function addSSHKey(name: string, publicKey: string): Promise<SSHKeyResponse> {
  const res = await fetch(`${API_BASE}/ssh-keys`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, public_key: publicKey }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.detail || 'Failed to add SSH key')
  }
  return res.json()
}

export async function deleteSSHKey(id: string): Promise<void> {
  const res = await fetch(`${API_BASE}/ssh-keys/${id}`, { method: 'DELETE' })
  if (!res.ok && res.status !== 204) throw new Error('Failed to delete SSH key')
}

// Billing
export async function fetchBillingUsage(period?: string): Promise<UsageResponse> {
  const params = new URLSearchParams()
  if (period) params.set('period', period)
  const qs = params.toString()
  const res = await fetch(`${API_BASE}/billing/usage${qs ? `?${qs}` : ''}`)
  if (!res.ok) throw new Error('Failed to fetch billing usage')
  return res.json()
}

// Spending Limits
export async function getSpendingLimit(): Promise<SpendingLimitResponse> {
  const res = await fetch(`${API_BASE}/billing/spending-limit`)
  if (!res.ok) {
    if (res.status === 404) throw new Error('no_limit')
    throw new Error('Failed to fetch spending limit')
  }
  return res.json()
}

export async function setSpendingLimit(limitDollars: number): Promise<SpendingLimitResponse> {
  const res = await fetch(`${API_BASE}/billing/spending-limit`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ monthly_limit_dollars: limitDollars }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.detail || 'Failed to set spending limit')
  }
  return res.json()
}

export async function deleteSpendingLimit(): Promise<void> {
  const res = await fetch(`${API_BASE}/billing/spending-limit`, { method: 'DELETE' })
  if (!res.ok && res.status !== 204) throw new Error('Failed to delete spending limit')
}

// Pricing Comparison
export async function fetchPricingComparison(): Promise<PricingComparisonResponse> {
  const res = await fetch(`${API_BASE}/pricing/comparison`)
  if (!res.ok) throw new Error('Failed to fetch pricing comparison')
  return res.json()
}

// Credit Balance
export async function fetchBalance(): Promise<BalanceResponse> {
  const res = await fetch(`${API_BASE}/billing/balance`)
  if (!res.ok) throw new Error('Failed to fetch balance')
  return res.json()
}

export async function purchaseCredits(amountCents: number, successUrl: string, cancelUrl: string): Promise<PurchaseCreditsResponse> {
  const res = await fetch(`${API_BASE}/billing/credits/purchase`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ amount_cents: amountCents, success_url: successUrl, cancel_url: cancelUrl }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.detail || 'Failed to purchase credits')
  }
  return res.json()
}

export async function redeemCreditCode(code: string): Promise<RedeemCodeResponse> {
  const res = await fetch(`${API_BASE}/billing/credits/redeem`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ code }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.detail || 'Failed to redeem code')
  }
  return res.json()
}

export async function updateAutoPay(enabled: boolean, thresholdCents: number, amountCents: number): Promise<BalanceResponse> {
  const res = await fetch(`${API_BASE}/billing/auto-pay`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ enabled, threshold_cents: thresholdCents, amount_cents: amountCents }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.detail || 'Failed to update auto-pay')
  }
  return res.json()
}

export async function fetchTransactions(limit?: number, before?: string): Promise<TransactionsListResponse> {
  const params = new URLSearchParams()
  if (limit) params.set('limit', String(limit))
  if (before) params.set('before', before)
  const qs = params.toString()
  const res = await fetch(`${API_BASE}/billing/transactions${qs ? `?${qs}` : ''}`)
  if (!res.ok) throw new Error('Failed to fetch transactions')
  return res.json()
}
