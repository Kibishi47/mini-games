<template>
  <div class="flex flex-col items-center select-none w-full max-w-lg mx-auto">
    <!-- Grille de jeu -->
    <div class="grid gap-2.5 mb-8 w-full">
      <div
        v-for="(row, rIdx) in rows"
        :key="rIdx"
        class="flex justify-center gap-2.5"
      >
        <div
          v-for="(tile, cIdx) in row"
          :key="cIdx"
          :class="[
            'w-14 h-14 sm:w-16 sm:h-16 flex items-center justify-center font-black text-2xl rounded-2xl border-2 transition-all duration-300 transform',
            getTileClass(tile, rIdx === activeRowIndex && !!tile.letter)
          ]"
        >
          <span :class="{ 'animate-pop': rIdx === activeRowIndex && !!tile.letter }">
            {{ tile.letter }}
          </span>
        </div>
      </div>
    </div>

    <!-- Clavier Virtuel réactif -->
    <div class="w-full space-y-2">
      <div
        v-for="(keyboardRow, idx) in keyboardRows"
        :key="idx"
        class="flex justify-center gap-1.5"
      >
        <button
          v-for="key in keyboardRow"
          :key="key"
          @click="handleKeyPress(key)"
          :class="[
            'flex items-center justify-center font-bold text-sm sm:text-base rounded-xl transition-all duration-150 active:scale-95 shadow-md',
            key.length > 1 ? 'px-3 sm:px-4 h-12 bg-slate-800 text-slate-200 hover:bg-slate-700' : 'w-10 sm:w-11 h-12',
            getKeyClass(key)
          ]"
        >
          <span v-if="key !== 'DEL'">{{ key }}</span>
          <Delete v-else class="w-5 h-5 text-slate-300" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { Delete } from 'lucide-vue-next'
import { useGameStore, type TileEvaluation } from '~/stores/game'

const props = defineProps<{
  onSubmit: (guess: string) => void
}>()

const gameStore = useGameStore()

const keyboardRows = [
  ['A', 'Z', 'E', 'R', 'T', 'Y', 'U', 'I', 'O', 'P'],
  ['Q', 'S', 'D', 'F', 'G', 'H', 'J', 'K', 'L', 'M'],
  ['ENTER', 'W', 'X', 'C', 'V', 'B', 'N', 'DEL'],
]

const activeRowIndex = computed(() => gameStore.myAttempts.length)

// Construction réactive des lignes (historique + ligne courante + lignes vides)
const rows = computed(() => {
  const result: TileEvaluation[][] = []
  const wordLen = gameStore.wordLength
  const maxAttempts = gameStore.maxAttempts

  // 1. Tentatives passées
  for (const attempt of gameStore.myAttempts) {
    result.push(attempt)
  }

  // 2. Ligne courante en cours de frappe
  if (result.length < maxAttempts && !gameStore.isFinished) {
    const currentRow: TileEvaluation[] = []
    const inputChars = gameStore.currentInput.split('')
    for (let i = 0; i < wordLen; i++) {
      currentRow.push({
        letter: inputChars[i] || '',
        status: 'tbd',
      })
    }
    result.push(currentRow)
  }

  // 3. Lignes vides restantes
  while (result.length < maxAttempts) {
    const emptyRow: TileEvaluation[] = []
    for (let i = 0; i < wordLen; i++) {
      emptyRow.push({ letter: '', status: 'empty' })
    }
    result.push(emptyRow)
  }

  return result
})

const getTileClass = (tile: TileEvaluation, isTyping: boolean) => {
  if (tile.status === 'correct') {
    return 'bg-emerald-600 border-emerald-500 text-white shadow-lg shadow-emerald-600/30'
  }
  if (tile.status === 'present') {
    return 'bg-amber-500 border-amber-400 text-white shadow-lg shadow-amber-500/30'
  }
  if (tile.status === 'absent') {
    return 'bg-slate-800 border-slate-700 text-slate-400'
  }
  if (isTyping) {
    return 'bg-brand-surface/90 border-indigo-400 text-white scale-105 shadow-md shadow-indigo-500/20'
  }
  return 'bg-brand-surface/40 border-white/10 text-white'
}

const getKeyClass = (key: string) => {
  if (key === 'ENTER' || key === 'DEL') return ''
  const status = gameStore.keyboardStatus[key]
  if (status === 'correct') return 'bg-emerald-600 text-white hover:bg-emerald-500'
  if (status === 'present') return 'bg-amber-500 text-white hover:bg-amber-400'
  if (status === 'absent') return 'bg-slate-900 text-slate-500 cursor-not-allowed opacity-50'
  return 'bg-slate-800 text-white hover:bg-slate-700'
}

const handleKeyPress = (key: string) => {
  if (gameStore.isFinished) return

  if (key === 'DEL' || key === 'BACKSPACE') {
    if (gameStore.currentInput.length > 0) {
      gameStore.currentInput = gameStore.currentInput.slice(0, -1)
    }
  } else if (key === 'ENTER') {
    if (gameStore.currentInput.length === gameStore.wordLength) {
      props.onSubmit(gameStore.currentInput)
      gameStore.currentInput = ''
    }
  } else if (/^[A-Z]$/i.test(key) && gameStore.currentInput.length < gameStore.wordLength) {
    gameStore.currentInput += key.toUpperCase()
  }
}

// Écoute des touches du clavier physique desktop
const onKeyDown = (e: KeyboardEvent) => {
  if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
    return // Laisser le chat tranquille
  }
  if (e.key === 'Backspace') {
    handleKeyPress('DEL')
  } else if (e.key === 'Enter') {
    handleKeyPress('ENTER')
  } else if (/^[a-zA-Z]$/.test(e.key)) {
    handleKeyPress(e.key.toUpperCase())
  }
}

onMounted(() => {
  window.addEventListener('keydown', onKeyDown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeyDown)
})
</script>
