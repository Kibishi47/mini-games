<template>
  <div class="min-h-screen bg-board-cream text-ink-black flex flex-col justify-between selection:bg-game-yellow selection:text-ink-black">
    <!-- Header Game Festival -->
    <header class="border-b-[4px] border-ink-black bg-board-white sticky top-0 z-40">
      <div class="max-w-7xl mx-auto px-6 h-20 flex items-center justify-between">
        
        <!-- Info Salle & Manche -->
        <div class="flex items-center space-x-6">
          <div>
            <div class="flex items-center space-x-2">
              <span class="font-display font-black text-xs uppercase text-game-blue">Salle</span>
              <span class="font-condensed font-black text-lg text-ink-black">{{ roomCode }}</span>
            </div>
            <div class="font-display font-black text-2xl uppercase tracking-tight text-ink-black">
              Manche {{ gameStore.currentRound }} / {{ gameStore.maxRounds }}
            </div>
          </div>

          <!-- Mode Spectateur Badge -->
          <AppBadge v-if="gameStore.isSpectator" variant="spectator">
            <Eye class="w-3.5 h-3.5 mr-1 text-ink-black" />
            <span>Spectateur Actif</span>
          </AppBadge>
        </div>

        <!-- Chronomètre Absolu Capsule Pop -->
        <div class="flex items-center space-x-4">
          <div
            :class="[
              'px-5 py-2 rounded-2xl font-condensed font-black text-2xl flex items-center space-x-2 border-[3px] border-ink-black transition-none select-none',
              remainingSeconds <= 10
                ? 'bg-game-red text-board-white animate-pop-pulse shadow-pop-sm'
                : 'bg-game-yellow text-ink-black shadow-pop-sm'
            ]"
          >
            <Clock class="w-6 h-6" />
            <span>{{ remainingSeconds }}s</span>
            <!-- Mascotte paniquée si chrono < 10s -->
            <GameMascot v-if="remainingSeconds <= 10" name="dice" mood="panic" size="sm" class="ml-1" />
          </div>

          <!-- Bouton Quitter -->
          <AppButton
            variant="neutral"
            size="sm"
            @click="leaveRoom"
          >
            Quitter
          </AppButton>
        </div>
      </div>
    </header>

    <!-- Plateau de Jeu Principal -->
    <main class="flex-1 max-w-7xl w-full mx-auto px-6 py-6 grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
      
      <!-- Colonne Centrale : Grille Wordle Personnelle avec Réaction Mascotte en direct -->
      <div class="lg:col-span-7 flex flex-col items-center justify-center relative">
        <!-- Mascotte animée sur le côté qui réagit en direct -->
        <div class="hidden xl:block absolute -left-16 top-6">
          <GameMascot
            name="dice"
            :mood="gameStore.isSolved ? 'happy' : (gameStore.isFinished && !gameStore.isSolved) ? 'dead' : 'idle'"
            size="lg"
          />
        </div>

        <div v-if="!gameStore.isSpectator" class="w-full">
          <WordleGrid :on-submit="submitGuess" />
        </div>
        <div v-else class="w-full max-w-md my-12">
          <AppCard variant="white" shadow="lg" class="text-center space-y-4 p-8">
            <GameMascot name="meeple" mood="idle" size="lg" class="mx-auto" />
            <h3 class="font-display font-black text-2xl uppercase text-ink-black">Mode Spectateur</h3>
            <p class="font-body text-sm font-semibold text-ink-black/70">
              Partie rejointe en cours. Vous intégrerez le tournoi dès la manche suivante !
            </p>
          </AppCard>
        </div>
      </div>

      <!-- Colonne Droite : Adversaires en direct & Chat Pop -->
      <div class="lg:col-span-5 flex flex-col space-y-5">
        <!-- Aperçu Masqué des Adversaires (State-Masking) -->
        <OpponentPreview
          :opponents="gameStore.opponents"
          :max-attempts="gameStore.maxAttempts"
        />

        <!-- Chat In-Game -->
        <div class="h-64">
          <ChatPanel :on-send="sendChat" />
        </div>
      </div>

    </main>

    <!-- Modal Récapitulatif de Manche & Scoreboard Final -->
    <ScoreboardModal
      :is-master="roomStore.isMaster"
      :on-rematch="requestRematch"
    />

    <!-- Toasts / Erreurs -->
    <div v-if="errorMessage" class="fixed bottom-6 right-6 z-50 bg-game-red text-board-white border-[3px] border-ink-black px-5 py-3 rounded-2xl shadow-pop-md font-display font-black uppercase text-sm flex items-center space-x-2">
      <AlertTriangle class="w-5 h-5" />
      <span>{{ errorMessage }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Eye, Clock, AlertTriangle } from 'lucide-vue-next'
import { useRoute } from 'vue-router'
import { useRoomStore } from '~/stores/room'
import { useGameStore } from '~/stores/game'
import { useAuthStore } from '~/stores/auth'
import { useWebSocket } from '~/composables/useWebSocket'
import GameMascot from '~/components/ui/GameMascot.vue'
import AppButton from '~/components/ui/AppButton.vue'
import AppBadge from '~/components/ui/AppBadge.vue'
import AppCard from '~/components/ui/AppCard.vue'
import WordleGrid from '~/components/wordle/WordleGrid.vue'
import OpponentPreview from '~/components/wordle/OpponentPreview.vue'
import ChatPanel from '~/components/chat/ChatPanel.vue'
import ScoreboardModal from '~/components/scoreboard/ScoreboardModal.vue'

const route = useRoute()
const roomCode = String(route.params.code)

const roomStore = useRoomStore()
const gameStore = useGameStore()
const authStore = useAuthStore()

const {
  errorMessage,
  sendChat,
  submitGuess,
  requestRematch,
  leaveRoom,
} = useWebSocket(roomCode)

// Décompte local synchronisé via ends_at absolu
const now = ref(Date.now())
let timerInterval: NodeJS.Timeout | null = null

const remainingSeconds = computed(() => {
  if (!gameStore.endsAt) return 0
  const diff = Math.ceil((gameStore.endsAt.getTime() - now.value) / 1000)
  return Math.max(0, diff)
})

onMounted(() => {
  authStore.initAuth()
  timerInterval = setInterval(() => {
    now.value = Date.now()
  }, 500)
})

onUnmounted(() => {
  if (timerInterval) clearInterval(timerInterval)
})
</script>
