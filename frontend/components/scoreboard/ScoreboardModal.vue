<template>
  <div v-if="gameStore.showRoundSummary" class="fixed inset-0 bg-ink-black/80 backdrop-blur-none flex items-center justify-center z-50 p-4">
    <div class="bg-board-white max-w-2xl w-full p-6 sm:p-8 rounded-3xl border-[4px] border-ink-black shadow-pop-lg relative flex flex-col space-y-6">
      
      <!-- Bouton de fermeture modale (X) -->
      <button
        @click="closeModal"
        class="absolute top-4 right-4 w-9 h-9 flex items-center justify-center rounded-xl border-2 border-ink-black bg-board-cream hover:bg-game-red hover:text-board-white text-ink-black font-black transition-none shadow-pop-xs active:translate-x-[2px] active:translate-y-[2px]"
        title="Fermer la fenêtre (inspecter le plateau / chat)"
      >
        <X class="w-5 h-5" />
      </button>

      <!-- Titre et Mot Secret Révélé sous forme de tuiles Pop Modernistes -->
      <div class="text-center space-y-3">
        <div class="inline-flex items-center space-x-2 border-2 border-ink-black px-3 py-1 rounded-full bg-game-pink font-condensed text-xs uppercase mb-1">
          <Trophy class="w-4 h-4 text-ink-black" />
          <span>{{ gameStore.isGameOver ? 'Palmarès Final' : `Fin Manche ${gameStore.currentRound}/${gameStore.maxRounds}` }}</span>
        </div>

        <h3 class="font-display font-black text-3xl sm:text-4xl uppercase tracking-tight text-ink-black">
          {{ gameStore.isGameOver ? 'Victoire au Sommet !' : 'Tour Terminé !' }}
        </h3>
        
        <p class="font-display font-bold uppercase text-xs text-ink-black/60">Le mot secret était :</p>
        
        <!-- Tuiles pop-modernistes individuelles avec bordure 3px noire -->
        <div class="flex items-center justify-center gap-2 flex-wrap pt-1">
          <div
            v-for="(letter, idx) in targetLetters"
            :key="idx"
            class="w-12 h-12 sm:w-14 sm:h-14 rounded-2xl bg-game-green text-board-white border-[3px] border-ink-black shadow-pop-sm flex items-center justify-center font-condensed font-black text-2xl sm:text-3xl uppercase transform transition-transform hover:-translate-y-1"
          >
            {{ letter }}
          </div>
        </div>
      </div>

      <!-- Piste de Score de Jeu de Société -->
      <div class="space-y-3 max-h-64 overflow-y-auto pr-1">
        <div
          v-for="(player, index) in scoreboardPlayers"
          :key="player.id"
          :class="[
            'flex items-center justify-between p-3.5 rounded-2xl border-[3px] border-ink-black shadow-pop-xs transition-none',
            index === 0 ? 'bg-game-yellow/30' : 'bg-board-cream'
          ]"
        >
          <!-- Rang, Mascotte et Nom -->
          <div class="flex items-center space-x-3">
            <span
              :class="[
                'w-8 h-8 rounded-xl border-2 border-ink-black font-condensed font-black text-sm flex items-center justify-center shadow-pop-xs',
                index === 0 ? 'bg-game-yellow text-ink-black' :
                index === 1 ? 'bg-board-white text-ink-black' :
                index === 2 ? 'bg-game-pink text-ink-black' : 'bg-board-cream text-ink-black/60'
              ]"
            >
              #{{ index + 1 }}
            </span>

            <GameMascot :name="(player.mascot as any) || 'dice'" mood="happy" size="md" />

            <div>
              <div class="font-display font-black text-sm uppercase text-ink-black">{{ player.nickname }}</div>
              <div class="font-condensed text-xs uppercase text-ink-black/70 flex items-center gap-2">
                <span v-if="gameStore.roundSummary?.round_scores?.[player.id]">
                  +{{ gameStore.roundSummary.round_scores[player.id] }} pts
                </span>
                <span v-else>0 pt</span>
                <span
                  v-if="gameStore.roundSummary?.solve_times?.[player.id]"
                  class="text-[11px] font-bold text-game-green bg-game-green/15 px-2 py-0.5 rounded-md border border-ink-black/30"
                >
                  ⏱ {{ gameStore.roundSummary.solve_times[player.id] }}s
                </span>
              </div>
            </div>
          </div>

          <!-- Score Total sur la piste -->
          <div class="text-right">
            <div class="text-game-blue font-condensed font-black text-xl">
              {{ player.score }} pts
            </div>
            <div class="text-[10px] font-display font-bold uppercase text-ink-black/60">
              Total session
            </div>
          </div>
        </div>
      </div>

      <!-- Barre de compte à rebours animée pop-moderniste (si manche intermédiaire) -->
      <div v-if="!gameStore.isGameOver" class="space-y-1.5">
        <div class="flex items-center justify-between text-xs font-display font-black uppercase text-ink-black">
          <span class="flex items-center gap-1.5">
            <Clock class="w-4 h-4 text-game-blue" />
            <span>Prochaine manche</span>
          </span>
          <span class="font-condensed text-sm font-black text-game-blue">{{ countdownRemaining }}s</span>
        </div>
        <div class="w-full h-4 bg-board-cream rounded-full border-[3px] border-ink-black overflow-hidden p-0.5 shadow-pop-xs">
          <div
            class="h-full bg-game-yellow rounded-full transition-all duration-300 ease-linear border-r-2 border-ink-black"
            :style="{ width: `${countdownPercent}%` }"
          />
        </div>
      </div>

      <!-- Actions : Copier Émojis, Fermer & Commandes Master -->
      <div class="flex flex-col sm:flex-row gap-3 pt-2">
        <AppButton
          variant="neutral"
          size="md"
          class="flex-1"
          @click="copyEmojiGrid"
        >
          <Share2 class="w-5 h-5 mr-2" />
          <span>{{ copied ? 'Copié !' : 'Partager mes émojis' }}</span>
        </AppButton>

        <!-- Bouton Master : Lancer la manche suivante immédiatement -->
        <AppButton
          v-if="!gameStore.isGameOver && isMaster"
          variant="success"
          size="md"
          class="flex-1"
          @click="onNextRound"
        >
          <Play class="w-5 h-5 mr-2 fill-current" />
          <span>Lancer la manche suivante</span>
        </AppButton>

        <!-- Bouton Master : Retour au Lobby (si fin de partie) -->
        <AppButton
          v-if="gameStore.isGameOver && isMaster"
          variant="success"
          size="md"
          class="flex-1"
          @click="onReturnLobby"
        >
          <RotateCcw class="w-5 h-5 mr-2" />
          <span>Retourner au Lobby</span>
        </AppButton>
      </div>

      <!-- Message d'attente pour les non-masters -->
      <div v-if="gameStore.isGameOver && !isMaster" class="text-center font-display font-bold text-xs uppercase text-game-blue animate-pulse">
        En attente du Master pour retourner au Lobby...
      </div>
      <div v-else-if="!gameStore.isGameOver && !isMaster" class="text-center font-display font-bold text-xs uppercase text-ink-black/70">
        En attente du Master ou du compte à rebours...
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, onMounted, onUnmounted } from 'vue'
import { Trophy, Share2, RotateCcw, X, Clock, Play } from 'lucide-vue-next'
import { useGameStore } from '~/stores/game'
import { useRoomStore } from '~/stores/room'
import GameMascot from '~/components/ui/GameMascot.vue'
import AppButton from '~/components/ui/AppButton.vue'

const props = defineProps<{
  isMaster: boolean
  onRematch?: () => void
  onReturnLobby?: () => void
  onNextRound?: () => void
}>()

const gameStore = useGameStore()
const roomStore = useRoomStore()
const copied = ref(false)

const countdownRemaining = ref(10)
const totalCountdown = ref(10)
let timerId: any = null

const startCountdown = (seconds: number) => {
  if (timerId) clearInterval(timerId)
  totalCountdown.value = seconds > 0 ? seconds : 10
  countdownRemaining.value = totalCountdown.value
  timerId = setInterval(() => {
    if (countdownRemaining.value > 0) {
      countdownRemaining.value -= 1
    } else {
      clearInterval(timerId)
    }
  }, 1000)
}

watch(() => gameStore.showRoundSummary, (shown) => {
  if (shown && !gameStore.isGameOver) {
    const sec = gameStore.roundSummary?.countdown_sec || 10
    startCountdown(sec)
  } else {
    if (timerId) clearInterval(timerId)
  }
}, { immediate: true })

onUnmounted(() => {
  if (timerId) clearInterval(timerId)
})

const countdownPercent = computed(() => {
  if (totalCountdown.value <= 0) return 0
  return Math.min(100, Math.max(0, (countdownRemaining.value / totalCountdown.value) * 100))
})

const targetLetters = computed(() => {
  const word = gameStore.targetWord || ''
  return word.split('')
})

const closeModal = () => {
  gameStore.showRoundSummary = false
}

const scoreboardPlayers = computed(() => {
  return [...roomStore.players].sort((a, b) => b.score - a.score)
})

const copyEmojiGrid = async () => {
  const grid = gameStore.emojiGrid || '🟩🟨⬛'
  const text = `MiniGames Wordle Multijoueur\nManche ${gameStore.currentRound}/${gameStore.maxRounds}\n\n${grid}\nFestival du Jeu de Société !`
  
  try {
    await navigator.clipboard.writeText(text)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 2500)
  } catch (err) {
    console.error('Erreur lors de la copie', err)
  }
}
</script>
