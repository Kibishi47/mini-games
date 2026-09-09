import { defineStore } from 'pinia'

export interface User {
  id: string
  username: string
  display_username: string
  email?: string
  avatar_url: string
  is_guest: boolean
  created_at: string
}

export interface UserStats {
  user_id: string
  game_type: string
  games_played: number
  games_won: number
  win_streak: number
  highest_score: number
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null as User | null,
    stats: null as UserStats | null,
    token: null as string | null,
    isLoading: false,
  }),

  getters: {
    isAuthenticated: (state) => !!state.token,
    isGuest: (state) => state.user?.is_guest ?? true,
    displayName: (state) => state.user?.display_username || state.user?.username || 'Joueur',
    avatar: (state) => state.user?.avatar_url || 'https://api.dicebear.com/7.x/bottts/svg?seed=default',
  },

  actions: {
    initAuth() {
      if (process.client) {
        const savedToken = localStorage.getItem('mg_token')
        const savedUser = localStorage.getItem('mg_user')
        if (savedToken) {
          this.token = savedToken
        }
        if (savedUser) {
          try {
            this.user = JSON.parse(savedUser)
          } catch (e) {
            // ignore
          }
        }
        if (this.token) {
          this.fetchMe()
        }
      }
    },

    setSession(token: string, user: User) {
      this.token = token
      this.user = user
      if (process.client) {
        localStorage.setItem('mg_token', token)
        localStorage.setItem('mg_user', JSON.stringify(user))
      }
    },

    logout() {
      this.token = null
      this.user = null
      this.stats = null
      if (process.client) {
        localStorage.removeItem('mg_token')
        localStorage.removeItem('mg_user')
      }
    },

    async fetchMe() {
      if (!this.token) return
      const config = useRuntimeConfig()
      try {
        const res = await $fetch<{ user: User; stats: UserStats }>(`${config.public.apiUrl}/api/users/me`, {
          headers: { Authorization: `Bearer ${this.token}` },
        })
        this.user = res.user
        this.stats = res.stats
        if (process.client) {
          localStorage.setItem('mg_user', JSON.stringify(res.user))
        }
      } catch (err) {
        this.logout()
      }
    },

    async loginGuest(customName?: string) {
      this.isLoading = true
      const config = useRuntimeConfig()
      try {
        const res = await $fetch<{ token: string; user: User }>(`${config.public.apiUrl}/api/auth/guest`, {
          method: 'POST',
          body: { custom_name: customName || '' },
        })
        this.setSession(res.token, res.user)
        return res
      } finally {
        this.isLoading = false
      }
    },

    async loginLocal(username: string, password: string) {
      this.isLoading = true
      const config = useRuntimeConfig()
      try {
        const res = await $fetch<{ token: string; user: User }>(`${config.public.apiUrl}/api/auth/login`, {
          method: 'POST',
          body: { username, password },
        })
        this.setSession(res.token, res.user)
        return res
      } finally {
        this.isLoading = false
      }
    },

    async registerLocal(username: string, display: string, email: string, password: string) {
      this.isLoading = true
      const config = useRuntimeConfig()
      try {
        const res = await $fetch<{ token: string; user: User }>(`${config.public.apiUrl}/api/auth/register`, {
          method: 'POST',
          body: { username, display_username: display, email, password },
        })
        this.setSession(res.token, res.user)
        return res
      } finally {
        this.isLoading = false
      }
    },

    async updateProfile(displayUsername: string, avatarURL: string) {
      if (!this.token) return
      const config = useRuntimeConfig()
      const res = await $fetch<User>(`${config.public.apiUrl}/api/users/profile`, {
        method: 'PUT',
        headers: { Authorization: `Bearer ${this.token}` },
        body: { display_username: displayUsername, avatar_url: avatarURL },
      })
      this.user = res
      if (process.client) {
        localStorage.setItem('mg_user', JSON.stringify(res))
      }
    },
  },
})
