import { apiFetch } from '@/api/client'

export type MeResponse = { username: string; csrfToken: string }

export function login(username: string, password: string) {
  return apiFetch<MeResponse>('/admin/api/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  })
}

export function me() {
  return apiFetch<MeResponse>('/admin/api/me')
}

export function logout(csrfToken: string) {
  return apiFetch<{ ok: string }>('/admin/api/logout', {
    method: 'POST',
    headers: { 'X-CSRF-Token': csrfToken },
  })
}

