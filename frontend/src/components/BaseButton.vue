<script setup lang="ts">
import { LoaderCircle } from 'lucide-vue-next'

withDefaults(
  defineProps<{
    variant?: 'primary' | 'ghost'
    loading?: boolean
    disabled?: boolean
    type?: 'button' | 'submit'
  }>(),
  { variant: 'primary', loading: false, disabled: false, type: 'submit' },
)
</script>

<template>
  <button
    :type="type"
    :disabled="disabled || loading"
    class="flex w-full items-center justify-center gap-2 rounded-md px-4 py-3 font-semibold transition-colors duration-150 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-bright disabled:cursor-not-allowed disabled:opacity-55"
    :class="
      variant === 'primary'
        ? 'bg-primary text-white hover:bg-primary-bright'
        : 'border border-white/15 bg-transparent text-text hover:border-primary-bright hover:text-primary-bright'
    "
  >
    <LoaderCircle v-if="loading" class="size-4 animate-spin" aria-hidden="true" />
    <slot />
  </button>
</template>
