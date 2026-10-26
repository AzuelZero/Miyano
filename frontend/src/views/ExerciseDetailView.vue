<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ChevronLeft, ImageOff } from 'lucide-vue-next'
import AppShell from '../components/AppShell.vue'
import BaseButton from '../components/BaseButton.vue'
import BaseCard from '../components/BaseCard.vue'
import { useExerciseDetail } from '../features/exercises/composables'
import { mediaSrc } from '../features/exercises/services'

const route = useRoute()
const { data, isPending, isError, refetch } = useExerciseDetail(route.params.id as string)

const lang = ref<'es' | 'en'>('es')
const gif = computed(() => (data.value ? mediaSrc(data.value.gif_url) : null))

const translation = computed(() => {
  if (!data.value) return null
  return data.value.translations.find((t) => t.lang === lang.value) ?? null
})
</script>

<template>
  <AppShell>
    <div class="flex flex-col gap-4 pt-4">
      <RouterLink
        to="/exercises"
        class="w-fit rounded-md px-2 py-1 -ml-2 text-sm font-semibold text-muted transition-colors hover:text-primary-bright focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-bright"
      >
        <ChevronLeft class="mr-1 inline size-4" aria-hidden="true" /> Catálogo
      </RouterLink>

      <div v-if="isPending" class="animate-pulse">
        <div class="mx-auto aspect-square max-w-sm rounded-lg bg-secondary/40" />
        <div class="mx-auto mt-4 h-7 w-2/3 rounded bg-secondary/40" />
      </div>

      <BaseCard v-else-if="isError" class="border-danger/40">
        <p class="font-semibold">No se pudo cargar el ejercicio</p>
        <p class="mt-1 text-sm text-muted">Comprueba que la API está corriendo y reintenta.</p>
        <BaseButton variant="ghost" type="button" class="mt-4" @click="() => refetch()">Reintentar</BaseButton>
      </BaseCard>

      <template v-else-if="data">
        <div class="mx-auto w-full max-w-sm">
          <div class="aspect-square overflow-hidden rounded-lg border border-white/6 bg-background">
            <img v-if="gif" :src="gif" :alt="`Técnica de ${data.name}`" class="size-full object-contain" />
            <div v-else class="flex size-full flex-col items-center justify-center gap-2 text-muted">
              <ImageOff class="size-8" aria-hidden="true" />
              <p class="px-6 text-center text-sm">Sin GIF (ejecuta scripts/fetch-media.ps1)</p>
            </div>
          </div>
        </div>

        <div class="flex items-start justify-between gap-3">
          <div>
            <h1 class="font-display text-2xl font-bold leading-tight">{{ data.name }}</h1>
            <p class="mt-1 flex flex-wrap gap-2 text-sm text-muted">
              <span class="rounded-full border border-white/15 px-2.5 py-0.5">{{ data.equipment }}</span>
              <span v-if="data.force_type" class="rounded-full border border-white/15 px-2.5 py-0.5">{{ data.force_type }}</span>
            </p>
          </div>

          <div
            class="flex shrink-0 rounded-full border border-white/15 p-0.5"
            role="group"
            aria-label="Idioma de las instrucciones"
          >
            <button
              v-for="option in ['es', 'en'] as const"
              :key="option"
              type="button"
              :aria-pressed="lang === option"
              class="rounded-full px-3 py-1 text-sm font-semibold transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-bright"
              :class="lang === option ? 'bg-primary-bright/15 text-primary-bright' : 'text-muted hover:text-text'"
              @click="lang = option"
            >
              {{ option.toUpperCase() }}
            </button>
          </div>
        </div>

        <BaseCard v-if="translation?.instructions || translation?.instruction_steps?.length">
          <p class="text-label uppercase text-muted">Cómo se hace</p>
          <p v-if="translation.instructions" class="mt-2 leading-relaxed">
            {{ translation.instructions }}
          </p>
          <ol v-if="translation.instruction_steps?.length" class="mt-3 space-y-2">
            <li
              v-for="(step, i) in translation.instruction_steps"
              :key="i"
              class="flex gap-3 leading-relaxed"
            >
              <span class="tabular-nums flex size-6 shrink-0 items-center justify-center rounded-full bg-secondary text-sm font-bold text-text">
                {{ i + 1 }}
              </span>
              {{ step }}
            </li>
          </ol>
        </BaseCard>
        <p v-else class="text-sm text-muted">Sin instrucciones en {{ lang === 'es' ? 'español' : 'inglés' }}.</p>
      </template>
    </div>
  </AppShell>
</template>
