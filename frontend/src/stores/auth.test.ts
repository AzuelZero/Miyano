import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useAuthStore } from './auth'
import type { TokenPair } from '../types'

const pair = (n: number): TokenPair => ({
  access_token: `access-${n}`,
  refresh_token: `refresh-${n}`,
  token_type: 'Bearer',
})

function mockFetch(...responses: Array<{ status: number; body?: unknown }>) {
  const calls: Array<{ url: string; headers: Record<string, string>; body?: string }> = []
  const queue = [...responses]
  globalThis.fetch = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input)
    const headers = Object.fromEntries(new Headers(init?.headers)) as Record<string, string>
    calls.push({ url, headers, body: init?.body as string | undefined })
    const next = queue.shift() ?? { status: 500 }
    return new Response(next.body ? JSON.stringify(next.body) : null, {
      status: next.status,
      headers: { 'Content-Type': 'application/json' },
    })
  })
  return calls
}

describe('auth store', () => {
  beforeEach(() => {
    localStorage.clear()
    setActivePinia(createPinia())
  })
  afterEach(() => vi.restoreAllMocks())

  it('login persists tokens in state and localStorage', async () => {
    mockFetch({ status: 200, body: pair(1) })
    const store = useAuthStore()

    await store.login('azuel@example.com', 'password123')

    expect(store.accessToken).toBe('access-1')
    expect(localStorage.getItem('miyano.access')).toBe('access-1')
    expect(store.isAuthenticated).toBe(true)
  })

  it('logout clears everything', async () => {
    mockFetch({ status: 200, body: pair(1) })
    const store = useAuthStore()
    await store.login('azuel@example.com', 'password123')

    store.logout()

    expect(store.accessToken).toBeNull()
    expect(localStorage.getItem('miyano.access')).toBeNull()
    expect(store.isAuthenticated).toBe(false)
  })

  it('restores a previous session from localStorage', () => {
    localStorage.setItem('miyano.access', 'access-old')
    localStorage.setItem('miyano.refresh', 'refresh-old')

    const store = useAuthStore()

    expect(store.accessToken).toBe('access-old')
    expect(store.isAuthenticated).toBe(true)
  })
})

describe('api client refresh-retry', () => {
  beforeEach(() => {
    localStorage.clear()
    setActivePinia(createPinia())
  })
  afterEach(() => vi.restoreAllMocks())

  it('retries once with the fresh token after a 401', async () => {
    localStorage.setItem('miyano.access', 'access-old')
    localStorage.setItem('miyano.refresh', 'refresh-old')
    const store = useAuthStore()
    // Sequence: GET with old token → 401; POST /auth/refresh → 200 (new pair); retry → 200.
    const calls = mockFetch(
      { status: 401, body: { error: 'missing or invalid token' } },
      { status: 200, body: pair(2) },
      { status: 200, body: { id: 'beginner', name: 'Baja forma', scale_factor: 0.7 } },
    )

    const { apiJson } = await import('../services/api')
    const level = await apiJson<{ id: string }>('/me/fitness-level')

    expect(level.id).toBe('beginner')
    expect(calls).toHaveLength(3)
    expect(calls[0].headers.authorization).toBe('Bearer access-old') // sent with the stale token
    expect(calls[1].url).toContain('/auth/refresh')
    expect(calls[2].headers.authorization).toBe('Bearer access-2') // retried with the fresh one
    expect(store.isAuthenticated).toBe(true) // session survived the rotation
  })

  it('clears the session when refresh also fails', async () => {
    mockFetch(
      { status: 401, body: { error: 'missing or invalid token' } },
      { status: 401, body: { error: 'invalid refresh token' } },
    )
    localStorage.setItem('miyano.access', 'access-old')
    localStorage.setItem('miyano.refresh', 'refresh-old')
    const store = useAuthStore()

    const { apiJson, ApiError } = await import('../services/api')
    await expect(apiJson('/me/fitness-level')).rejects.toBeInstanceOf(ApiError)

    expect(store.isAuthenticated).toBe(false)
    expect(localStorage.getItem('miyano.access')).toBeNull()
  })
})
