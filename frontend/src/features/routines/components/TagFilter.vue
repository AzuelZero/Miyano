<script setup lang="ts">
defineProps<{
  tags: string[]
  modelValue?: string // '' = all
}>()

const emit = defineEmits<{ 'update:modelValue': [tag: string] }>()
</script>

<template>
  <div class="-mx-5 overflow-x-auto px-5 pb-1" role="group" aria-label="Filtrar por etiqueta">
    <div class="flex w-max gap-2">
      <button
        type="button"
        :aria-pressed="(modelValue ?? '') === ''"
        class="shrink-0 rounded-full border px-3.5 py-1.5 text-sm font-semibold transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-bright"
        :class="
          (modelValue ?? '') === ''
            ? 'border-primary-bright bg-primary-bright/15 text-primary-bright'
            : 'border-white/15 text-muted hover:border-white/30 hover:text-text'
        "
        @click="emit('update:modelValue', '')"
      >
        Todas
      </button>
      <button
        v-for="tag in tags"
        :key="tag"
        type="button"
        :aria-pressed="modelValue === tag"
        class="shrink-0 rounded-full border px-3.5 py-1.5 text-sm font-semibold transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-bright"
        :class="
          modelValue === tag
            ? 'border-primary-bright bg-primary-bright/15 text-primary-bright'
            : 'border-white/15 text-muted hover:border-white/30 hover:text-text'
        "
        @click="emit('update:modelValue', tag)"
      >
        {{ tag }}
      </button>
    </div>
  </div>
</template>
