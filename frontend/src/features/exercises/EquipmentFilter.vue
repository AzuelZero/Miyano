<script setup lang="ts">
import type { EquipmentFacet } from './filtering'

defineProps<{
  facets: EquipmentFacet[]
  active: string // '' = all
}>()

const emit = defineEmits<{ select: [equipment: string] }>()
</script>

<template>
  <div class="-mx-5 overflow-x-auto px-5 pb-1" role="group" aria-label="Filtrar por equipo">
    <div class="flex w-max gap-2">
      <button
        type="button"
        :aria-pressed="active === ''"
        class="shrink-0 rounded-full border px-3.5 py-1.5 text-sm font-semibold transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-bright"
        :class="active === '' ? 'border-primary-bright bg-primary-bright/15 text-primary-bright' : 'border-white/15 text-muted hover:border-white/30 hover:text-text'"
        @click="emit('select', '')"
      >
        Todos
      </button>
      <button
        v-for="facet in facets"
        :key="facet.equipment"
        type="button"
        :aria-pressed="active === facet.equipment"
        class="shrink-0 rounded-full border px-3.5 py-1.5 text-sm font-semibold transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-bright"
        :class="active === facet.equipment ? 'border-primary-bright bg-primary-bright/15 text-primary-bright' : 'border-white/15 text-muted hover:border-white/30 hover:text-text'"
        @click="emit('select', facet.equipment)"
      >
        {{ facet.equipment }}
        <span class="tabular-nums ml-1 text-xs opacity-70">{{ facet.count }}</span>
      </button>
    </div>
  </div>
</template>
