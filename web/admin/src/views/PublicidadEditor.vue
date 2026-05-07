<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { createAd, listAds, updateAd } from '@/api/ads'
import ImageUpload from '@/components/ImageUpload.vue'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const isEdit = computed(() => typeof route.params.id === 'string')
const error = ref('')

const form = ref({
  title: '',
  imagePath: '',
  active: false,
  slot: 1,
})

/** Alt para subida y sitio: título del banner (o texto genérico si aún no hay título). */
const imageAltForUpload = computed(() => form.value.title.trim() || 'Banner publicitario')

onMounted(async () => {
  if (!isEdit.value) return
  const data = await listAds()
  const found = data.banners.find((x) => x.id === String(route.params.id))
  if (!found) {
    error.value = 'Banner no encontrado'
    return
  }
  form.value = {
    title: found.title,
    imagePath: found.imagePath,
    active: found.active,
    slot: found.slot,
  }
})

async function onSave() {
  const payload = {
    title: form.value.title,
    imagePath: form.value.imagePath,
    active: form.value.active,
    slot: form.value.slot,
  }
  try {
    if (isEdit.value) {
      await updateAd(String(route.params.id), payload, auth.csrfToken)
    } else {
      await createAd(payload, auth.csrfToken)
    }
    await router.push('/publicidad')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'No se pudo guardar'
  }
}
</script>

<template>
  <main class="flex min-h-screen flex-col p-6">
    <h1 class="mb-6 text-2xl font-semibold">{{ isEdit ? 'Editar banner' : 'Nuevo banner' }}</h1>
    <div class="grid gap-4 rounded-xl border border-neutral-200 bg-white p-6">
      <input v-model="form.title" placeholder="Título del banner" class="rounded-lg border border-neutral-300 px-3 py-2" />
      <select v-model.number="form.slot" class="rounded-lg border border-neutral-300 px-3 py-2">
        <option :value="1">Slot 1</option>
        <option :value="2">Posición 2: Debajo del artículo</option>
      </select>
      <label class="inline-flex items-center gap-2">
        <input v-model="form.active" type="checkbox">
        Activo
      </label>
      <ImageUpload
        v-model:imagePath="form.imagePath"
        hide-alt-field
        :image-path="form.imagePath"
        :image-alt="imageAltForUpload"
      />
      <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
      <div class="flex flex-wrap gap-2">
        <button class="rounded-lg bg-neutral-900 px-4 py-2 text-white" @click="onSave">Guardar</button>
        <router-link class="rounded-lg border border-neutral-300 px-4 py-2" to="/publicidad">Cancelar</router-link>
      </div>
    </div>
    <div class="mt-auto border-t border-neutral-200 pt-6">
      <router-link class="inline-flex rounded-lg border border-neutral-300 px-4 py-2 text-sm text-neutral-800" to="/">
        ← Inicio
      </router-link>
    </div>
  </main>
</template>

