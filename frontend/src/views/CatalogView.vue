<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { SearchX } from 'lucide-vue-next'
import AppShell from '../components/AppShell.vue'
import BaseButton from '../components/BaseButton.vue'
import BaseCard from '../components/BaseCard.vue'
import ExerciseCard from '../features/exercises/ExerciseCard.vue'
import EquipmentFilter from '../features/exercises/EquipmentFilter.vue'
import SearchBar from '../features/exercises/SearchBar.vue'
import { useExercises } from '../features/exercises/composables'
import { equipmentFacets, filterExercises, PAGE_SIZE } from '../features/exercises/filtering'
import { useDebouncedRef } from '../composables/useDebouncedRef'

const { data, isPending, isError, refetch } = useExercises()

// raw input updates instantly; the filter follows after the user stops typing
const rawQuery = ref('')
const query = useDebouncedRef('', 250)
watch(rawQuery, (v) => (query.value = v))

const equipment = ref('')
const visibleCount = ref(PAGE_SIZE)
// any filter change restarts pagination
watch([query, equipment], () => (visibleCount.value = PAGE_SIZE))

const facets = computed(() => (data.value ? equipmentFacets(data.value) : []))
const filtered = computed(() =>
  data.value ? filterExercises(data.value, { query: query.value, equipment: equipment.value }) : [],
)
const visible = computed(() => filtered.value.slice(0, visibleCount.value))
const hasMore = computed(() => visibleCount.value < filtered.value.length)
const hasFilters = computed(() => query.value !== '' || equipment.value !== '')

function clearFilters(): void {
  rawQuery.value = ''
  equipment.value = ''
}
</script>

<template>
  <AppShell>
    <div class="flex flex-col gap-4 pt-4">
      <header>
        <h1 class="font-display text-3xl font-bold tracking-[0.04em]">Catálogo</h1>
        <p v-if="!isPending && data" class="tabular-nums mt-1 text-sm text-muted">
          {{ visible.length }} de {{ filtered.length }}<span v-if="hasFilters"> ({{ data.length }} en total)</span> ejercicios
        </p>
      </header>

      <SearchBar v-model="rawQuery" />
      <EquipmentFilter v-if="facets.length" :facets="facets" :active="equipment" @select="equipment = $event" />

      <div v-if="isPending" class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4" aria-label="Cargando catálogo">
        <div v-for="n in 8" :key="n" class="animate-pulse rounded-lg border border-white/6 bg-surface p-3">
          <div class="aspect-square rounded-md bg-secondary/40" />
          <div class="mt-3 h-4 w-3/4 rounded bg-secondary/40" />
          <div class="mt-2 h-3 w-1/2 rounded bg-secondary/40" />
        </div>
      </div>

      <BaseCard v-else-if="isError" class="border-danger/40">
        <p class="font-semibold">No se pudo cargar el catálogo</p>
        <p class="mt-1 text-sm text-muted">Comprueba que la API está corriendo y reintenta.</p>
        <BaseButton variant="ghost" type="button" class="mt-4" @click="() => refetch()">Reintentar</BaseButton>
      </BaseCard>

      <div v-else-if="visible.length" class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
        <ExerciseCard v-for="exercise in visible" :key="exercise.id" :exercise="exercise" />
      </div>

      <div v-else class="flex flex-col items-center gap-3 rounded-lg border border-white/6 bg-surface px-6 py-12 text-center">
        <SearchX class="size-8 text-muted" aria-hidden="true" />
        <p class="font-semibold">Ningún ejercicio coincide</p>
        <p class="text-sm text-muted">Prueba con otra búsqueda o quita el filtro de equipo.</p>
        <BaseButton v-if="hasFilters" variant="ghost" type="button" class="mt-2" @click="clearFilters">
          Limpiar filtros
        </BaseButton>
      </div>

      <BaseButton v-if="hasMore" variant="ghost" type="button" @click="visibleCount += PAGE_SIZE">
        Cargar más ({{ filtered.length - visibleCount }} restantes)
      </BaseButton>
    </div>
  </AppShell>
</template>
