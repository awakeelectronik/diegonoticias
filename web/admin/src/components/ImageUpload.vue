<script setup lang="ts">
import { ref } from 'vue'
import { uploadImage } from '@/api/images'
import { useAuthStore } from '@/stores/auth'

const props = defineProps<{
  imagePath: string
  imageAlt: string
}>()
const emit = defineEmits<{
  (e: 'update:imagePath', value: string): void
  (e: 'update:imageAlt', value: string): void
}>()

const auth = useAuthStore()
const uploading = ref(false)
const error = ref('')

async function onFileChange(ev: Event) {
  const input = ev.target as HTMLInputElement
  const f = input.files?.[0]
  if (!f) return
  error.value = ''
  uploading.value = true
  try {
    const data = await uploadImage(f, auth.csrfToken, props.imageAlt)
    emit('update:imagePath', data.basePath)
    emit('update:imageAlt', data.alt || props.imageAlt)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Error al subir imagen'
  } finally {
    uploading.value = false
  }
}
</script>

<template>
  <div class="grid gap-2">
    <label class="text-sm text-neutral-600">Imagen</label>
    <input type="file" accept="image/*" @change="onFileChange">
    <p v-if="uploading" class="text-sm text-neutral-500">Procesando imagen…</p>
    <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
    <input
      :value="imageAlt"
      class="rounded-lg border border-neutral-300 px-3 py-2"
      placeholder="Alt de imagen"
      @input="emit('update:imageAlt', ($event.target as HTMLInputElement).value)"
    >
    <p v-if="imagePath" class="text-xs text-neutral-500">Base path: {{ imagePath }}</p>
  </div>
</template>

