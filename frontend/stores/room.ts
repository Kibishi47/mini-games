import { defineStore } from 'pinia'
import { useProfileStore } from '~/stores/profile'

export type RoomStatus = 'in_lobby' | 'in_game' | 'closed'
export type PlayerRole = 'master' | 'player' | 'spectator'

export interface RoomPlayer {
  id: string
  nickname: string
  mascot: string
  color: string
  role: PlayerRole
  is_master: boolean
  is_spectator: boolean
  is_muted: boolean
  is_connected: boolean
  score: number
  location?: 'lobby' | 'in_game'
}

export interface RoomSettings {
  game_type: string
  word_length: number
  round_duration: number
  max_rounds: number
  max_attempts: number
  language: string
}

export type RoundSubState = 'playing' | 'round_ended' | 'game_over'

export interface Room {
  code: string
  status: RoomStatus
  round_state?: RoundSubState
  master_id: string
  settings: RoomSettings
  current_round: number
  ends_at?: string
  next_round_at?: string
  revealed_word?: string
  players: RoomPlayer[]
}

export interface ChatMessage {
  id: string
  sender_id: string
  sender: string
  mascot: string
  color: string
  content: string
  is_system: boolean
  created_at: string
}

export const useRoomStore = defineStore('room', {
  state: () => ({
    currentRoom: null as Room | null,
    chatMessages: [] as ChatMessage[],
  }),

  getters: {
    players: (state) => state.currentRoom?.players || [],
    activePlayers: (state) => (state.currentRoom?.players || []).filter(p => !p.is_spectator),
    spectators: (state) => (state.currentRoom?.players || []).filter(p => p.is_spectator),
    
    isMaster: (state) => {
      const profile = useProfileStore()
      const me = state.currentRoom?.players.find(p => p.nickname === profile.nickname)
      return !!me?.is_master
    },

    me: (state) => {
      const profile = useProfileStore()
      return state.currentRoom?.players.find(p => p.nickname === profile.nickname)
    },

    masterPlayer: (state) => state.currentRoom?.players.find(p => p.is_master),
  },

  actions: {
    setRoom(room: Room) {
      this.currentRoom = room
    },

    addChatMessage(msg: ChatMessage) {
      this.chatMessages.push(msg)
      if (this.chatMessages.length > 50) {
        this.chatMessages.shift()
      }
    },

    setChatHistory(messages: ChatMessage[]) {
      this.chatMessages = messages
    },

    clearRoom() {
      this.currentRoom = null
      this.chatMessages = []
    },

    updatePlayerLocation(userId: string, location: 'lobby' | 'in_game') {
      if (!this.currentRoom?.players) return
      const player = this.currentRoom.players.find(p => p.id === userId)
      if (player) {
        player.location = location
      }
    },

    resetPlayerScores() {
      if (!this.currentRoom?.players) return
      this.currentRoom.players.forEach(p => {
        p.score = 0
      })
    },
  },
})
