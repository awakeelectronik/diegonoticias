<script setup lang="ts">
import { computed, ref } from 'vue'
import { uploadImage } from '@/api/images'
import { useAuthStore } from '@/stores/auth'

const props = defineProps<{
  imagePath: string
  imageAlt: string
  /** Si es true, no se muestra el campo de texto para el alt (p. ej. publicidad: se usa solo el título). */
  hideAltField?: boolean
}>()
const emit = defineEmits<{
  (e: 'update:imagePath', value: string): void
  (e: 'update:imageAlt', value: string): void
}>()

const auth = useAuthStore()
const uploading = ref(false)
const error = ref('')
const lastUploadedAt = ref<number | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)

const previewUrl = computed(() => (props.imagePath ? `${props.imagePath}-640.webp` : ''))

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
    lastUploadedAt.value = Date.now()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Error al subir imagen'
  } finally {
    uploading.value = false
    if (input) input.value = ''
  }
}

function onRemove() {
  emit('update:imagePath', '')
  lastUploadedAt.value = null
  if (fileInput.value) fileInput.value.value = ''
}

function onPickAnother() {
  fileInput.value?.click()
}
</script>

<template>
  <div class="grid gap-2">
    <label class="text-sm text-neutral-600">Imagen</label>

    <div v-if="previewUrl" class="grid gap-2">
      <div class="overflow-hidden rounded-xl border border-neutral-200 bg-neutral-50">
        <img
          :src="previewUrl"
          :alt="imageAlt"
          class="block h-48 w-full object-cover"
          @error="error = 'No se pudo cargar la previsualización (verifica que se generaron las variantes en el VPS)'"
        >
      </div>
      <p class="text-xs text-emerald-700">
        Imagen lista ✓ <span class="text-neutral-400">— variantes 320/640/1024/1600 generadas</span>
      </p>
      <p class="break-all text-xs text-neutral-500">{{ imagePath }}</p>
      <div class="flex flex-wrap gap-2">
        <button
          type="button"
          class="rounded-lg border border-neutral-300 px-3 py-1 text-sm hover:bg-neutral-50"
          @click="onPickAnother"
        >
          Reemplazar
        </button>
        <button
          type="button"
          class="rounded-lg border border-red-300 px-3 py-1 text-sm text-red-700 hover:bg-red-50"
          @click="onRemove"
        >
          Quitar imagen
        </button>
      </div>
    </div>

    <input
      ref="fileInput"
      type="file"
      accept="image/*"
      :class="previewUrl ? 'hidden' : ''"
      @change="onFileChange"
    >

    <p v-if="uploading" class="text-sm text-neutral-500">Procesando imagen…</p>
    <p v-if="error" class="text-sm text-red-600">{{ error }}</p>

    <input
      v-if="!hideAltField"
      :value="imageAlt"
      class="rounded-lg border border-neutral-300 px-3 py-2"
      placeholder="Alt de imagen"
      @input="emit('update:imageAlt', ($event.target as HTMLInputElement).value)"
    >
  </div>
</template>
