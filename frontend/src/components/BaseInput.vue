<script setup lang="ts">
import { useId } from 'vue'

const model = defineModel<string>({ required: true })

withDefaults(
  defineProps<{
    label: string
    type?: string
    placeholder?: string
    error?: string | null
    hint?: string | null
  }>(),
  { type: 'text', placeholder: '', error: null, hint: null },
)

const id = useId()
</script>

<template>
  <div>
    <label :for="id" class="mb-1.5 ml-0.5 block text-label uppercase text-muted">
      {{ label }}
    </label>
    <input
      :id="id"
      v-model="model"
      :type="type"
      :placeholder="placeholder"
      :aria-invalid="!!error || undefined"
      :aria-describedby="error ? `${id}-error` : undefined"
      class="w-full rounded-md border bg-background px-3.5 py-2.5 text-text caret-primary-bright transition-colors placeholder:text-muted/70 focus:border-transparent focus:outline-2 focus:outline-offset-1 focus:outline-primary-bright"
      :class="error ? 'border-danger' : 'border-white/20 hover:border-white/30'"
    />
    <p v-if="error" :id="`${id}-error`" class="mt-1.5 text-sm text-danger">
      {{ error }}
    </p>
    <p v-else-if="hint" class="mt-1.5 text-sm text-muted">{{ hint }}</p>
  </div>
</template>
