<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const username = ref('diego')
const password = ref('')
const error = ref('')
const auth = useAuthStore()
const router = useRouter()

async function onSubmit() {
  error.value = ''
  try {
    await auth.login(username.value, password.value)
    await router.push('/articulos')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'No se pudo iniciar sesión'
  }
}
</script>

<template>
  <main class="flex min-h-screen items-center justify-center p-6">
    <form class="w-full max-w-sm rounded-2xl bg-white p-6 shadow-sm" @submit.prevent="onSubmit">
      <h1 class="mb-4 text-xl font-semibold">Ingresar al admin</h1>
      <label class="mb-3 block">
        <span class="mb-1 block text-sm text-neutral-600">Usuario</span>
        <input v-model="username" class="w-full rounded-lg border border-neutral-300 px-3 py-2" />
      </label>
      <label class="mb-4 block">
        <span class="mb-1 block text-sm text-neutral-600">Contraseña</span>
        <input v-model="password" type="password" class="w-full rounded-lg border border-neutral-300 px-3 py-2" />
      </label>
      <p v-if="error" class="mb-3 text-sm text-red-600">{{ error }}</p>
      <button
        type="submit"
        :disabled="auth.loading"
        class="w-full rounded-lg bg-neutral-900 px-4 py-2 text-white disabled:opacity-50"
      >
        {{ auth.loading ? 'Entrando…' : 'Entrar' }}
      </button>
    </form>
  </main>
</template>

