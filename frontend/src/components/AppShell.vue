<script setup lang="ts">
import { watch } from 'vue'
import { Dumbbell, House, LogOut } from 'lucide-vue-next'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

// App chrome: minimal topbar + centered content + bottom nav (authenticated
// screens only; existing sections only — no dead UI, craft-floor).

const router = useRouter()
const auth = useAuthStore()

// A session that dies mid-screen (expired refresh) must not leave the user
// staring at loading states forever — send them to the login screen.
watch(
  () => auth.isAuthenticated,
  (authed) => {
    if (!authed) router.push('/login')
  },
)

const navItems = [
  { to: '/home', label: 'Hoy', icon: House },
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

    <main class="mx-auto w-full max-w-[640px] flex-1 px-5" :class="auth.isAuthenticated ? 'pb-24' : 'pb-10'">
      <slot />
    </main>

    <nav
      v-if="auth.isAuthenticated"
      aria-label="Secciones"
      class="fixed inset-x-0 bottom-0 mx-auto w-full max-w-[640px] border-t border-white/6 bg-surface/95 px-5 pb-[max(env(safe-area-inset-bottom),0.5rem)] pt-2 backdrop-blur"
    >
      <div class="grid grid-cols-2 gap-2">
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
