<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { deleteAd, listAds, type Banner } from '@/api/ads'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const items = ref<Banner[]>([])
const error = ref('')

async function load() {
  try {
    const data = await listAds()
    items.value = data.banners
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'No se pudo cargar publicidad'
  }
}

onMounted(load)

async function onDelete(id: string) {
  if (!confirm('¿Eliminar banner?')) return
  await deleteAd(id, auth.csrfToken)
  await load()
}
</script>

<template>
  <main class="flex min-h-screen flex-col p-6">
    <header class="mb-6 flex items-center justify-between">
      <h1 class="text-2xl font-semibold">Publicidad</h1>
      <router-link class="rounded-lg bg-neutral-900 px-4 py-2 text-white" to="/publicidad/nueva">Nuevo banner</router-link>
    </header>
    <p v-if="error" class="mb-3 text-sm text-red-600">{{ error }}</p>
    <ul class="space-y-3">
      <li v-for="b in items" :key="b.id" class="rounded-lg border border-neutral-200 bg-white p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="font-medium">{{ b.title }}</p>
            <p class="text-sm text-neutral-500">Slot {{ b.slot }} · {{ b.active ? 'Activo' : 'Inactivo' }}</p>
          </div>
          <div class="flex gap-2">
            <router-link class="rounded border border-neutral-300 px-3 py-1 text-sm" :to="`/publicidad/${b.id}/editar`">Editar</router-link>
            <button class="rounded border border-red-300 px-3 py-1 text-sm text-red-700" @click="onDelete(b.id)">Borrar</button>
          </div>
        </div>
      </li>
    </ul>
    <div class="mt-auto border-t border-neutral-200 pt-6">
      <router-link class="inline-flex rounded-lg border border-neutral-300 px-4 py-2 text-sm text-neutral-800" to="/">
        ← Inicio
      </router-link>
    </div>
  </main>
</template>

