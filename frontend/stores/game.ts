import { defineStore } from 'pinia'

export type TileStatus = 'correct' | 'present' | 'absent' | 'empty' | 'tbd'

export interface TileEvaluation {
  letter: string
  status: TileStatus
}

export interface MaskedTile {
  status: TileStatus
}

export interface OpponentProgress {
  user_id: string
  display_username: string
  masked_rows: MaskedTile[][]
  is_solved: boolean
  is_finished: boolean
}

export interface PlayerSummary {
  user_id: string
  display_username: string
  avatar_url: string
  is_solved: boolean
  attempts_count: number
  score_delta: number
  total_score: number
  emoji_grid: string
}

export const useGameStore = defineStore('game', {
  state: () => ({
    currentRound: 1,
    maxRounds: 3,
    wordLength: 5,
    maxAttempts: 6,
    endsAt: null as Date | null,
    roundDuration: 60,
    
    // Grille active du joueur
    myAttempts: [] as TileEvaluation[][],
    currentInput: '',
    
    // Adversaires
    opponents: [] as OpponentProgress[],
    
    // Statut personnel
    isSolved: false,
    isFinished: false,
    myScore: 0,

    // Fin de manche
    targetWord: '',
    roundSummaries: [] as PlayerSummary[],
    isGameOver: false,
    showRoundSummary: false,
  }),

  getters: {
    isSpectator: () => {
      const room = useRoomStore()
      return room.myRole === 'spectator'
    },
    
    keyboardStatus: (state) => {
      const statusMap: Record<string, TileStatus> = {}
      for (const row of state.myAttempts) {
        for (const tile of row) {
          const char = tile.letter.toUpperCase()
          const current = statusMap[char]
          if (tile.status === 'correct') {
            statusMap[char] = 'correct'
          } else if (tile.status === 'present' && current !== 'correct') {
            statusMap[char] = 'present'
          } else if (tile.status === 'absent' && !current) {
            statusMap[char] = 'absent'
          }
        }
      }
      return statusMap
    },
  },

  actions: {
    startNewRound(payload: {
      round: number
      max_rounds: number
      word_length: number
      max_attempts: number
      ends_at: string
      round_duration: number
    }) {
      this.currentRound = payload.round
      this.maxRounds = payload.max_rounds
      this.wordLength = payload.word_length
      this.maxAttempts = payload.max_attempts
      this.endsAt = new Date(payload.ends_at)
      this.roundDuration = payload.round_duration

      this.myAttempts = []
      this.currentInput = ''
      this.opponents = []
      this.isSolved = false
      this.isFinished = false
      this.targetWord = ''
      this.roundSummaries = []
      this.showRoundSummary = false
    },

    updateOpponentProgress(payload: {
      user_id: string
      row_index: number
      masked_row: MaskedTile[]
      is_solved: boolean
      is_finished: boolean
    }) {
      const opp = this.opponents.find(o => o.user_id === payload.user_id)
      if (opp) {
        opp.masked_rows[payload.row_index] = payload.masked_row
        opp.is_solved = payload.is_solved
        opp.is_finished = payload.is_finished
      } else {
        const room = useRoomStore()
        const player = room.players.find(p => p.user_id === payload.user_id)
        const newOpp: OpponentProgress = {
          user_id: payload.user_id,
          display_username: player?.display_username || 'Adversaire',
          masked_rows: [payload.masked_row],
          is_solved: payload.is_solved,
          is_finished: payload.is_finished,
        }
        this.opponents.push(newOpp)
      }
    },

    setRoundEnd(payload: {
      target_word: string
      round: number
      max_rounds: number
      summaries: PlayerSummary[]
      is_game_over: boolean
    }) {
      this.targetWord = payload.target_word
      this.roundSummaries = payload.summaries
      this.isGameOver = payload.is_game_over
      this.showRoundSummary = true
    },

    resetGame() {
      this.myAttempts = []
      this.currentInput = ''
      this.opponents = []
      this.isSolved = false
      this.isFinished = false
      this.targetWord = ''
      this.roundSummaries = []
      this.showRoundSummary = false
      this.isGameOver = false
    },
  },
})
