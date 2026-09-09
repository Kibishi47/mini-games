<template>
  <div class="min-h-screen bg-[#0B0F19] text-slate-100 flex flex-col justify-between">
    <!-- Header Game -->
    <header class="border-b border-white/10 bg-brand-surface/30 backdrop-blur-md sticky top-0 z-40">
      <div class="max-w-7xl mx-auto px-6 h-20 flex items-center justify-between">
        
        <!-- Info Salle & Manche -->
        <div class="flex items-center space-x-6">
          <div>
            <span class="text-xs font-bold uppercase tracking-wider text-indigo-400">Salle {{ roomCode }}</span>
            <div class="font-black text-xl text-white">
              Manche {{ gameStore.currentRound }} / {{ gameStore.maxRounds }}
            </div>
          </div>

          <!-- Mode Spectateur Badge -->
          <div v-if="gameStore.isSpectator" class="px-3 py-1 rounded-full bg-indigo-500/20 border border-indigo-500/30 text-indigo-300 text-xs font-bold flex items-center space-x-1.5">
            <Eye class="w-4 h-4" />
            <span>Mode Spectateur Actif</span>
          </div>
        </div>

        <!-- Chronomètre Absolu Synchronisé ends_at -->
        <div class="flex items-center space-x-4">
          <div :class="[
            'px-5 py-2.5 rounded-2xl font-black text-xl flex items-center space-x-2 border transition-all',
            remainingSeconds <= 10
              ? 'bg-rose-500/20 border-rose-500/40 text-rose-400 animate-pulse'
              : 'glass-card border-white/10 text-white'
          ]">
            <Clock class="w-5 h-5 text-indigo-400" />
            <span>{{ remainingSeconds }}s</span>
          </div>

          <!-- Bouton Quitter -->
          <button
            @click="leaveRoom"
            class="glass-card px-4 py-2.5 rounded-xl border border-white/10 hover:bg-rose-500/20 hover:text-rose-400 text-slate-400 text-xs font-bold transition-all"
          >
            Quitter
          </button>
        </div>
      </div>
    </header>

    <!-- Plateau de Jeu Principal -->
    <main class="flex-1 max-w-7xl w-full mx-auto px-6 py-6 grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
      
      <!-- Colonne Centrale : Grille Wordle Personnelle -->
      <div class="lg:col-span-6 flex flex-col items-center justify-center">
        <div v-if="!gameStore.isSpectator" class="w-full">
          <WordleGrid :on-submit="submitGuess" />
        </div>
        <div v-else class="glass-panel p-8 rounded-3xl border border-white/10 text-center space-y-4 max-w-md my-12">
          <Eye class="w-12 h-12 text-indigo-400 mx-auto" />
          <h3 class="text-xl font-bold text-white">Vous êtes spectateur</h3>
          <p class="text-sm text-slate-400">
            Vous avez rejoint la salle pendant une manche active. Vous participerez activement dès la manche suivante !
          </p>
        </div>
      </div>

      <!-- Colonne Droite : Adversaires en direct & Chat -->
      <div class="lg:col-span-6 flex flex-col space-y-6">
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
    <div v-if="errorMessage" class="fixed bottom-6 right-6 z-50 bg-rose-500/90 text-white px-5 py-3 rounded-2xl shadow-2xl backdrop-blur-md flex items-center space-x-2 border border-rose-400/30 text-sm font-semibold animate-in slide-in-from-bottom-5">
      <AlertTriangle class="w-5 h-5 text-white" />
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
