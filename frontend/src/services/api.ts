/**
 * API client: one thin wrapper over native fetch.
 *
 * The auth store binds itself at boot via bindAuth() — this module never
 * imports the store, avoiding a circular dependency. Every request carries
 * the Bearer token; a 401 triggers ONE refresh attempt and a single retry.
 */

const BASE_URL = import.meta.env.VITE_API_URL ?? '/api/v1'

export class ApiError extends Error {
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

type TokenProvider = () => string | null
type RefreshFlow = () => Promise<boolean>

let tokenProvider: TokenProvider = () => null
let refreshFlow: RefreshFlow | null = null

/** Called once by the auth store when it initializes. */
export function bindAuth(provider: TokenProvider, refresher: RefreshFlow): void {
  tokenProvider = provider
  refreshFlow = refresher
}

export async function apiFetch(path: string, init: RequestInit = {}, retry = true): Promise<Response> {
  const token = tokenProvider()
  const response = await fetch(`${BASE_URL}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init.headers,
    },
  })

  if (response.status === 401 && retry && refreshFlow) {
    const refreshed = await refreshFlow()
    if (refreshed) return apiFetch(path, init, false)
  }
  return response
}

/** JSON helper: throws ApiError(status, server message) on !ok. */
export async function apiJson<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await apiFetch(path, init)
  if (!response.ok) {
    let message = `HTTP ${response.status}`
    try {
      const body: unknown = await response.json()
      if (isErrorBody(body)) message = body.error
    } catch {
      // non-JSON error body: keep the generic message
    }
    throw new ApiError(response.status, message)
  }
  if (response.status === 204) return undefined as T
  return (await response.json()) as T
}

function isErrorBody(body: unknown): body is { error: string } {
  return (
    typeof body === 'object' &&
    body !== null &&
    typeof (body as { error?: unknown }).error === 'string'
  )
}
