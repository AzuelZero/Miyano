<script setup lang="ts">
import { LogOut } from 'lucide-vue-next'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

// App chrome: minimal topbar + centered content. The bottom nav lands with
// the catalog (B2) — one screen needs no navigation (craft-floor: no dead UI).

const router = useRouter()
const auth = useAuthStore()

function logout(): void {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <div class="flex min-h-dvh flex-col">
    <header class="flex items-center justify-between px-5 py-4">
      <RouterLink to="/home" class="font-display text-xl font-bold tracking-[0.08em] text-text">
        ミヤノ
      </RouterLink>
      <button
        v-if="auth.isAuthenticated"
        type="button"
        class="flex items-center gap-1.5 rounded-md px-3 py-2 text-sm font-semibold text-muted transition-colors hover:text-primary-bright focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-bright"
        @click="logout"
      >
        <LogOut class="size-4" aria-hidden="true" />
        Salir
      </button>
    </header>

    <main class="mx-auto w-full max-w-[640px] flex-1 px-5 pb-10">
      <slot />
    </main>
  </div>
</template>
