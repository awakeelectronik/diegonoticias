<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { parse, ALL } from 'partial-json'
import { createArticle, getArticle, updateArticle } from '@/api/articles'
import ImageUpload from '@/components/ImageUpload.vue'
import StreamingTextarea from '@/components/StreamingTextarea.vue'
import { useSSE } from '@/composables/useSSE'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const isEdit = computed(() => typeof route.params.slug === 'string')
const error = ref('')
const rawText = ref('')
const streamBuffer = ref('')
const sse = useSSE()

const form = ref({
  title: '',
  slug: '',
  description: '',
  tone: 'informativo',
  category: 'general',
  image: '',
  imageAlt: '',
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
      image: item.image ?? '',
      imageAlt: item.imageAlt ?? '',
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

async function onGenerate() {
  error.value = ''
  streamBuffer.value = ''
  form.value.body = ''
  await sse.start(
    '/admin/api/articulos/generar',
    { rawText: rawText.value, tone: form.value.tone, titleHint: form.value.title, hasImage: !!form.value.image },
    (chunk) => {
      streamBuffer.value += chunk
      try {
        const partial = parse(streamBuffer.value, ALL) as {
          title?: string
          body?: string
          metaDescription?: string
          category?: string
          imageAlt?: string
        }
        if (partial.title) form.value.title = partial.title
        if (partial.metaDescription) form.value.description = partial.metaDescription
        if (partial.category) form.value.category = partial.category
        if (partial.imageAlt) form.value.imageAlt = partial.imageAlt
        if (partial.body) form.value.body = partial.body
      } catch {
        // partial-json puede fallar en chunks intermedios incompletos
      }
    },
    () => {},
    (msg) => {
      error.value = msg
    },
  )
}
</script>

<template>
  <main class="min-h-screen p-6">
    <h1 class="mb-6 text-2xl font-semibold">{{ isEdit ? 'Editar artículo' : 'Nuevo artículo' }}</h1>
    <div class="grid gap-4">
      <textarea v-model="rawText" rows="5" placeholder="Texto crudo para IA" class="rounded-lg border border-neutral-300 px-3 py-2" />
      <input v-model="form.title" placeholder="Título" class="rounded-lg border border-neutral-300 px-3 py-2" />
      <input v-model="form.slug" placeholder="Slug (opcional al crear)" class="rounded-lg border border-neutral-300 px-3 py-2" />
      <input v-model="form.description" placeholder="Descripción" class="rounded-lg border border-neutral-300 px-3 py-2" />
      <input v-model="form.category" placeholder="Categoría" class="rounded-lg border border-neutral-300 px-3 py-2" />
      <ImageUpload v-model:imagePath="form.image" v-model:imageAlt="form.imageAlt" :image-path="form.image" :image-alt="form.imageAlt" />
      <StreamingTextarea v-model="form.body" :streaming="sse.streaming" />
      <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
      <div class="flex gap-3">
        <button
          v-if="!isEdit"
          class="rounded-lg bg-[var(--color-accent,#c8553d)] px-4 py-2 text-white"
          :disabled="sse.streaming || !rawText.trim()"
          @click="onGenerate"
        >
          {{ sse.streaming ? 'Generando…' : 'Generar' }}
        </button>
        <button
          v-if="!isEdit && sse.streaming"
          class="rounded-lg border border-neutral-300 px-4 py-2"
          @click="sse.stop"
        >
          Detener
        </button>
        <button class="rounded-lg bg-neutral-900 px-4 py-2 text-white" @click="onSave">Guardar</button>
        <router-link class="rounded-lg border border-neutral-300 px-4 py-2" to="/articulos">Cancelar</router-link>
      </div>
    </div>
  </main>
</template>

