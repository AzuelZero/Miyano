<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { Dumbbell } from 'lucide-vue-next'
import AppShell from '../components/AppShell.vue'
import BaseButton from '../components/BaseButton.vue'
import BaseCard from '../components/BaseCard.vue'
import { getFitnessLevel } from '../services/me'

// Server state through TanStack Query: cache key, loading/error for free.
const { data, isPending, isError, refetch } = useQuery({
  queryKey: ['fitness-level'],
  queryFn: getFitnessLevel,
})

// The one authored moment: the scale factor counts up on arrival.
// Reduced-motion users get the final value immediately.
const displayedFactor = ref('0.00')
const targetFactor = computed(() => data.value?.scale_factor ?? 0)

onMounted(() => animateFactor())
watch(targetFactor, () => animateFactor())

function animateFactor(): void {
  const target = targetFactor.value
  if (
    target === 0 ||
    window.matchMedia('(prefers-reduced-motion: reduce)').matches
  ) {
    displayedFactor.value = target.toFixed(2)
    return
  }
  const start = performance.now()
  const durationMs = 700
  const step = (now: number): void => {
    const t = Math.min(1, (now - start) / durationMs)
    const eased = 1 - Math.pow(2, -10 * t) // exponential ease-out
    displayedFactor.value = (target * eased).toFixed(2)
    if (t < 1) requestAnimationFrame(step)
    else displayedFactor.value = target.toFixed(2)
  }
  requestAnimationFrame(step)
}
</script>

<template>
  <AppShell>
    <div class="flex flex-col gap-6 pt-6">
      <header>
        <h1 class="font-display text-3xl font-bold tracking-[0.04em]">Hoy</h1>
        <p class="mt-1 text-muted">Tu entrenamiento, tu ritmo</p>
      </header>

      <BaseCard v-if="isPending" class="animate-pulse">
        <p class="text-muted">Cargando tu nivel…</p>
      </BaseCard>

      <BaseCard v-else-if="isError" class="border-danger/40">
        <p class="font-semibold text-text">No se pudo cargar tu nivel</p>
        <p class="mt-1 text-sm text-muted">Comprueba que la API está corriendo y reintenta.</p>
        <BaseButton variant="ghost" type="button" class="mt-4" @click="() => refetch()">
          Reintentar
        </BaseButton>
      </BaseCard>

      <BaseCard v-else-if="data">
        <p class="text-label uppercase text-muted">Tu nivel de forma</p>
        <div class="mt-2 flex items-baseline justify-between">
          <p class="text-2xl font-bold">{{ data.name }}</p>
          <p class="tabular-nums text-3xl font-bold text-primary-bright">
            ×{{ displayedFactor }}
          </p>
        </div>
        <p class="mt-3 text-sm leading-relaxed text-muted">
          Tus rutinas se sirven escaladas a este factor. Cuando mejores de forma, sube de
          nivel y la misma rutina te pedirá más.
        </p>
      </BaseCard>

      <BaseCard class="flex items-start gap-4">
        <Dumbbell class="mt-1 size-6 shrink-0 text-primary-bright" aria-hidden="true" />
        <div>
          <p class="font-semibold">El catálogo y tus rutinas llegan en la próxima lección</p>
          <p class="mt-1 text-sm leading-relaxed text-muted">
            1.324 ejercicios con GIFs de técnica y rutinas que se escalan solas ya te esperan
            en la API — el frontend los consume en el siguiente bloque.
          </p>
        </div>
      </BaseCard>
    </div>
  </AppShell>
</template>
