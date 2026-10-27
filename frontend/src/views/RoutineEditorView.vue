<script setup lang="ts">
import { computed, ref, useId, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useQueryClient } from '@tanstack/vue-query'
import { ChevronLeft, X } from 'lucide-vue-next'
import AppShell from '../components/AppShell.vue'
import BaseButton from '../components/BaseButton.vue'
import BaseCard from '../components/BaseCard.vue'
import BaseInput from '../components/BaseInput.vue'
import { ApiError } from '../services/api'
import { useExercises } from '../features/exercises/composables'
import ExercisePicker from '../features/routines/components/ExercisePicker.vue'
import RoutineItemEditor from '../features/routines/components/RoutineItemEditor.vue'
import { useRoutine } from '../features/routines/composables'
import { createRoutine, updateRoutine } from '../features/routines/services'
import {
  estimateTotalTimeS,
  formatDuration,
  newEditorItem,
  parseTags,
  toRoutineInput,
  type EditorItem,
} from '../features/routines/logic'
import type { Exercise } from '../types'

const route = useRoute()
const router = useRouter()
const queryClient = useQueryClient()

// Two modes by route: /routines/new (no id) and /routines/:id/edit.
const routeId = typeof route.params.id === 'string' ? route.params.id : null
const isEdit = routeId !== null

// Edit mode only: the query exists solely when there is an id to hydrate from.
const routineQuery = isEdit && routeId ? useRoutine(routeId) : null

// Catalog cache (same ['exercises'] key as the catalog view — deduped fetch).
const { data: catalogData } = useExercises()

const name = ref('')
const description = ref('')
const tags = ref<string[]>([])
const tagDraft = ref('')
const items = ref<EditorItem[]>([])
const formError = ref<string | null>(null)
const saving = ref(false)

const tagsInputId = useId()

// Live estimate: every item edit re-derives the total (Regla del Dígito: tabular).
const estimate = computed(() => formatDuration(estimateTotalTimeS(items.value)))

const hydrated = ref(false)
const editLoadError = computed(() => routineQuery?.isError.value ?? false)

// Hydrate the draft once the routine arrives (edit mode).
watch(
  () => routineQuery?.data.value,
  (routine) => {
    if (!routine || hydrated.value) return
    hydrated.value = true
    name.value = routine.name
    description.value = routine.description ?? ''
    tags.value = [...routine.tags]
    items.value = routine.exercises.map((entry) => ({
      exerciseId: entry.exercise_id,
      exerciseName:
        catalogData.value?.find((exercise) => exercise.id === entry.exercise_id)?.name ??
        'Ejercicio',
      sets: entry.target_sets ?? 3,
      reps: entry.target_reps,
      durationS: entry.target_duration_s,
      restS: entry.rest_s ?? 60,
    }))
  },
  { immediate: true },
)

// Fill placeholder names once the catalog cache lands.
watch(
  () => catalogData.value,
  (catalog) => {
    if (!catalog) return
    items.value = items.value.map((item) => {
      const found = catalog.find((exercise) => exercise.id === item.exerciseId)
      return found ? { ...item, exerciseName: found.name } : item
    })
  },
)

// The draft input commits comma-separated tags on enter/blur; parseTags dedupes.
function commitTags(): void {
  tags.value = parseTags([...tags.value, tagDraft.value].join(','))
  tagDraft.value = ''
}

function removeTag(tag: string): void {
  tags.value = tags.value.filter((t) => t !== tag)
}

function addExercise(exercise: Exercise): void {
  items.value = [...items.value, newEditorItem(exercise)]
}

function updateItem(index: number, item: EditorItem): void {
  items.value = items.value.map((current, i) => (i === index ? item : current))
}

function moveItem(index: number, direction: -1 | 1): void {
  const target = index + direction
  if (target < 0 || target >= items.value.length) return
  const next = [...items.value]
  const [moved] = next.splice(index, 1)
  next.splice(target, 0, moved)
  items.value = next
}

function removeItem(index: number): void {
  items.value = items.value.filter((_, i) => i !== index)
}

function validate(): string | null {
  if (name.value.trim() === '') return 'Ponle un nombre a la rutina.'
  if (items.value.length === 0) return 'Añade al menos un ejercicio con el buscador.'
  return null
}

async function save(): Promise<void> {
  const problem = validate()
  if (problem) {
    formError.value = problem
    return
  }
  formError.value = null
  saving.value = true
  const input = toRoutineInput(name.value, description.value, items.value, tags.value)
  try {
    const saved = isEdit && routeId ? await updateRoutine(routeId, input) : await createRoutine(input)
    await queryClient.invalidateQueries({ queryKey: ['routines'] })
    if (isEdit && routeId) {
      await queryClient.invalidateQueries({ queryKey: ['routine', routeId] })
    }
    router.push(`/routines/${saved.id}`)
  } catch (err) {
    formError.value =
      err instanceof ApiError
        ? err.message
        : 'No se pudo conectar con el servidor. ¿Está la API corriendo?'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <AppShell>
    <div class="flex flex-col gap-4 pt-4">
      <RouterLink
        :to="isEdit && routeId ? `/routines/${routeId}` : '/routines'"
        class="-ml-2 w-fit rounded-md px-2 py-1 text-sm font-semibold text-muted transition-colors hover:text-primary-bright focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-bright"
      >
        <ChevronLeft class="mr-1 inline size-4" aria-hidden="true" />
        {{ isEdit ? 'Rutina' : 'Rutinas' }}
      </RouterLink>

      <h1 class="font-display text-3xl font-bold tracking-[0.04em]">
        {{ isEdit ? 'Editar rutina' : 'Nueva rutina' }}
      </h1>

      <BaseCard v-if="editLoadError" class="border-danger/40">
        <p class="font-semibold">No se pudo cargar la rutina</p>
        <p class="mt-1 text-sm text-muted">
          Se ha borrado o la API no responde. Vuelve a la lista y reintenta.
        </p>
      </BaseCard>

      <div v-else-if="isEdit && !hydrated" class="animate-pulse" aria-label="Cargando rutina">
        <div class="h-14 rounded-lg bg-secondary/40" />
        <div class="mt-3 h-24 rounded-lg bg-secondary/40" />
        <div class="mt-3 h-24 rounded-lg bg-secondary/40" />
      </div>

      <form v-else class="flex flex-col gap-4" novalidate @submit.prevent="save">
        <BaseCard class="flex flex-col gap-4">
          <BaseInput v-model="name" label="Nombre" placeholder="P. ej. Tren superior en casa" />
          <BaseInput v-model="description" label="Descripción" placeholder="Opcional" />

          <div>
            <label :for="tagsInputId" class="mb-1.5 ml-0.5 block text-label uppercase text-muted">
              Etiquetas
            </label>
            <input
              :id="tagsInputId"
              v-model="tagDraft"
              type="text"
              placeholder="pierna, casa"
              class="w-full rounded-md border border-white/20 bg-background px-3.5 py-2.5 text-text caret-primary-bright transition-colors placeholder:text-muted/70 hover:border-white/30 focus:border-transparent focus:outline-2 focus:outline-offset-1 focus:outline-primary-bright"
              @change="commitTags"
              @keydown.enter.prevent="commitTags"
            />
            <p v-if="tags.length" class="mt-2 flex flex-wrap gap-1.5">
              <span
                v-for="tag in tags"
                :key="tag"
                class="flex items-center gap-1 rounded-full border border-white/15 py-0.5 pl-2 pr-1 text-xs text-muted"
              >
                {{ tag }}
                <button
                  type="button"
                  :aria-label="`Quitar ${tag}`"
                  class="rounded-full p-0.5 text-muted transition-colors hover:text-danger focus-visible:outline-2 focus-visible:outline-primary-bright"
                  @click="removeTag(tag)"
                >
                  <X class="size-3" aria-hidden="true" />
                </button>
              </span>
            </p>
          </div>
        </BaseCard>

        <BaseCard class="flex flex-col gap-3">
          <p class="text-label uppercase text-muted">Ejercicios</p>
          <ExercisePicker @select="addExercise" />

          <p v-if="items.length === 0" class="text-sm text-muted">
            Aún no hay ejercicios. Busca arriba para añadir el primero.
          </p>
          <ul v-else class="flex flex-col gap-3">
            <RoutineItemEditor
              v-for="(item, index) in items"
              :key="`${index}-${item.exerciseId}`"
              :item="item"
              :index="index"
              @update="updateItem(index, $event)"
              @move="moveItem(index, $event)"
              @remove="removeItem(index)"
            />
          </ul>
        </BaseCard>

        <p v-if="formError" class="text-sm text-danger" role="alert">{{ formError }}</p>

        <BaseButton :loading="saving">
          {{ isEdit ? 'Guardar cambios' : 'Crear rutina' }}
        </BaseButton>
      </form>

      <div
        class="sticky bottom-20 z-10 rounded-lg border border-white/6 bg-surface/95 px-4 py-3 backdrop-blur"
      >
        <p class="text-sm text-muted">
          Tiempo estimado: <span class="tabular-nums font-semibold text-text">{{ estimate }}</span>
        </p>
      </div>
    </div>
  </AppShell>
</template>
