<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { createArticle, getArticle, updateArticle } from '@/api/articles'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const isEdit = computed(() => typeof route.params.slug === 'string')
const error = ref('')

const form = ref({
  title: '',
  slug: '',
  description: '',
  tone: 'informativo',
  category: 'general',
  body: '',
})

onMounted(async () => {
  if (!isEdit.value) return
  try {
    const item = await getArticle(String(route.params.slug))
    form.value = {
      title: item.title,
      slug: item.slug,
      description: item.description,
      tone: item.tone,
      category: item.category,
      body: item.body,
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'No se pudo cargar'
  }
})

async function onSave() {
  error.value = ''
  try {
    if (isEdit.value) {
      await updateArticle(String(route.params.slug), form.value, auth.csrfToken)
    } else {
      await createArticle(form.value, auth.csrfToken)
    }
    await router.push('/articulos')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'No se pudo guardar'
  }
}
</script>

<template>
  <main class="min-h-screen p-6">
    <h1 class="mb-6 text-2xl font-semibold">{{ isEdit ? 'Editar artículo' : 'Nuevo artículo' }}</h1>
    <div class="grid gap-4">
      <input v-model="form.title" placeholder="Título" class="rounded-lg border border-neutral-300 px-3 py-2" />
      <input v-model="form.slug" placeholder="Slug (opcional al crear)" class="rounded-lg border border-neutral-300 px-3 py-2" />
      <input v-model="form.description" placeholder="Descripción" class="rounded-lg border border-neutral-300 px-3 py-2" />
      <input v-model="form.category" placeholder="Categoría" class="rounded-lg border border-neutral-300 px-3 py-2" />
      <textarea v-model="form.body" rows="12" placeholder="Contenido Markdown" class="rounded-lg border border-neutral-300 px-3 py-2"></textarea>
      <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
      <div class="flex gap-3">
        <button class="rounded-lg bg-neutral-900 px-4 py-2 text-white" @click="onSave">Guardar</button>
        <router-link class="rounded-lg border border-neutral-300 px-4 py-2" to="/articulos">Cancelar</router-link>
      </div>
    </div>
  </main>
</template>

