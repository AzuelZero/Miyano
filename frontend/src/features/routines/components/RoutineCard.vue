<script setup lang="ts">
import type { RoutineSummary } from '../../../types'
import { formatDuration } from '../logic'

defineProps<{ routine: RoutineSummary }>()
</script>

<template>
  <RouterLink
    :to="`/routines/${routine.id}`"
    class="flex flex-col gap-2.5 rounded-lg border border-white/6 bg-surface p-4 transition-colors hover:border-primary-bright/40 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-bright"
  >
    <p class="font-semibold leading-snug">{{ routine.name }}</p>
    <p v-if="routine.description" class="line-clamp-1 text-sm leading-relaxed text-muted">
      {{ routine.description }}
    </p>
    <p v-if="routine.tags.length" class="flex flex-wrap gap-1.5">
      <span
        v-for="tag in routine.tags"
        :key="tag"
        class="rounded-full border border-white/15 px-2 py-0.5 text-xs text-muted"
      >
        {{ tag }}
      </span>
    </p>
    <p class="text-sm text-muted">
      <span class="tabular-nums">{{ routine.exercise_count }}</span> ejercicios ·
      <span class="tabular-nums">{{ formatDuration(routine.total_time_s) }}</span>
    </p>
  </RouterLink>
</template>
