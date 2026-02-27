const BASE = `${import.meta.env.VITE_API_URL}/api/v1`

export interface CreateURLRequest {
  target_url: string
  slug?: string
  ttl: string
  password?: string
}

export interface CreateURLResponse {
  slug: string
  short_url: string
  expires_at: string
  protected: boolean
  manage_token: string
}

export interface CheckSlugResponse {
  available: boolean
  suggestion?: string
}

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error((body as { error?: string }).error ?? res.statusText)
  }
  return res.json() as Promise<T>
}

export async function createURL(req: CreateURLRequest): Promise<CreateURLResponse> {
  const res = await fetch(`${BASE}/urls`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  })
  return handleResponse<CreateURLResponse>(res)
}

export async function checkSlug(slug: string): Promise<CheckSlugResponse> {
  const res = await fetch(`${BASE}/urls/check/${encodeURIComponent(slug)}`)
  return handleResponse<CheckSlugResponse>(res)
}

export async function unlockURL(slug: string, password: string): Promise<{ target_url: string }> {
  const res = await fetch(`${BASE}/urls/${encodeURIComponent(slug)}/unlock`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ password }),
  })
  return handleResponse<{ target_url: string }>(res)
}

export async function expireURL(slug: string, manageToken: string): Promise<void> {
  const res = await fetch(`${BASE}/urls/${encodeURIComponent(slug)}/expire`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ manage_token: manageToken }),
  })
  await handleResponse<void>(res)
}
