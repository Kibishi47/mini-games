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
  nickname: string
  mascot: string
  color: string
  masked_rows: MaskedTile[][]
  is_solved: boolean
  is_finished: boolean
  attempts_cnt: number
}

export interface RoundSummary {
  round: number
  max_rounds?: number
  secret_word: string
  reason: string
  winner_name: string
  round_scores: Record<string, number>
  solve_times?: Record<string, number>
  total_scores: Record<string, number>
  countdown_sec?: number
}

export const useGameStore = defineStore('game', {
  state: () => ({
    currentRound: 1,
    maxRounds: 3,
    wordLength: 5,
    maxAttempts: 6,
    endsAt: null as Date | null,
    roundDuration: 60,

    // Grille du joueur courant
    myAttempts: [] as TileEvaluation[][],
    currentInput: '',

    // Concurrents
    opponents: [] as OpponentProgress[],

    // État personnel
    isSolved: false,
    isFinished: false,
    myRoundScore: 0,

    // Fin de manche / Fin de partie
    targetWord: '',
    roundSummary: null as RoundSummary | null,
    isGameOver: false,
    showRoundSummary: false,
  }),

  getters: {
    keyboardStatus: (state) => {
      const map: Record<string, TileStatus> = {}
      for (const row of state.myAttempts) {
        for (const tile of row) {
          const letter = tile.letter.toUpperCase()
          const current = map[letter]
          if (tile.status === 'correct') {
            map[letter] = 'correct'
          } else if (tile.status === 'present' && current !== 'correct') {
            map[letter] = 'present'
          } else if (tile.status === 'absent' && !current) {
            map[letter] = 'absent'
          }
        }
      }
      return map
    },

    emojiGrid: (state) => {
      return state.myAttempts
        .map(row =>
          row
            .map(tile => {
              if (tile.status === 'correct') return '🟩'
              if (tile.status === 'present') return '🟨'
              return '⬛'
            })
            .join('')
        )
        .join('\n')
    },
  },

  actions: {
    startRound(payload: { round: number; max_rounds: number; word_length: number; max_attempts: number; ends_at: string; round_duration: number }) {
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
      this.myRoundScore = 0
      this.targetWord = ''
      this.showRoundSummary = false
      this.isGameOver = false
    },

    setGuessResult(payload: { attempts: TileEvaluation[][]; is_solved: boolean; is_finished: boolean; round_score: number }) {
      this.myAttempts = payload.attempts
      this.isSolved = payload.is_solved
      this.isFinished = payload.is_finished
      this.myRoundScore = payload.round_score
      this.currentInput = ''
    },

    updateOpponentProgress(payload: OpponentProgress) {
      const idx = this.opponents.findIndex(o => o.user_id === payload.user_id)
      if (idx !== -1) {
        this.opponents[idx] = payload
      } else {
        this.opponents.push(payload)
      }
    },

    endRound(payload: RoundSummary) {
      this.roundSummary = payload
      this.targetWord = payload.secret_word
      this.showRoundSummary = true
      this.isGameOver = payload.round >= this.maxRounds
    },

    syncGameState(payload: any) {
      if (payload.word_length) this.wordLength = payload.word_length
      if (payload.max_attempts) this.maxAttempts = payload.max_attempts
      if (payload.ends_at) this.endsAt = new Date(payload.ends_at)
      if (payload.attempts) this.myAttempts = payload.attempts
      this.isSolved = !!payload.is_solved
      this.isFinished = !!payload.is_finished
      if (payload.opponents) this.opponents = payload.opponents
    },

    typeLetter(letter: string) {
      if (this.isFinished) return
      if (this.currentInput.length < this.wordLength) {
        this.currentInput += letter.toUpperCase()
      }
    },

    deleteLetter() {
      if (this.isFinished) return
      this.currentInput = this.currentInput.slice(0, -1)
    },

    resetGame() {
      this.myAttempts = []
      this.currentInput = ''
      this.opponents = []
      this.isSolved = false
      this.isFinished = false
      this.roundSummary = null
      this.showRoundSummary = false
      this.isGameOver = false
    },
  },
})
