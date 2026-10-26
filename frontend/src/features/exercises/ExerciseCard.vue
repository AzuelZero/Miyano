<script setup lang="ts">
import { computed } from 'vue'
import type { Exercise } from '../../types'
import { mediaSrc } from './services'

const props = defineProps<{ exercise: Exercise }>()

const thumb = computed(() => mediaSrc(props.exercise.image_url))
const initial = computed(() => props.exercise.name.charAt(0).toUpperCase())
</script>

<template>
  <RouterLink
    :to="`/exercises/${exercise.id}`"
    class="group flex flex-col gap-2 rounded-lg border border-white/6 bg-surface p-3 transition-colors hover:border-primary-bright/40 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-bright"
  >
    <div class="relative aspect-square overflow-hidden rounded-md bg-background">
      <img
        v-if="thumb"
        :src="thumb"
        :alt="exercise.name"
        loading="lazy"
        class="size-full object-cover transition-transform duration-200 group-hover:scale-105"
      />
      <span v-else class="flex size-full items-center justify-center text-3xl font-bold text-muted" aria-hidden="true">
        {{ initial }}
      </span>
    </div>
    <p class="line-clamp-2 text-sm font-semibold leading-snug">{{ exercise.name }}</p>
    <p class="text-label uppercase text-muted">{{ exercise.equipment }}</p>
  </RouterLink>
</template>
