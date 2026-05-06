<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

async function onLogout() {
  await auth.logout()
  await router.push('/login')
}
</script>

<template>
  <main class="min-h-screen p-6">
    <header class="mb-6 flex items-center justify-between">
      <h1 class="text-2xl font-semibold">Hola, {{ auth.username }}</h1>
      <button class="rounded-lg border border-neutral-300 px-4 py-2" @click="onLogout">Cerrar sesión</button>
    </header>
    <section class="rounded-xl border border-neutral-200 bg-white p-6">
      <p class="text-neutral-700">Placeholder de artículos (Fase 2).</p>
    </section>
  </main>
</template>

