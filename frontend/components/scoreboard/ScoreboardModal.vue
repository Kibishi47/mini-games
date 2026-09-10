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

      <!-- Titre et Mot Secret Révélé -->
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

      <!-- Grand Podium Final à 3 marches (si fin de partie) -->
      <div v-if="gameStore.isGameOver && scoreboardPlayers.length >= 2" class="pt-4 pb-2">
        <div class="flex items-end justify-center gap-3 sm:gap-4 max-w-md mx-auto">
          <!-- 2ème Place (Argent) -->
          <div v-if="scoreboardPlayers[1]" class="flex-1 flex flex-col items-center">
            <GameMascot :name="(scoreboardPlayers[1].mascot as any) || 'domino'" mood="happy" size="md" class="mb-1" />
            <div class="font-display font-black text-xs uppercase text-ink-black truncate max-w-[80px] sm:max-w-[100px]">
              {{ scoreboardPlayers[1].nickname }}
            </div>
            <div class="font-condensed font-black text-xs text-game-blue mb-1">
              {{ scoreboardPlayers[1].score }} pts
            </div>
            <div class="w-full h-24 bg-board-white rounded-t-2xl border-[3px] border-ink-black flex flex-col items-center justify-center shadow-pop-xs">
              <span class="font-display font-black text-2xl text-ink-black">#2</span>
              <span class="text-[9px] font-condensed uppercase font-bold text-ink-black/60">Argent</span>
            </div>
          </div>

          <!-- 1ère Place (Or) -->
          <div v-if="scoreboardPlayers[0]" class="flex-1 flex flex-col items-center">
            <div class="text-xs font-black uppercase text-game-yellow mb-0.5 animate-bounce">Champion</div>
            <GameMascot :name="(scoreboardPlayers[0].mascot as any) || 'dice'" mood="happy" size="lg" class="mb-1" />
            <div class="font-display font-black text-sm uppercase text-ink-black truncate max-w-[90px] sm:max-w-[110px]">
              {{ scoreboardPlayers[0].nickname }}
            </div>
            <div class="font-condensed font-black text-sm text-game-blue mb-1">
              {{ scoreboardPlayers[0].score }} pts
            </div>
            <div class="w-full h-32 bg-game-yellow rounded-t-2xl border-[3.5px] border-ink-black flex flex-col items-center justify-center shadow-pop-sm">
              <span class="font-display font-black text-3xl text-ink-black">#1</span>
              <span class="text-[10px] font-condensed uppercase font-black text-ink-black/80">Vainqueur</span>
            </div>
          </div>

          <!-- 3ème Place (Bronze) -->
          <div v-if="scoreboardPlayers[2]" class="flex-1 flex flex-col items-center">
            <GameMascot :name="(scoreboardPlayers[2].mascot as any) || 'meeple'" mood="happy" size="md" class="mb-1" />
            <div class="font-display font-black text-xs uppercase text-ink-black truncate max-w-[80px] sm:max-w-[100px]">
              {{ scoreboardPlayers[2].nickname }}
            </div>
            <div class="font-condensed font-black text-xs text-game-blue mb-1">
              {{ scoreboardPlayers[2].score }} pts
            </div>
            <div class="w-full h-18 bg-game-pink/40 rounded-t-2xl border-[3px] border-ink-black flex flex-col items-center justify-center shadow-pop-xs">
              <span class="font-display font-black text-xl text-ink-black">#3</span>
              <span class="text-[9px] font-condensed uppercase font-bold text-ink-black/60">Bronze</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Piste de Score de Jeu de Société -->
      <div class="space-y-3 max-h-56 overflow-y-auto pr-1">
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

      <!-- Barre de compte à rebours fluide 60 FPS CSS pur (8 secondes) -->
      <div v-if="!gameStore.isGameOver" class="space-y-1.5">
        <div class="flex items-center justify-between text-xs font-display font-black uppercase text-ink-black">
          <span class="flex items-center gap-1.5">
            <Clock class="w-4 h-4 text-game-blue" />
            <span>Prochaine manche</span>
          </span>
          <span class="font-condensed text-sm font-black text-game-blue">{{ countdownSeconds }}s</span>
        </div>
        <div class="w-full h-4 bg-board-cream rounded-full border-[3px] border-ink-black overflow-hidden p-0.5 shadow-pop-xs">
          <div
            :key="countdownKey"
            class="h-full bg-game-yellow rounded-full border-r-2 border-ink-black animate-progress-linear"
            :style="{ animationDuration: `${totalDuration}s` }"
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

        <!-- Bouton individuel : Retourner au Lobby (disponible pour tous en game_over) -->
        <AppButton
          v-if="gameStore.isGameOver"
          variant="primary"
          size="md"
          class="flex-1"
          @click="handleIndividualReturnLobby"
        >
          <RotateCcw class="w-5 h-5 mr-2" />
          <span>Retourner au Lobby</span>
        </AppButton>

        <!-- Bouton Master : Retour au Lobby pour tout le monde (si fin de partie) -->
        <AppButton
          v-if="gameStore.isGameOver && isMaster"
          variant="success"
          size="md"
          class="flex-1"
          @click="onReturnLobby"
        >
          <RotateCcw class="w-5 h-5 mr-2" />
          <span>Revanche / Tous au Lobby</span>
        </AppButton>
      </div>

      <!-- Message d'attente pour les non-masters si manche en cours -->
      <div v-if="!gameStore.isGameOver && !isMaster" class="text-center font-display font-bold text-xs uppercase text-ink-black/70">
        En attente du Master ou du compte à rebours…
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, onUnmounted } from 'vue'
import { Trophy, Share2, RotateCcw, X, Clock, Play } from 'lucide-vue-next'
import { useGameStore } from '~/stores/game'
import { useRoomStore } from '~/stores/room'
import GameMascot from '~/components/ui/GameMascot.vue'
import AppButton from '~/components/ui/AppButton.vue'

const props = defineProps<{
  isMaster: boolean
  onRematch?: () => void
  onReturnLobby?: () => void
  onIndividualReturnLobby?: () => void
  onNextRound?: () => void
}>()

const gameStore = useGameStore()
const roomStore = useRoomStore()
const copied = ref(false)

const handleIndividualReturnLobby = () => {
  gameStore.showRoundSummary = false
  if (props.onIndividualReturnLobby) {
    props.onIndividualReturnLobby()
  }
}

const countdownKey = ref(0)
const countdownSeconds = ref(8)
const totalDuration = ref(8)
let secondInterval: any = null

watch(() => gameStore.showRoundSummary, (shown) => {
  if (shown && !gameStore.isGameOver) {
    const duration = gameStore.roundSummary?.countdown_sec || 8
    totalDuration.value = duration
    countdownSeconds.value = duration
    countdownKey.value += 1

    if (secondInterval) clearInterval(secondInterval)
    secondInterval = setInterval(() => {
      if (countdownSeconds.value > 0) {
        countdownSeconds.value -= 1
      } else {
        clearInterval(secondInterval)
      }
    }, 1000)
  } else {
    if (secondInterval) clearInterval(secondInterval)
  }
}, { immediate: true })

onUnmounted(() => {
  if (secondInterval) clearInterval(secondInterval)
})

const targetLetters = computed(() => {
  const word = gameStore.targetWord || ''
  return word.split('')
})

const closeModal = () => {
  gameStore.showRoundSummary = false
  if (gameStore.isGameOver && props.onIndividualReturnLobby) {
    props.onIndividualReturnLobby()
  }
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

<style scoped>
@keyframes progressBarLinear {
  from {
    width: 100%;
  }
  to {
    width: 0%;
  }
}

.animate-progress-linear {
  animation-name: progressBarLinear;
  animation-timing-function: linear;
  animation-fill-mode: forwards;
}
</style>
