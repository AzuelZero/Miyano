import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { bindAuth } from '../services/api'
import * as authApi from '../services/auth'
import type { TokenPair } from '../types'

const ACCESS_KEY = 'miyano.access'
const REFRESH_KEY = 'miyano.refresh'

/**
 * Session state (Pinia). Server data lives in TanStack Query instead —
 * never duplicate server state here (see add-frontend-scaffold design #5).
 *
 * Interim: tokens in localStorage; Capacitor Preferences replaces this in
 * the mobile block (documented debt).
 */
export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref<string | null>(localStorage.getItem(ACCESS_KEY))
  const refreshToken = ref<string | null>(localStorage.getItem(REFRESH_KEY))
  const isAuthenticated = computed(() => accessToken.value !== null)

  function persist(pair: TokenPair): void {
    accessToken.value = pair.access_token
    refreshToken.value = pair.refresh_token
    localStorage.setItem(ACCESS_KEY, pair.access_token)
    localStorage.setItem(REFRESH_KEY, pair.refresh_token)
  }

  function clear(): void {
    accessToken.value = null
    refreshToken.value = null
    localStorage.removeItem(ACCESS_KEY)
    localStorage.removeItem(REFRESH_KEY)
  }

  async function login(email: string, password: string): Promise<void> {
    persist(await authApi.login(email, password))
  }

  async function register(email: string, password: string, displayName: string): Promise<void> {
    persist(await authApi.register(email, password, displayName))
  }

  /** One refresh attempt for the api client's 401-retry. Clears session on failure. */
  async function tryRefresh(): Promise<boolean> {
    if (!refreshToken.value) return false
    try {
      persist(await authApi.refresh(refreshToken.value))
      return true
    } catch {
      clear()
      return false
    }
  }

  function logout(): void {
    clear()
  }

  // Bind this store into the api client (breaks the circular dependency).
  bindAuth(() => accessToken.value, tryRefresh)

  return { accessToken, refreshToken, isAuthenticated, login, register, logout, tryRefresh }
})
