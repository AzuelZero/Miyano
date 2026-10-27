<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useExercises } from '../../exercises/composables'
import { filterExercises } from '../../exercises/filtering'
import SearchBar from '../../exercises/SearchBar.vue'
import type { Exercise } from '../../../types'
import { useDebouncedRef } from '../../../composables/useDebouncedRef'

const emit = defineEmits<{ select: [exercise: Exercise] }>()

const MAX_RESULTS = 10

// Same live-input → debounced-filter pattern as CatalogView.
const rawQuery = ref('')
const query = useDebouncedRef('', 250)
watch(rawQuery, (v) => (query.value = v))

// Same ['exercises'] cache key as the catalog — vue-query dedupes the call.
const { data } = useExercises()

const results = computed(() => {
  const q = query.value.trim()
  if (q === '' || !data.value) return []
  return filterExercises(data.value, { query: q, equipment: '' }).slice(0, MAX_RESULTS)
})

function pick(exercise: Exercise): void {
  emit('select', exercise)
  rawQuery.value = '' // collapse the list, ready for the next search
}
</script>

<template>
  <div class="flex flex-col gap-2">
    <SearchBar v-model="rawQuery" />

    <div
      v-if="results.length"
      class="overflow-hidden rounded-lg border border-white/6 bg-surface"
      aria-label="Resultados de búsqueda"
    >
      <button
        v-for="exercise in results"
        :key="exercise.id"
        type="button"
        class="flex w-full items-center justify-between gap-3 border-b border-white/6 px-4 py-3 text-left transition-colors last:border-b-0 hover:bg-secondary/40 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-bright"
        @click="pick(exercise)"
      >
        <span class="min-w-0 truncate text-sm font-semibold">{{ exercise.name }}</span>
        <span class="shrink-0 text-label uppercase text-muted">{{ exercise.equipment }}</span>
      </button>
    </div>

    <p v-else-if="query.trim() !== ''" class="px-1 text-sm text-muted">
      Ningún ejercicio coincide.
    </p>
    <p v-else class="px-1 text-sm text-muted">Busca en el catálogo para añadir ejercicios.</p>
  </div>
</template>
