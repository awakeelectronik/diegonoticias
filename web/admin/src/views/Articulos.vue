<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { deleteArticle, listArticles, type Article } from '@/api/articles'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const items = ref<Article[]>([])
const error = ref('')

onMounted(async () => {
  try {
    const data = await listArticles()
    items.value = data.items
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'No se pudo cargar'
  }
})

async function onLogout() {
  await auth.logout()
  await router.push('/login')
}

async function onDelete(slug: string) {
  if (!confirm('¿Borrar artículo?')) return
  await deleteArticle(slug, auth.csrfToken)
  items.value = items.value.filter((i) => i.slug !== slug)
}
</script>

<template>
  <main class="min-h-screen p-6">
    <header class="mb-6 flex items-center justify-between">
      <h1 class="text-2xl font-semibold">Hola, {{ auth.username }}</h1>
      <div class="flex gap-2">
        <router-link class="rounded-lg border border-neutral-300 px-4 py-2" to="/publicidad">Publicidad</router-link>
        <router-link class="rounded-lg border border-neutral-300 px-4 py-2" to="/ajustes">Ajustes</router-link>
        <router-link class="rounded-lg bg-neutral-900 px-4 py-2 text-white" to="/articulos/nuevo">Nuevo</router-link>
        <button class="rounded-lg border border-neutral-300 px-4 py-2" @click="onLogout">Cerrar sesión</button>
      </div>
    </header>
    <section class="rounded-xl border border-neutral-200 bg-white p-6">
      <p v-if="error" class="mb-3 text-sm text-red-600">{{ error }}</p>
      <ul class="space-y-3">
        <li v-for="item in items" :key="item.slug" class="flex items-center justify-between rounded-lg border border-neutral-200 p-3">
          <div>
            <p class="font-medium">{{ item.title }}</p>
            <p class="text-sm text-neutral-500">/{{ item.slug }}/</p>
          </div>
          <div class="flex gap-2">
            <router-link class="rounded border border-neutral-300 px-3 py-1 text-sm" :to="`/articulos/${item.slug}/editar`">Editar</router-link>
            <button class="rounded border border-red-300 px-3 py-1 text-sm text-red-700" @click="onDelete(item.slug)">Borrar</button>
          </div>
        </li>
      </ul>
    </section>
  </main>
</template>

