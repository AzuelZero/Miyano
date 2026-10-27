<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'
import { ClipboardList, Plus, SearchX } from 'lucide-vue-next'
import AppShell from '../components/AppShell.vue'
import BaseButton from '../components/BaseButton.vue'
import BaseCard from '../components/BaseCard.vue'
import RoutineCard from '../features/routines/components/RoutineCard.vue'
import TagFilter from '../features/routines/components/TagFilter.vue'
import { useRoutines } from '../features/routines/composables'
import { listRoutines } from '../features/routines/services'

const route = useRoute()
const router = useRouter()

const activeTag = ref(typeof route.query.tag === 'string' ? route.query.tag : '')

// Chip clicks drive the URL (?tag=, shareable state); browser back/forward
// drive the chips. Each side ignores echoes of its own change.
watch(
  () => route.query.tag,
  (tag) => {
    const next = typeof tag === 'string' ? tag : ''
    if (next !== activeTag.value) activeTag.value = next
  },
)

watch(activeTag, (tag) => {
  const current = typeof route.query.tag === 'string' ? route.query.tag : undefined
  if ((tag === '' ? undefined : tag) !== current) {
    void router.replace({ query: tag === '' ? {} : { tag } })
  }
})

// Unfiltered listing: feeds the tag facets. Shares the ['routines', null]
// cache entry with the display query below (one fetch while no tag is active).
const { data: allRoutines } = useRoutines()

// Display list: server-filtered by the selected tag. The reactive key mirrors
// the contracted ['routines', tag ?? null] shape; useRoutines takes a plain
// string captured at setup, so the tag switch is keyed here instead.
const { data, isPending, isError, refetch } = useQuery({
  queryKey: computed(() => ['routines', activeTag.value === '' ? null : activeTag.value]),
  queryFn: () => listRoutines(activeTag.value === '' ? undefined : activeTag.value),
})

// Tag facets: union of the tags of every routine in the (unfiltered) listing.
const allTags = computed<string[]>(() => {
  const seen = new Set<string>()
  for (const routine of allRoutines.value ?? []) for (const tag of routine.tags) seen.add(tag)
  return [...seen].sort((a, b) => a.localeCompare(b))
})
</script>

<template>
  <AppShell>
    <div class="flex flex-col gap-4 pt-4">
      <header class="flex items-end justify-between gap-3">
        <div>
          <h1 class="font-display text-3xl font-bold tracking-[0.04em]">Rutinas</h1>
          <p v-if="!isPending && data" class="tabular-nums mt-1 text-sm text-muted">
            {{ data.length }} rutinas<span v-if="activeTag"> con «{{ activeTag }}»</span>
          </p>
        </div>
        <RouterLink
          to="/routines/new"
          class="flex shrink-0 items-center justify-center gap-2 rounded-md bg-primary px-4 py-2.5 font-semibold text-white transition-colors duration-150 hover:bg-primary-bright focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-bright"
        >
          <Plus class="size-4" aria-hidden="true" />
          Nueva rutina
        </RouterLink>
      </header>

      <TagFilter v-if="allTags.length" v-model="activeTag" :tags="allTags" />

      <div v-if="isPending" class="flex flex-col gap-3" aria-label="Cargando rutinas">
        <div v-for="n in 4" :key="n" class="animate-pulse rounded-lg border border-white/6 bg-surface p-4">
          <div class="h-5 w-2/3 rounded bg-secondary/40" />
          <div class="mt-3 h-3 w-1/3 rounded bg-secondary/40" />
          <div class="mt-3 h-3 w-1/2 rounded bg-secondary/40" />
        </div>
      </div>

      <BaseCard v-else-if="isError" class="border-danger/40">
        <p class="font-semibold">No se pudo cargar las rutinas</p>
        <p class="mt-1 text-sm text-muted">Comprueba que la API está corriendo y reintenta.</p>
        <BaseButton variant="ghost" type="button" class="mt-4" @click="() => refetch()">Reintentar</BaseButton>
      </BaseCard>

      <div v-else-if="data?.length" class="flex flex-col gap-3">
        <RoutineCard v-for="routine in data" :key="routine.id" :routine="routine" />
      </div>

      <div
        v-else-if="activeTag !== ''"
        class="flex flex-col items-center gap-3 rounded-lg border border-white/6 bg-surface px-6 py-12 text-center"
      >
        <SearchX class="size-8 text-muted" aria-hidden="true" />
        <p class="font-semibold">Ninguna rutina con esta etiqueta</p>
        <p class="text-sm text-muted">Prueba con otra etiqueta o mira todas tus rutinas.</p>
        <BaseButton variant="ghost" type="button" class="mt-2" @click="activeTag = ''">
          Ver todas
        </BaseButton>
      </div>

      <div v-else class="flex flex-col items-center gap-3 rounded-lg border border-white/6 bg-surface px-6 py-12 text-center">
        <ClipboardList class="size-8 text-muted" aria-hidden="true" />
        <p class="font-semibold">Aún no hay rutinas</p>
        <p class="text-sm text-muted">Crea tu primera rutina y se escalará sola a tu nivel.</p>
        <BaseButton type="button" class="mt-2" @click="router.push('/routines/new')">
          Crear una rutina
        </BaseButton>
      </div>
    </div>
  </AppShell>
</template>
