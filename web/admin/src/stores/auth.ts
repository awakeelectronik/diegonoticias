import { defineStore } from 'pinia'
import * as authApi from '@/api/auth'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    username: '',
    csrfToken: '',
    checked: false,
    isAuthenticated: false,
    loading: false,
  }),
  actions: {
    async fetchMe() {
      try {
        const data = await authApi.me()
        this.username = data.username
        this.csrfToken = data.csrfToken
        this.isAuthenticated = true
      } catch {
        this.username = ''
        this.csrfToken = ''
        this.isAuthenticated = false
      } finally {
        this.checked = true
      }
    },
    async login(username: string, password: string) {
      this.loading = true
      try {
        const data = await authApi.login(username, password)
        this.username = data.username
        this.csrfToken = data.csrfToken
        this.isAuthenticated = true
      } finally {
        this.loading = false
      }
    },
    async logout() {
      if (!this.csrfToken) return
      await authApi.logout(this.csrfToken)
      this.username = ''
      this.csrfToken = ''
      this.isAuthenticated = false
    },
  },
})

