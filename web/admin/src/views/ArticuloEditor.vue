<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { createArticle, getArticle, listArticles, updateArticle, type Article } from '@/api/articles'
import ImageUpload from '@/components/ImageUpload.vue'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const isEdit = computed(() => typeof route.params.slug === 'string')
const error = ref('')
const saving = ref(false)
const categories = ref<string[]>([])

const form = ref({
  title: '',
  slug: '',
  date: '',
  description: '',
  tone: '',
  category: 'general',
  image: '',
  imageAlt: '',
  body: '',
})

const bodyParagraphs = computed(() => form.value.body.split(/\n\s*\n/).filter(Boolean))
const wordCount = computed(() => form.value.body.trim().split(/\s+/).filter(Boolean).length)
const canSave = computed(() => !!form.value.title.trim() && !!form.value.body.trim() && !saving.value)
// El alt no se pide: se usa el título, que es lo que el autor ya escribió.
const imageAltForUpload = computed(
  () => form.value.imageAlt?.trim() || form.value.title.trim() || 'Imagen de la noticia',
)

function onImagePathUpdate(path: string) {
  form.value.image = path
}

function onImageAltUpdate(alt: string) {
  form.value.imageAlt = alt
}

onMounted(async () => {
  listArticles()
    .then((data) => {
      categories.value = [...new Set(data.items.map((i) => i.category).filter(Boolean))].sort()
    })
    .catch(() => {})

  if (!isEdit.value) return
  try {
    const item = await getArticle(String(route.params.slug))
    form.value = {
      title: item.title,
      slug: item.slug,
      date: item.date ?? '',
      description: item.description,
      tone: item.tone ?? '',
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
  if (!form.value.title.trim()) {
    error.value = 'Escribe el título del artículo'
    return
  }
  if (!form.value.body.trim()) {
    error.value = 'Escribe el cuerpo del artículo'
    return
  }
  saving.value = true
  try {
    // `date` va solo si el artículo ya la tiene: el backend no acepta cadena vacía
    // y, al editar, reenviarla evita que la publicación cambie de fecha.
    const payload: Article = { ...form.value }
    if (payload.image && !payload.imageAlt?.trim()) payload.imageAlt = payload.title.trim()
    if (!payload.date) delete payload.date
    if (isEdit.value) {
      await updateArticle(String(route.params.slug), payload, auth.csrfToken)
    } else {
      delete payload.date
      await createArticle({ ...payload, slug: '' }, auth.csrfToken)
    }
    await router.push('/articulos')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'No se pudo guardar'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <main class="min-h-screen p-6">
    <h1 class="mb-6 text-2xl font-semibold">{{ isEdit ? 'Editar artículo' : 'Nuevo artículo' }}</h1>
    <div class="grid max-w-2xl gap-4">
      <div class="grid gap-1">
        <label class="text-sm text-neutral-600" for="titulo">Título</label>
        <input
          id="titulo"
          v-model="form.title"
          placeholder="Título de la noticia"
          class="rounded-lg border border-neutral-300 px-3 py-2 text-lg"
        >
      </div>

      <div class="grid gap-1">
        <label class="text-sm text-neutral-600" for="cuerpo">Cuerpo de la noticia</label>
        <textarea
          id="cuerpo"
          v-model="form.body"
          rows="16"
          placeholder="Escribe o pega aquí la noticia completa.&#10;&#10;Deja una línea en blanco entre párrafos."
          class="rounded-lg border border-neutral-300 px-3 py-2 leading-relaxed"
        />
        <p class="text-xs text-neutral-500">
          {{ wordCount }} palabras · una línea en blanco separa párrafos · puedes usar
          <code>**negrita**</code> y <code>## subtítulo</code>
        </p>
      </div>

      <ImageUpload
        hide-alt-field
        :image-path="form.image"
        :image-alt="imageAltForUpload"
        @update:imagePath="onImagePathUpdate"
        @update:imageAlt="onImageAltUpdate"
      />

      <div class="grid gap-1">
        <label class="text-sm text-neutral-600" for="categoria">Categoría</label>
        <input
          id="categoria"
          v-model="form.category"
          list="categorias-existentes"
          placeholder="general"
          class="rounded-lg border border-neutral-300 px-3 py-2"
        >
        <datalist id="categorias-existentes">
          <option v-for="c in categories" :key="c" :value="c" />
        </datalist>
      </div>

      <p v-if="error" class="text-sm text-red-600">{{ error }}</p>

      <div class="flex flex-wrap gap-3">
        <button
          class="rounded-lg bg-neutral-900 px-4 py-2 text-white disabled:opacity-50"
          :disabled="!canSave"
          @click="onSave"
        >
          {{ saving ? 'Guardando…' : 'Publicar' }}
        </button>
        <router-link class="rounded-lg border border-neutral-300 px-4 py-2" to="/articulos">
          Cancelar
        </router-link>
      </div>

      <article
        v-if="form.body.trim()"
        class="mt-4 rounded-xl border border-neutral-200 bg-white p-5"
      >
        <p class="mb-1 text-xs uppercase tracking-wide text-neutral-500">
          Vista previa · {{ wordCount }} palabras · categoría: {{ form.category || '—' }}
        </p>
        <h2 class="mb-2 text-xl font-semibold">{{ form.title || '(sin título)' }}</h2>
        <p v-for="(p, i) in bodyParagraphs" :key="i" class="mb-3 leading-relaxed text-neutral-800">
          {{ p }}
        </p>
      </article>
    </div>
  </main>
</template>
