import { apiJson } from './api'
import type { TokenPair } from '../types'

export async function login(email: string, password: string): Promise<TokenPair> {
  return apiJson<TokenPair>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  })
}

export async function register(email: string, password: string, displayName: string): Promise<TokenPair> {
  return apiJson<TokenPair>('/auth/register', {
    method: 'POST',
    body: JSON.stringify({ email, password, display_name: displayName }),
  })
}

export async function refresh(refreshToken: string): Promise<TokenPair> {
  return apiJson<TokenPair>(
    '/auth/refresh',
    {
      method: 'POST',
      body: JSON.stringify({ refresh_token: refreshToken }),
    },
    false, // never retry the refresh itself: a 401 here IS the answer
  )
}
