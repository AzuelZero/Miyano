<script setup lang="ts">
import { watch } from 'vue'
import { ClipboardList, Dumbbell, House, LogOut, WifiOff } from 'lucide-vue-next'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useSyncStore } from '../stores/sync'

// App chrome: minimal topbar + centered content + bottom nav (authenticated
// screens only; existing sections only — no dead UI, craft-floor).

const router = useRouter()
const auth = useAuthStore()
const sync = useSyncStore()
sync.listen()

// A session that dies mid-screen (expired refresh) must not leave the user
// staring at loading states forever — send them to the login screen.
watch(
  () => auth.isAuthenticated,
  (authed) => {
    if (!authed) router.push('/login')
  },
)

// "De vuelta — sincronizado": brief acknowledgment when connectivity returns.
const BACK_TOAST_MS = 3000
let backToastTimer: ReturnType<typeof setTimeout> | undefined
watch(
  () => sync.backOnline,
  (back) => {
    if (!back) return
    clearTimeout(backToastTimer)
    backToastTimer = setTimeout(() => sync.acknowledgeBackOnline(), BACK_TOAST_MS)
  },
)

const navItems = [
  { to: '/home', label: 'Hoy', icon: House },
  { to: '/routines', label: 'Rutinas', icon: ClipboardList },
  { to: '/exercises', label: 'Catálogo', icon: Dumbbell },
]

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

    <!-- Offline state lives under the header, ALWAYS honest (design decision 8). -->
    <div
      v-if="auth.isAuthenticated && !sync.online"
      role="status"
      class="mx-auto w-full max-w-[640px] px-5"
    >
      <div class="flex items-center gap-2 rounded-lg border border-white/10 bg-surface px-3.5 py-2.5 text-sm text-muted">
        <WifiOff class="size-4 shrink-0 text-muted" aria-hidden="true" />
        <span>Sin conexión — tus cambios se guardan en este dispositivo y suben al volver.</span>
      </div>
    </div>

    <div
      v-else-if="auth.isAuthenticated && sync.backOnline"
      role="status"
      class="mx-auto w-full max-w-[640px] px-5"
    >
      <div class="flex items-center gap-2 rounded-lg border border-primary-bright/30 bg-surface px-3.5 py-2.5 text-sm text-text">
        <span class="text-primary-bright">De vuelta — sincronizado.</span>
      </div>
    </div>

    <main class="mx-auto w-full max-w-[640px] flex-1 px-5" :class="auth.isAuthenticated ? 'pb-24' : 'pb-10'">
      <slot />
    </main>

    <nav
      v-if="auth.isAuthenticated"
      aria-label="Secciones"
      class="fixed inset-x-0 bottom-0 mx-auto w-full max-w-[640px] border-t border-white/6 bg-surface/95 px-5 pb-[max(env(safe-area-inset-bottom),0.5rem)] pt-2 backdrop-blur"
    >
      <div class="grid grid-cols-3 gap-2">
        <RouterLink
          v-for="item in navItems"
          :key="item.to"
          :to="item.to"
          class="flex flex-col items-center gap-0.5 rounded-md py-1.5 text-muted transition-colors hover:text-text focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-bright"
          active-class="!text-primary-bright"
        >
          <component :is="item.icon" class="size-5" aria-hidden="true" />
          <span class="text-label">{{ item.label }}</span>
        </RouterLink>
      </div>
    </nav>
  </div>
</template>
