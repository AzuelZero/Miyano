<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useQueryClient } from '@tanstack/vue-query'
import { ChevronLeft, Pencil, Trash2 } from 'lucide-vue-next'
import AppShell from '../components/AppShell.vue'
import BaseButton from '../components/BaseButton.vue'
import BaseCard from '../components/BaseCard.vue'
import { ApiError } from '../services/api'
import { putFitnessLevel } from '../services/me'
import { useExercises } from '../features/exercises/composables'
import { useRoutine } from '../features/routines/composables'
import { deleteRoutine } from '../features/routines/services'
import { FITNESS_LEVELS, formatDuration } from '../features/routines/logic'
import type { RoutineExercise } from '../types'

const route = useRoute()
const router = useRouter()
const queryClient = useQueryClient()

const id = route.params.id as string
const { data, isPending, isError, error, refetch } = useRoutine(id)

const isNotFound = computed(() => error.value instanceof ApiError && error.value.status === 404)

// Exercise names come from the catalog cache (same ['exercises'] key as the
// catalog view — vue-query dedupes the call). 'Ejercicio' while it loads.
const { data: catalog } = useExercises()

function exerciseName(exerciseId: string): string {
  return catalog.value?.find((exercise) => exercise.id === exerciseId)?.name ?? 'Ejercicio'
}

interface RoutineRow {
  key: string
  name: string
  volume: string // '3 × 12' or '3 × 45 s'
  rest: string | null // '90 s'
}

function toRows(entries: RoutineExercise[]): RoutineRow[] {
  return entries.map((entry) => {
    const sets = entry.target_sets ?? '—'
    const volume =
      entry.target_reps !== undefined
        ? `${sets} × ${entry.target_reps}`
        : entry.target_duration_s !== undefined
          ? `${sets} × ${entry.target_duration_s} s`
          : '—'
    return {
      key: `${entry.position}-${entry.exercise_id}`,
      name: exerciseName(entry.exercise_id),
      volume,
      rest: entry.rest_s !== undefined ? `${entry.rest_s} s` : null,
    }
  })
}

const baseRows = computed(() => (data.value ? toRows(data.value.exercises) : []))
const scaledRows = computed(() => (data.value ? toRows(data.value.scaled_exercises) : []))

// Level switch: PUT the new level, then refetch the routine (it re-scales)
// and the home level card.
const levelPending = ref(false)
const levelError = ref<string | null>(null)

async function setLevel(levelId: string): Promise<void> {
  if (levelPending.value || !data.value || data.value.level.id === levelId) return
  levelPending.value = true
  levelError.value = null
  try {
    await putFitnessLevel(levelId)
    await queryClient.invalidateQueries({ queryKey: ['routine', id] })
    await queryClient.invalidateQueries({ queryKey: ['fitness-level'] })
  } catch (err) {
    levelError.value =
      err instanceof ApiError
        ? err.message
        : 'No se pudo guardar tu nivel. Inténtalo de nuevo en un momento.'
  } finally {
    levelPending.value = false
  }
}

const isDeleting = ref(false)
const deleteError = ref<string | null>(null)

function confirmDelete(): void {
  if (!data.value || isDeleting.value) return
  const confirmed = window.confirm('¿Borrar esta rutina? Esta acción no se puede deshacer.')
  if (!confirmed) return
  isDeleting.value = true
  deleteError.value = null
  void deleteRoutine(id)
    .then(() => queryClient.invalidateQueries({ queryKey: ['routines'] }))
    .then(() => router.push('/routines'))
    .catch((err: unknown) => {
      deleteError.value =
        err instanceof ApiError
          ? err.message
          : 'No se pudo borrar la rutina. Inténtalo de nuevo en un momento.'
    })
    .finally(() => {
      isDeleting.value = false
    })
}
</script>

<template>
  <AppShell>
    <div class="flex flex-col gap-4 pt-4">
      <RouterLink
        to="/routines"
        class="-ml-2 w-fit rounded-md px-2 py-1 text-sm font-semibold text-muted transition-colors hover:text-primary-bright focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-bright"
      >
        <ChevronLeft class="mr-1 inline size-4" aria-hidden="true" /> Rutinas
      </RouterLink>

      <div v-if="isPending" class="animate-pulse" aria-label="Cargando rutina">
        <div class="h-8 w-2/3 rounded bg-secondary/40" />
        <div class="mt-3 h-4 w-1/2 rounded bg-secondary/40" />
        <div class="mt-6 h-44 rounded-lg bg-secondary/40" />
      </div>

      <BaseCard v-else-if="isNotFound" class="border-danger/40">
        <p class="font-semibold">Esta rutina no existe</p>
        <p class="mt-1 text-sm text-muted">Se ha borrado o pertenece a otra cuenta.</p>
        <BaseButton variant="ghost" type="button" class="mt-4" @click="router.push('/routines')">
          Ver mis rutinas
        </BaseButton>
      </BaseCard>

      <BaseCard v-else-if="isError" class="border-danger/40">
        <p class="font-semibold">No se pudo cargar la rutina</p>
        <p class="mt-1 text-sm text-muted">Comprueba que la API está corriendo y reintenta.</p>
        <BaseButton variant="ghost" type="button" class="mt-4" @click="() => refetch()">
          Reintentar
        </BaseButton>
      </BaseCard>

      <template v-else-if="data">
        <header>
          <h1 class="font-display text-3xl font-bold tracking-[0.04em]">{{ data.name }}</h1>
          <p v-if="data.description" class="mt-1.5 leading-relaxed text-muted">
            {{ data.description }}
          </p>
          <p v-if="data.tags.length" class="mt-2 flex flex-wrap gap-1.5">
            <span
              v-for="tag in data.tags"
              :key="tag"
              class="rounded-full border border-white/15 px-2 py-0.5 text-xs text-muted"
            >
              {{ tag }}
            </span>
          </p>
        </header>

        <BaseCard>
          <div class="flex items-baseline justify-between gap-3">
            <h2 class="text-lg font-semibold">Base</h2>
            <p class="tabular-nums text-sm text-muted">{{ formatDuration(data.total_time_s) }}</p>
          </div>
          <ul class="mt-2">
            <li
              v-for="row in baseRows"
              :key="row.key"
              class="flex items-baseline justify-between gap-3 border-b border-white/6 py-2.5 last:border-b-0"
            >
              <p class="min-w-0 truncate">{{ row.name }}</p>
              <p class="tabular-nums shrink-0 text-sm text-muted">
                {{ row.volume }}<span v-if="row.rest"> · descanso {{ row.rest }}</span>
              </p>
            </li>
          </ul>
        </BaseCard>

        <BaseCard>
          <div class="flex items-baseline justify-between gap-3">
            <h2 class="text-lg font-semibold">
              Tu nivel: {{ data.level.name }}
              <span class="tabular-nums">×{{ data.level.scale_factor.toFixed(2) }}</span>
            </h2>
            <p class="tabular-nums text-sm text-muted">
              {{ formatDuration(data.total_time_scaled_s) }}
            </p>
          </div>
          <ul class="mt-2">
            <li
              v-for="row in scaledRows"
              :key="row.key"
              class="flex items-baseline justify-between gap-3 border-b border-white/6 py-2.5 last:border-b-0"
            >
              <p class="min-w-0 truncate">{{ row.name }}</p>
              <p class="tabular-nums shrink-0 text-sm text-muted">
                {{ row.volume }}<span v-if="row.rest"> · descanso {{ row.rest }}</span>
              </p>
            </li>
          </ul>
        </BaseCard>

        <BaseCard>
          <p class="text-label uppercase text-muted">Tu nivel de forma</p>
          <div class="mt-3 flex flex-wrap gap-2" role="group" aria-label="Cambiar tu nivel de forma">
            <button
              v-for="level in FITNESS_LEVELS"
              :key="level.id"
              type="button"
              :disabled="levelPending"
              :aria-pressed="data.level.id === level.id"
              class="rounded-full border px-3.5 py-1.5 text-sm font-semibold transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-bright disabled:cursor-not-allowed disabled:opacity-55"
              :class="
                data.level.id === level.id
                  ? 'border-primary-bright bg-primary-bright/15 text-primary-bright'
                  : 'border-white/15 text-muted hover:border-white/30 hover:text-text'
              "
              @click="setLevel(level.id)"
            >
              {{ level.label }}
            </button>
          </div>
          <p v-if="levelError" class="mt-2 text-sm text-danger" role="alert">{{ levelError }}</p>
          <p class="mt-3 text-sm leading-relaxed text-muted">
            Al cambiar de nivel, tus rutinas se recalculan a la nueva escala.
          </p>
        </BaseCard>

        <div class="grid grid-cols-2 gap-3">
          <BaseButton
            variant="ghost"
            type="button"
            @click="router.push(`/routines/${data.id}/edit`)"
          >
            <Pencil class="size-4" aria-hidden="true" />
            Editar
          </BaseButton>
          <button
            type="button"
            :disabled="isDeleting"
            class="flex items-center justify-center gap-2 rounded-md border border-danger/50 px-4 py-3 font-semibold text-danger transition-colors duration-150 hover:bg-danger/10 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-bright disabled:cursor-not-allowed disabled:opacity-55"
            @click="confirmDelete"
          >
            <Trash2 class="size-4" aria-hidden="true" />
            Borrar
          </button>
        </div>
        <p v-if="deleteError" class="text-sm text-danger" role="alert">{{ deleteError }}</p>
      </template>
    </div>
  </AppShell>
</template>
