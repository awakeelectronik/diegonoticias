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
const justGenerated = ref(false)

const TONES: { id: string; label: string }[] = [
  { id: 'informativo', label: 'Informativo (directo, factual, neutral)' },
  { id: 'profesional', label: 'Profesional (formal pero accesible)' },
  { id: 'institucional', label: 'Institucional (voz oficial)' },
  { id: 'academico', label: 'Académico (preciso, contextual)' },
  { id: 'cronica', label: 'Crónica (narrativo, descriptivo)' },
  { id: 'editorial', label: 'Editorial (opinión argumentada)' },
  { id: 'conversacional', label: 'Conversacional (cercano, simple)' },
  { id: 'pedagogico', label: 'Pedagógico (didáctico para no expertos)' },
  { id: 'dramatico', label: 'Dramático (tensión y urgencia)' },
  { id: 'sensacionalista', label: 'Sensacionalista (impactante, sin inventar)' },
]

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
const bodyParagraphs = computed(() => form.value.body.split(/\n\s*\n/).filter(Boolean))
const wordCount = computed(() => form.value.body.trim().split(/\s+/).filter(Boolean).length)

function onImagePathUpdate(path: string) {
  form.value.image = path
}

function onImageAltUpdate(alt: string) {
  form.value.imageAlt = alt
}

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
  if (!form.value.body.trim()) {
    error.value = 'Genera el contenido con IA antes de guardar'
    return
  }
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
  if (!rawText.value.trim()) {
    error.value = 'Pega el texto crudo antes de generar'
    return
  }
  generating.value = true
  justGenerated.value = false
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
    justGenerated.value = true
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
        v-model="rawText"
        rows="5"
        placeholder="Texto crudo para IA (lo que pegas aquí alimenta a Generar)"
        class="rounded-lg border border-neutral-300 px-3 py-2"
      />
      <label class="text-sm text-neutral-600">Título (opcional, la IA puede proponerlo)</label>
      <input
        v-model="form.title"
        placeholder="Título del artículo"
        class="rounded-lg border border-neutral-300 px-3 py-2"
      />
      <label class="text-sm text-neutral-600">Tono</label>
      <select v-model="form.tone" class="rounded-lg border border-neutral-300 px-3 py-2">
        <option v-for="t in TONES" :key="t.id" :value="t.id">{{ t.label }}</option>
      </select>
      <ImageUpload
        hide-alt-field
        :image-path="form.image"
        :image-alt="imageAltForUpload"
        @update:imagePath="onImagePathUpdate"
        @update:imageAlt="onImageAltUpdate"
      />
      <p v-if="error" class="text-sm text-red-600">{{ error }}</p>

      <div class="flex flex-wrap gap-3">
        <button
          class="rounded-lg bg-[var(--color-accent,#c8553d)] px-4 py-2 text-white disabled:opacity-50"
          :disabled="generating || !rawText.trim()"
          @click="onGenerate"
        >
          {{ generating ? 'Generando…' : 'Generar' }}
        </button>
        <button
          class="rounded-lg bg-neutral-900 px-4 py-2 text-white disabled:opacity-50"
          :disabled="generating || !form.body.trim()"
          @click="onSave"
        >
          Guardar
        </button>
        <router-link class="rounded-lg border border-neutral-300 px-4 py-2" to="/articulos">
          Cancelar
        </router-link>
      </div>

      <article
        v-if="form.body.trim()"
        class="mt-4 rounded-xl border border-neutral-200 bg-white p-5"
        :class="{ 'ring-2 ring-emerald-300': justGenerated }"
      >
        <p class="mb-1 text-xs uppercase tracking-wide text-neutral-500">
          Vista previa · {{ wordCount }} palabras · categoría: {{ form.category || '—' }}
        </p>
        <h2 class="mb-2 text-xl font-semibold">{{ form.title || '(sin título)' }}</h2>
        <p v-if="form.description" class="mb-3 text-sm italic text-neutral-600">
          {{ form.description }}
        </p>
        <p v-for="(p, i) in bodyParagraphs" :key="i" class="mb-3 leading-relaxed text-neutral-800">
          {{ p }}
        </p>
        <p v-if="form.imageAlt" class="mt-3 text-xs text-neutral-500">
          Alt de imagen: "{{ form.imageAlt }}"
        </p>
      </article>
    </div>
  </main>
</template>
