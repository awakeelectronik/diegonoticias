<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { deleteArticle, listArticles, type Article } from '@/api/articles'
import { getSettings } from '@/api/settings'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const items = ref<Article[]>([])
const error = ref('')
const homeHref = ref('/')

const displayUsername = computed(() => {
  const u = auth.username
  if (!u) return ''
  return u.charAt(0).toUpperCase() + u.slice(1)
})

function resolveHomeHref(siteUrl: string | undefined): string {
  const raw = siteUrl?.trim()
  if (!raw) return '/'
  try {
    return new URL(raw).href
  } catch {
    return '/'
  }
}

onMounted(async () => {
  try {
    const [data, settings] = await Promise.all([
      listArticles(),
      getSettings().catch(() => null),
    ])
    items.value = data.items
    homeHref.value = resolveHomeHref(settings?.siteUrl)
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
    <header class="mb-6 grid gap-4 md:grid-cols-[auto_1fr] md:items-center">
      <div class="flex flex-wrap items-baseline gap-x-3 gap-y-1">
        <h1 class="text-2xl font-semibold leading-tight">Hola, {{ displayUsername }}</h1>
        <a
          :href="homeHref"
          target="_blank"
          rel="noopener noreferrer"
          class="inline-flex min-h-10 shrink-0 items-center rounded-lg border border-neutral-300 px-3 py-1.5 text-sm font-medium text-neutral-900 hover:bg-neutral-50"
        >
          Ver página
        </a>
      </div>
      <div class="grid grid-cols-2 gap-2 sm:grid-cols-4 md:justify-self-end">
        <router-link class="inline-flex min-h-12 items-center justify-center rounded-lg border border-neutral-300 px-4 py-2 text-center" to="/publicidad">Publicidad</router-link>
        <router-link class="inline-flex min-h-12 items-center justify-center rounded-lg border border-neutral-300 px-4 py-2 text-center" to="/ajustes">Ajustes</router-link>
        <router-link class="inline-flex min-h-12 items-center justify-center rounded-lg bg-neutral-900 px-4 py-2 text-center text-white" to="/articulos/nuevo">Nuevo</router-link>
        <button class="inline-flex min-h-12 items-center justify-center rounded-lg border border-neutral-300 px-4 py-2 text-center" @click="onLogout">Cerrar sesión</button>
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

