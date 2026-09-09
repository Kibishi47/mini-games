<template>
  <div class="flex flex-col items-center select-none w-full max-w-xl mx-auto">
    <!-- Grille de jeu Principale -->
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
            'w-14 h-14 sm:w-16 sm:h-16 flex items-center justify-center font-condensed font-black text-3xl rounded-xl border-[3.5px] border-ink-black transition-none select-none',
            getTileClass(tile, rIdx === activeRowIndex && !!tile.letter)
          ]"
        >
          <span>{{ tile.letter }}</span>
        </div>
      </div>
    </div>

    <!-- Clavier Virtuel Façon Touches de Scrabble / Machine à écrire -->
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
            'flex items-center justify-center font-display font-black text-sm sm:text-base rounded-xl border-[2.5px] border-ink-black shadow-pop-xs transition-none active:translate-x-[2px] active:translate-y-[2px] active:shadow-none select-none',
            key.length > 1 ? 'px-3 sm:px-4 h-12 bg-game-blue text-board-white' : 'w-10 sm:w-11 h-12',
            getKeyClass(key)
          ]"
        >
          <span v-if="key !== 'DEL'">{{ key }}</span>
          <Delete v-else class="w-5 h-5 text-ink-black" />
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

// Construction réactive des lignes
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
    return 'bg-game-green text-board-white shadow-pop-xs'
  }
  if (tile.status === 'present') {
    return 'bg-game-yellow text-ink-black shadow-pop-xs'
  }
  if (tile.status === 'absent') {
    return 'pattern-hatch text-ink-black/40'
  }
  if (isTyping) {
    return 'bg-board-white text-ink-black shadow-pop-sm'
  }
  return 'bg-board-white text-ink-black'
}

const getKeyClass = (key: string) => {
  if (key === 'ENTER' || key === 'DEL') return ''
  const status = gameStore.keyboardStatus[key]
  if (status === 'correct') return 'bg-game-green text-board-white'
  if (status === 'present') return 'bg-game-yellow text-ink-black'
  if (status === 'absent') return 'pattern-hatch text-ink-black/40 cursor-not-allowed opacity-60 shadow-none'
  return 'bg-board-white text-ink-black hover:bg-board-cream'
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

const onKeyDown = (e: KeyboardEvent) => {
  if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
    return
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
