import { apiFetch } from '@/api/client'

export type Settings = {
  siteName: string
  siteDescription: string
  siteUrl: string
  defaultOgImage: string
  twitterHandle: string
  adsense: {
    enabled: boolean
    clientId: string
    slot1Id: string
    slot1Enabled: boolean
    slot2Id: string
    slot2Enabled: boolean
  }
}

export function getSettings() {
  return apiFetch<Settings>('/admin/api/ajustes')
}

export function updateSettings(payload: Settings, csrfToken: string) {
  return apiFetch<Settings>('/admin/api/ajustes', {
    method: 'PUT',
    headers: { 'X-CSRF-Token': csrfToken },
    body: JSON.stringify(payload),
  })
}

