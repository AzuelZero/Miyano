<script setup lang="ts">
import { computed, useId } from 'vue'
import { ChevronDown, ChevronUp, X } from 'lucide-vue-next'
import { DEFAULT_SET_SECONDS, type EditorItem } from '../logic'

type Mode = 'reps' | 'time'

const MODES = [
  { id: 'reps', label: 'Repeticiones' },
  { id: 'time', label: 'Tiempo' },
] as const

const props = defineProps<{
  item: EditorItem
  index: number
}>()

const emit = defineEmits<{
  update: [item: EditorItem]
  move: [direction: -1 | 1]
  remove: []
}>()

const uid = useId()

// An item without reps falls back to time mode (the newEditorItem default is reps).
const mode = computed<Mode>(() => (props.item.reps !== undefined ? 'reps' : 'time'))

function patch(changes: Partial<EditorItem>): void {
  emit('update', { ...props.item, ...changes })
}

// Switching mode clears the field that no longer applies.
function setMode(next: Mode): void {
  if (next === mode.value) return
  if (next === 'reps') patch({ reps: props.item.reps ?? 10, durationS: undefined })
  else patch({ durationS: props.item.durationS ?? DEFAULT_SET_SECONDS, reps: undefined })
}

// Number inputs clamp to their min/max on change (mirrors the API validation ranges).
function clampEvent(event: Event, min: number, max: number): number {
  const input = event.target as HTMLInputElement
  const parsed = Number(input.value)
  const clamped = Number.isFinite(parsed) ? Math.min(max, Math.max(min, Math.round(parsed))) : min
  input.value = String(clamped)
  return clamped
}

function updateSets(event: Event): void {
  patch({ sets: clampEvent(event, 1, 99) })
}

function updateReps(event: Event): void {
  patch({ reps: clampEvent(event, 1, 999) })
}

function updateDuration(event: Event): void {
  patch({ durationS: clampEvent(event, 5, 3600) })
}

function updateRest(event: Event): void {
  patch({ restS: clampEvent(event, 0, 600) })
}
</script>

<template>
  <li class="rounded-lg border border-white/6 bg-surface p-4">
    <div class="flex items-start justify-between gap-3">
      <div class="flex min-w-0 items-baseline gap-2.5">
        <span class="tabular-nums text-label uppercase text-muted">{{ index + 1 }}</span>
        <p class="truncate font-semibold">{{ item.exerciseName }}</p>
      </div>
      <div class="flex shrink-0 items-center gap-1">
        <button
          type="button"
          aria-label="Subir ejercicio"
          class="rounded-md p-1.5 text-muted transition-colors hover:text-text focus-visible:outline-2 focus-visible:outline-primary-bright"
          @click="emit('move', -1)"
        >
          <ChevronUp class="size-4" aria-hidden="true" />
        </button>
        <button
          type="button"
          aria-label="Bajar ejercicio"
          class="rounded-md p-1.5 text-muted transition-colors hover:text-text focus-visible:outline-2 focus-visible:outline-primary-bright"
          @click="emit('move', 1)"
        >
          <ChevronDown class="size-4" aria-hidden="true" />
        </button>
        <button
          type="button"
          aria-label="Quitar ejercicio"
          class="rounded-md p-1.5 text-muted transition-colors hover:text-danger focus-visible:outline-2 focus-visible:outline-primary-bright"
          @click="emit('remove')"
        >
          <X class="size-4" aria-hidden="true" />
        </button>
      </div>
    </div>

    <div
      class="mt-3 flex w-fit rounded-full border border-white/15 p-0.5"
      role="group"
      aria-label="Objetivo del ejercicio"
    >
      <button
        v-for="option in MODES"
        :key="option.id"
        type="button"
        :aria-pressed="mode === option.id"
        class="rounded-full px-3 py-1 text-sm font-semibold transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-bright"
        :class="mode === option.id ? 'bg-primary-bright/15 text-primary-bright' : 'text-muted hover:text-text'"
        @click="setMode(option.id)"
      >
        {{ option.label }}
      </button>
    </div>

    <div class="mt-3 grid grid-cols-3 gap-3">
      <div>
        <label :for="`${uid}-sets`" class="mb-1.5 ml-0.5 block text-label uppercase text-muted">
          Series
        </label>
        <input
          :id="`${uid}-sets`"
          type="number"
          min="1"
          max="99"
          :value="item.sets"
          class="w-full rounded-md border border-white/20 bg-background px-3 py-2.5 text-text transition-colors hover:border-white/30 focus:border-transparent focus:outline-2 focus:outline-offset-1 focus:outline-primary-bright"
          @change="updateSets"
        />
      </div>

      <div v-if="mode === 'reps'">
        <label :for="`${uid}-reps`" class="mb-1.5 ml-0.5 block text-label uppercase text-muted">
          Repeticiones
        </label>
        <input
          :id="`${uid}-reps`"
          type="number"
          min="1"
          max="999"
          :value="item.reps"
          class="w-full rounded-md border border-white/20 bg-background px-3 py-2.5 text-text transition-colors hover:border-white/30 focus:border-transparent focus:outline-2 focus:outline-offset-1 focus:outline-primary-bright"
          @change="updateReps"
        />
      </div>

      <div v-else>
        <label :for="`${uid}-duration`" class="mb-1.5 ml-0.5 block text-label uppercase text-muted">
          Tiempo (s)
        </label>
        <input
          :id="`${uid}-duration`"
          type="number"
          min="5"
          max="3600"
          :value="item.durationS"
          class="w-full rounded-md border border-white/20 bg-background px-3 py-2.5 text-text transition-colors hover:border-white/30 focus:border-transparent focus:outline-2 focus:outline-offset-1 focus:outline-primary-bright"
          @change="updateDuration"
        />
      </div>

      <div>
        <label :for="`${uid}-rest`" class="mb-1.5 ml-0.5 block text-label uppercase text-muted">
          Descanso (s)
        </label>
        <input
          :id="`${uid}-rest`"
          type="number"
          min="0"
          max="600"
          :value="item.restS"
          class="w-full rounded-md border border-white/20 bg-background px-3 py-2.5 text-text transition-colors hover:border-white/30 focus:border-transparent focus:outline-2 focus:outline-offset-1 focus:outline-primary-bright"
          @change="updateRest"
        />
      </div>
    </div>
  </li>
</template>
