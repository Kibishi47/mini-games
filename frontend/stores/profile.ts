import { defineStore } from 'pinia'
import type { MascotName } from '~/components/ui/GameMascot.vue'

export const MASCOTS: { id: MascotName; name: string; color: string; desc: string }[] = [
  { id: 'dice', name: 'Le Dé Sprinteur', color: '#FFD300', desc: 'Toujours en mouvement' },
  { id: 'domino', name: 'Le Domino Rusé', color: '#158A44', desc: 'Calculateur impassible' },
  { id: 'card', name: 'L\'As de Cœur', color: '#E63228', desc: 'Chaud bouillant' },
  { id: 'knight', name: 'Le Cavalier', color: '#121212', desc: 'Fonce tête baissée' },
  { id: 'd20', name: 'Le D20 Cosmique', color: '#F27A9B', desc: 'Maître du hasard critique' },
  { id: 'meeple', name: 'Le Meeple Arbitre', color: '#1D4ED8', desc: 'L\'esprit du jeu' },
]

export const useProfileStore = defineStore('profile', {
  state: () => {
    let nickname = ''
    let mascot: MascotName | '' = ''
    let color = '#FFD300'
    let sessionToken = ''
    let currentRoomCode = ''
    let isReady = false

    if (process.client) {
      try {
        const savedNick = localStorage.getItem('mg_nickname')
        const savedMascot = localStorage.getItem('mg_mascot') as MascotName | null
        const savedColor = localStorage.getItem('mg_color')
        const savedToken = localStorage.getItem('mg_session_token')
        const savedRoom = localStorage.getItem('mg_room_code')

        if (savedNick) nickname = savedNick
        if (savedMascot && MASCOTS.some(m => m.id === savedMascot)) {
          mascot = savedMascot
        } else if (savedMascot) {
          mascot = 'dice'
        }
        if (savedColor) color = savedColor
        if (savedToken) {
          sessionToken = savedToken
        } else {
          sessionToken = crypto.randomUUID ? crypto.randomUUID() : 'sess_' + Math.random().toString(36).substring(2, 15)
          localStorage.setItem('mg_session_token', sessionToken)
        }
        if (savedRoom) currentRoomCode = savedRoom
        isReady = true
      } catch (e) {
        // LocalStorage non accessible
      }
    }

    return {
      nickname: nickname || 'Joueur',
      mascot: mascot as MascotName | '',
      color,
      sessionToken,
      currentRoomCode,
      isHydrated: isReady,
      isReady,
    }
  },

  actions: {
    initProfile() {
      if (process.client) {
        try {
          const savedNick = localStorage.getItem('mg_nickname')
          const savedMascot = localStorage.getItem('mg_mascot') as MascotName | null
          const savedColor = localStorage.getItem('mg_color')
          let savedToken = localStorage.getItem('mg_session_token')
          const savedRoom = localStorage.getItem('mg_room_code')

          this.nickname = savedNick || this.nickname || 'Joueur'
          if (savedMascot && MASCOTS.some(m => m.id === savedMascot)) {
            this.mascot = savedMascot
          } else if (!this.mascot) {
            this.mascot = 'dice'
          }
          if (savedColor) this.color = savedColor
          if (!savedToken) {
            savedToken = crypto.randomUUID ? crypto.randomUUID() : 'sess_' + Math.random().toString(36).substring(2, 15)
            localStorage.setItem('mg_session_token', savedToken)
          }
          this.sessionToken = savedToken
          if (savedRoom) this.currentRoomCode = savedRoom
          this.isHydrated = true
          this.isReady = true
        } catch (e) {
          // ignore
        }
      }
    },

    setProfile(nickname: string, mascot: MascotName, color?: string) {
      this.nickname = nickname.trim() || 'Joueur'
      this.mascot = mascot
      if (color) {
        this.color = color
      } else {
        const found = MASCOTS.find(m => m.id === mascot)
        if (found) this.color = found.color
      }

      if (process.client) {
        localStorage.setItem('mg_nickname', this.nickname)
        localStorage.setItem('mg_mascot', this.mascot)
        localStorage.setItem('mg_color', this.color)
      }
    },

    setSession(token: string, roomCode: string) {
      this.sessionToken = token
      this.currentRoomCode = roomCode
      if (process.client) {
        localStorage.setItem('mg_session_token', token)
        localStorage.setItem('mg_room_code', roomCode)
      }
    },

    clearSession() {
      this.sessionToken = ''
      this.currentRoomCode = ''
      if (process.client) {
        localStorage.removeItem('mg_session_token')
        localStorage.removeItem('mg_room_code')
      }
    },
  },
})
