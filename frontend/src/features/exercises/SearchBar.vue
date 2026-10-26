<script setup lang="ts">
import { Search, X } from 'lucide-vue-next'

const model = defineModel<string>({ required: true })
</script>

<template>
  <div class="relative">
    <Search class="pointer-events-none absolute left-3.5 top-1/2 size-4 -translate-y-1/2 text-muted" aria-hidden="true" />
    <input
      v-model="model"
      type="search"
      placeholder="Buscar ejercicio…"
      aria-label="Buscar ejercicio"
      class="w-full rounded-md border border-white/20 bg-background py-2.5 pl-10 pr-10 text-text caret-primary-bright transition-colors placeholder:text-muted/70 hover:border-white/30 focus:border-transparent focus:outline-2 focus:outline-offset-1 focus:outline-primary-bright"
    />
    <button
      v-if="model"
      type="button"
      aria-label="Limpiar búsqueda"
      class="absolute right-2.5 top-1/2 -translate-y-1/2 rounded-md p-1.5 text-muted transition-colors hover:text-text focus-visible:outline-2 focus-visible:outline-primary-bright"
      @click="model = ''"
    >
      <X class="size-4" aria-hidden="true" />
    </button>
  </div>
</template>
