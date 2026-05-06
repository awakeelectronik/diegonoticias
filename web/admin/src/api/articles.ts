import { apiFetch } from '@/api/client'

export type Article = {
  title: string
  slug: string
  description: string
  tone: string
  category: string
  body: string
}

export function listArticles() {
  return apiFetch<{ items: Article[] }>('/admin/api/articulos')
}

export function getArticle(slug: string) {
  return apiFetch<Article>(`/admin/api/articulos/${slug}`)
}

export function createArticle(article: Article, csrfToken: string) {
  return apiFetch<Article>('/admin/api/articulos', {
    method: 'POST',
    headers: { 'X-CSRF-Token': csrfToken },
    body: JSON.stringify(article),
  })
}

export function updateArticle(slug: string, article: Article, csrfToken: string) {
  return apiFetch<Article>(`/admin/api/articulos/${slug}`, {
    method: 'PUT',
    headers: { 'X-CSRF-Token': csrfToken },
    body: JSON.stringify(article),
  })
}

export function deleteArticle(slug: string, csrfToken: string) {
  return apiFetch<{ ok: boolean }>(`/admin/api/articulos/${slug}`, {
    method: 'DELETE',
    headers: { 'X-CSRF-Token': csrfToken },
  })
}

