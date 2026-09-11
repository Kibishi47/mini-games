import { defineStore } from 'pinia'

export type SeatStatus = 'active' | 'folded' | 'all_in'

export interface PokerSeatView {
  user_id: string
  nickname: string
  mascot: string
  color: string
  chips: number
  bet: number
  status: SeatStatus
  is_dealer: boolean
  is_sb: boolean
  is_bb: boolean
  is_turn: boolean
  hole_cards?: string[]
}

export interface PokerTableState {
  hand_num: number
  street: 'preflop' | 'flop' | 'turn' | 'river' | 'showdown'
  community: string[]
  pot: number
  current_bet: number
  min_raise: number
  call_amount: number
  to_act: string
  action_ends_at: string
  small_blind: number
  big_blind: number
  is_completed: boolean
  seats: PokerSeatView[]
}

export interface PokerHandEndPlayer {
  user_id: string
  nickname: string
  folded: boolean
  winnings: number
  hole_cards?: string[]
  category?: string
}

export interface PokerHandEndResult {
  hand_num: number
  community: string[]
  players?: PokerHandEndPlayer[]
  uncontested?: boolean
  winner_id?: string
  winner_name?: string
  amount?: number
}

export const usePokerStore = defineStore('poker', {
  state: () => ({
    table: null as PokerTableState | null,
    lastHandEnd: null as PokerHandEndResult | null,
    showHandEnd: false,
  }),

  actions: {
    syncState(payload: PokerTableState) {
      this.table = payload
      // Une nouvelle main a démarré : masquer le bandeau de fin de main précédente
      if (!payload.is_completed) {
        this.showHandEnd = false
      }
    },

    handleHandEnd(payload: PokerHandEndResult) {
      this.lastHandEnd = payload
      this.showHandEnd = true
    },

    dismissHandEnd() {
      this.showHandEnd = false
    },

    reset() {
      this.table = null
      this.lastHandEnd = null
      this.showHandEnd = false
    },
  },
})
