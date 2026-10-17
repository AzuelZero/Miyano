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

const displayName = ref('')
const email = ref('')
const password = ref('')
const loading = ref(false)
const error = ref<string | null>(null)

function validate(): string | null {
  if (!displayName.value.trim()) return 'Ponle un nombre a tu perfil.'
  if (!email.value.trim()) return 'Falta el correo.'
  if (password.value.length < 8) return 'La contraseña necesita al menos 8 caracteres.'
  return null
}

async function submit(): Promise<void> {
  const problem = validate()
  if (problem) {
    error.value = problem
    return
  }
  loading.value = true
  error.value = null
  try {
    await auth.register(email.value.trim(), password.value, displayName.value.trim())
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
    <div class="mx-auto flex w-full max-w-sm flex-col pt-8 sm:pt-14">
      <RouterLink
        to="/login"
        class="mb-4 w-fit text-sm font-semibold text-muted transition-colors hover:text-primary-bright"
      >
        ← Volver
      </RouterLink>
      <h1 class="font-display text-3xl font-bold tracking-[0.04em]">Crea tu cuenta</h1>
      <p class="mt-1.5 text-muted">Tu nivel empieza en Baja forma y sube contigo.</p>

      <BaseCard class="mt-6">
        <form class="flex flex-col gap-4" novalidate @submit.prevent="submit">
          <BaseInput v-model="displayName" label="Nombre" placeholder="¿Cómo te llamas?" autocomplete="nickname" />
          <BaseInput v-model="email" label="Correo" type="email" placeholder="tu@correo.com" autocomplete="email" />
          <BaseInput
            v-model="password"
            label="Contraseña"
            type="password"
            autocomplete="new-password"
            hint="Mínimo 8 caracteres"
          />

          <p v-if="error" class="text-sm text-danger" role="alert">{{ error }}</p>

          <BaseButton :loading="loading">Crear cuenta</BaseButton>
        </form>
      </BaseCard>
    </div>
  </AppShell>
</template>
