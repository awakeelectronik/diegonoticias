<script setup lang="ts">
defineProps<{
  modelValue: string
  streaming: boolean
}>()
defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()
</script>

<template>
  <textarea
    :value="modelValue"
    rows="12"
    class="rounded-lg border border-neutral-300 px-3 py-2"
    @input="$emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
  />
  <p v-if="streaming" class="text-sm text-neutral-500">Generando en vivo…</p>
</template>

