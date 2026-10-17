<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import AppShell from '../components/AppShell.vue'
import BaseButton from '../components/BaseButton.vue'
import BaseCard from '../components/BaseCard.vue'
import BaseInput from '../components/BaseInput.vue'
import { authErrorMessage } from '../services/errors'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()

const email = ref('')
const password = ref('')
const loading = ref(false)
const error = ref<string | null>(null)

async function submit(): Promise<void> {
  if (!email.value.trim() || !password.value) {
    error.value = 'Rellena correo y contraseña para continuar.'
    return
  }
  loading.value = true
  error.value = null
  try {
    await auth.login(email.value.trim(), password.value)
    router.push('/home')
  } catch (err) {
    error.value = authErrorMessage(err)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <AppShell>
    <div class="flex flex-col items-center pt-10 sm:pt-16">
      <h1 class="font-display text-5xl font-bold tracking-[0.08em]">ミヤノ</h1>
      <p class="mt-2 text-muted">Tu entrenamiento, tu ritmo</p>

      <BaseCard class="mt-8 w-full max-w-sm">
        <form class="flex flex-col gap-4" novalidate @submit.prevent="submit">
          <BaseInput v-model="email" label="Correo" type="email" placeholder="tu@correo.com" autocomplete="email" />
          <BaseInput v-model="password" label="Contraseña" type="password" autocomplete="current-password" />

          <p v-if="error" class="text-sm text-danger" role="alert">{{ error }}</p>

          <BaseButton :loading="loading">Entrar</BaseButton>
        </form>

        <p class="mt-5 text-center text-sm text-muted">
          ¿Primera vez?
          <RouterLink to="/register" class="font-semibold text-primary-bright hover:underline">
            Crea tu cuenta
          </RouterLink>
        </p>
      </BaseCard>
    </div>
  </AppShell>
</template>
