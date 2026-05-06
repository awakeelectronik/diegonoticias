import { apiFetch } from '@/api/client'

export type Banner = {
  id: string
  title: string
  imagePath: string
  linkUrl: string
  active: boolean
  slot: number
}

export function listAds() {
  return apiFetch<{ banners: Banner[] }>('/admin/api/publicidad')
}

export function createAd(payload: Omit<Banner, 'id'>, csrfToken: string) {
  return apiFetch<Banner>('/admin/api/publicidad', {
    method: 'POST',
    headers: { 'X-CSRF-Token': csrfToken },
    body: JSON.stringify(payload),
  })
}

export function updateAd(id: string, payload: Omit<Banner, 'id'>, csrfToken: string) {
  return apiFetch<Banner>(`/admin/api/publicidad/${id}`, {
    method: 'PUT',
    headers: { 'X-CSRF-Token': csrfToken },
    body: JSON.stringify(payload),
  })
}

export function deleteAd(id: string, csrfToken: string) {
  return apiFetch<{ ok: boolean }>(`/admin/api/publicidad/${id}`, {
    method: 'DELETE',
    headers: { 'X-CSRF-Token': csrfToken },
  })
}

