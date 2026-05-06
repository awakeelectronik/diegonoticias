import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import LoginView from '@/views/Login.vue'
import ArticulosView from '@/views/Articulos.vue'
import ArticuloEditorView from '@/views/ArticuloEditor.vue'
import AjustesView from '@/views/Ajustes.vue'
import PublicidadView from '@/views/Publicidad.vue'
import PublicidadEditorView from '@/views/PublicidadEditor.vue'

export const router = createRouter({
  history: createWebHistory('/admin/'),
  routes: [
    { path: '/login', component: LoginView },
    { path: '/articulos', component: ArticulosView },
    { path: '/articulos/nuevo', component: ArticuloEditorView },
    { path: '/articulos/:slug/editar', component: ArticuloEditorView },
    { path: '/publicidad', component: PublicidadView },
    { path: '/publicidad/nueva', component: PublicidadEditorView },
    { path: '/publicidad/:id/editar', component: PublicidadEditorView },
    { path: '/ajustes', component: AjustesView },
    { path: '/', redirect: '/articulos' },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (!auth.checked) {
    await auth.fetchMe()
  }
  if (to.path !== '/login' && !auth.isAuthenticated) {
    return '/login'
  }
  if (to.path === '/login' && auth.isAuthenticated) {
    return '/articulos'
  }
  return true
})

