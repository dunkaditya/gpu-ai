import type {
  InstanceResponse,
  AvailableOffering,
  SSHKeyResponse,
  UsageResponse,
  CreateInstanceRequest,
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
