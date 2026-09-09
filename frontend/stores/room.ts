import { defineStore } from 'pinia'

export type RoomStatus = 'in_lobby' | 'in_game' | 'closed'
export type PlayerRole = 'master' | 'player' | 'spectator'

export interface RoomPlayer {
  user_id: string
  username: string
  display_username: string
  avatar_url: string
  role: PlayerRole
  is_muted: boolean
  is_connected: boolean
  score: number
}

export interface RoomSettings {
  game_type: string
  word_length: number
  round_duration: number
  max_rounds: number
  max_attempts: number
  language: string
}

export interface Room {
  code: string
  status: RoomStatus
  master_id: string
  settings: RoomSettings
  current_round: number
  ends_at?: string
  players: RoomPlayer[]
}

export interface ChatMessage {
  id: string
  sender_id: string
  sender: string
  avatar_url: string
  content: string
  is_system: boolean
  created_at: string
}

export const useRoomStore = defineStore('room', {
  state: () => ({
    currentRoom: null as Room | null,
    chatMessages: [] as ChatMessage[],
    myRole: 'player' as PlayerRole,
  }),

  getters: {
    isMaster: (state) => {
      const auth = useAuthStore()
      return state.currentRoom?.master_id === auth.user?.id
    },
    players: (state) => state.currentRoom?.players || [],
    activePlayers: (state) => (state.currentRoom?.players || []).filter(p => p.role !== 'spectator'),
    spectators: (state) => (state.currentRoom?.players || []).filter(p => p.role === 'spectator'),
    settings: (state) => state.currentRoom?.settings || {
      game_type: 'wordle',
      word_length: 5,
      round_duration: 60,
      max_rounds: 3,
      max_attempts: 6,
      language: 'fr',
    },
  },

  actions: {
    setRoom(room: Room) {
      this.currentRoom = room
      const auth = useAuthStore()
      const me = room.players.find(p => p.user_id === auth.user?.id)
      if (me) {
        this.myRole = me.role
      }
    },

    setChat(messages: ChatMessage[]) {
      this.chatMessages = messages
    },

    addChatMessage(msg: ChatMessage) {
      this.chatMessages.push(msg)
      if (this.chatMessages.length > 50) {
        this.chatMessages.shift()
      }
    },

    clearRoom() {
      this.currentRoom = null
      this.chatMessages = []
    },
  },
})
