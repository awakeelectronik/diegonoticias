<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { createArticle, generateArticleSync, getArticle, updateArticle } from '@/api/articles'
import ImageUpload from '@/components/ImageUpload.vue'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const isEdit = computed(() => typeof route.params.slug === 'string')
const error = ref('')
const rawText = ref('')
const generating = ref(false)

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

const imageAltForUpload = computed(() => form.value.title.trim() || 'Imagen de la noticia')

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
      await createArticle({ ...form.value, slug: '' }, auth.csrfToken)
    }
    await router.push('/articulos')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'No se pudo guardar'
  }
}

async function onGenerate() {
  error.value = ''
  generating.value = true
  form.value.body = ''
  try {
    const g = await generateArticleSync(
      {
        rawText: rawText.value,
        tone: form.value.tone,
        titleHint: form.value.title,
        hasImage: !!form.value.image,
      },
      auth.csrfToken,
    )
    form.value.title = g.title
    form.value.description = g.metaDescription
    form.value.category = g.category
    form.value.imageAlt = g.imageAlt
    form.value.body = g.body
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'No se pudo generar'
  } finally {
    generating.value = false
  }
}
</script>

<template>
  <main class="min-h-screen p-6">
    <h1 class="mb-6 text-2xl font-semibold">{{ isEdit ? 'Editar artículo' : 'Nuevo artículo' }}</h1>
    <div class="grid max-w-2xl gap-4">
      <textarea
        v-if="!isEdit"
        v-model="rawText"
        rows="5"
        placeholder="Texto crudo para IA"
        class="rounded-lg border border-neutral-300 px-3 py-2"
      />
      <input v-model="form.title" placeholder="Título" class="rounded-lg border border-neutral-300 px-3 py-2" />
      <ImageUpload
        v-model:imagePath="form.image"
        hide-alt-field
        :image-path="form.image"
        :image-alt="imageAltForUpload"
        @update:imageAlt="(v) => { form.imageAlt = v }"
      />
      <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
      <div class="flex flex-wrap gap-3">
        <button
          v-if="!isEdit"
          class="rounded-lg bg-[var(--color-accent,#c8553d)] px-4 py-2 text-white disabled:opacity-50"
          :disabled="generating || !rawText.trim()"
          @click="onGenerate"
        >
          {{ generating ? 'Generando…' : 'Generar' }}
        </button>
        <button class="rounded-lg bg-neutral-900 px-4 py-2 text-white" @click="onSave">Guardar</button>
        <router-link class="rounded-lg border border-neutral-300 px-4 py-2" to="/articulos">Cancelar</router-link>
      </div>
    </div>
  </main>
</template>
