<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getSettings, updateSettings, type Settings } from '@/api/settings'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const error = ref('')
const saved = ref(false)
const form = ref<Settings>({
  siteName: '',
  siteDescription: '',
  siteUrl: '',
  defaultOgImage: '',
  twitterHandle: '',
  adsense: {
    enabled: false,
    clientId: '',
    slot1Id: '',
    slot1Enabled: false,
    slot2Id: '',
    slot2Enabled: false,
  },
})

onMounted(async () => {
  try {
    form.value = await getSettings()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'No se pudo cargar ajustes'
  }
})

async function onSave() {
  saved.value = false
  error.value = ''
  try {
    form.value = await updateSettings(form.value, auth.csrfToken)
    saved.value = true
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'No se pudo guardar'
  }
}
</script>

<template>
  <main class="min-h-screen p-6">
    <h1 class="mb-6 text-2xl font-semibold">Ajustes</h1>
    <div class="grid gap-4 rounded-xl border border-neutral-200 bg-white p-6">
      <input v-model="form.siteName" placeholder="Nombre del sitio" class="rounded-lg border border-neutral-300 px-3 py-2" />
      <input v-model="form.siteDescription" placeholder="Descripción" class="rounded-lg border border-neutral-300 px-3 py-2" />
      <input v-model="form.siteUrl" placeholder="URL del sitio" class="rounded-lg border border-neutral-300 px-3 py-2" />
      <input v-model="form.twitterHandle" placeholder="Twitter/X handle" class="rounded-lg border border-neutral-300 px-3 py-2" />
      <label class="inline-flex items-center gap-2">
        <input v-model="form.adsense.enabled" type="checkbox">
        Habilitar AdSense
      </label>
      <input v-model="form.adsense.clientId" placeholder="AdSense client ID" class="rounded-lg border border-neutral-300 px-3 py-2" />
      <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
      <p v-if="saved" class="text-sm text-green-700">Guardado.</p>
      <button class="w-fit rounded-lg bg-neutral-900 px-4 py-2 text-white" @click="onSave">Guardar ajustes</button>
    </div>
  </main>
</template>

