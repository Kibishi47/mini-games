<template>
  <div class="min-h-screen bg-board-cream text-ink-black flex flex-col justify-between selection:bg-game-yellow selection:text-ink-black">
    <!-- Navbar Lobby Festival -->
    <header class="border-b-[4px] border-ink-black bg-board-white sticky top-0 z-40">
      <div class="max-w-7xl mx-auto px-6 h-20 flex items-center justify-between">
        <div class="flex items-center space-x-4">
          <AppButton
            variant="neutral"
            size="sm"
            @click="leaveRoom"
          >
            <ArrowLeft class="w-4 h-4 mr-1.5" />
            <span>Quitter</span>
          </AppButton>

          <!-- Cartouche Code de la Room avec découpe pointillée -->
          <div class="flex items-center space-x-3 border-[3px] border-dashed border-ink-black bg-game-yellow/20 px-4 py-1.5 rounded-2xl">
            <span class="font-display font-black text-xs uppercase text-ink-black/60">Salle</span>
            <span class="font-condensed font-black text-2xl tracking-widest text-ink-black">{{ roomCode }}</span>
            <button
              @click="copyRoomCode"
              title="Copier le code"
              class="p-1 border-2 border-ink-black bg-game-yellow hover:bg-[#ffe033] rounded-lg shadow-pop-xs active:translate-x-0.5 active:translate-y-0.5 active:shadow-none transition-none"
            >
              <Copy class="w-4 h-4 text-ink-black" />
            </button>
          </div>
        </div>

        <!-- Master Start Game Action (Pas de ready check !) -->
        <div class="flex items-center space-x-4">
          <div v-if="roomStore.isMaster">
            <AppButton
              variant="success"
              size="lg"
              @click="startGame"
            >
              <Play class="w-5 h-5 fill-current mr-2" />
              <span>Lancer la Partie (Master)</span>
            </AppButton>
          </div>
          <div v-else class="flex items-center space-x-2 border-[3px] border-ink-black bg-board-white px-4 py-2 rounded-2xl shadow-pop-xs">
            <GameMascot name="meeple" mood="idle" size="sm" />
            <span class="font-display font-bold text-xs uppercase text-ink-black">En attente du Master...</span>
          </div>
        </div>
      </div>
    </header>

    <!-- Corps du Lobby -->
    <main class="flex-1 max-w-7xl w-full mx-auto px-6 py-8 grid grid-cols-1 lg:grid-cols-12 gap-8">
      
      <!-- Colonne Gauche : Vignettes Joueurs & Réglages -->
      <div class="lg:col-span-5 flex flex-col space-y-6">
        
        <!-- Liste des Joueurs façon Cartes Vignettes -->
        <AppCard variant="white" shadow="md" class="flex-1 flex flex-col">
          <div class="flex items-center justify-between pb-3 border-b-[3px] border-ink-black">
            <div class="flex items-center space-x-2">
              <Users class="w-5 h-5 text-game-blue" />
              <h3 class="font-display font-black text-lg uppercase text-ink-black">Pions en lice</h3>
            </div>
            <AppBadge variant="master">{{ roomStore.players.length }} Joueur(s)</AppBadge>
          </div>

          <div class="mt-4 space-y-3 flex-1 overflow-y-auto max-h-[380px] pr-1">
            <div
              v-for="(p, idx) in roomStore.players"
              :key="p.user_id"
              class="flex items-center justify-between p-3 rounded-2xl border-[3px] border-ink-black bg-board-cream shadow-pop-xs"
            >
              <div class="flex items-center space-x-3 min-w-0">
                <!-- Mascotte associée à l'index du joueur -->
                <GameMascot :name="getMascotName(idx)" mood="idle" size="md" />
                
                <div class="truncate">
                  <div class="font-display font-black text-sm text-ink-black uppercase flex items-center space-x-1.5 truncate">
                    <span class="truncate">{{ p.display_username }}</span>
                    <Crown v-if="p.role === 'master'" class="w-4 h-4 text-game-yellow fill-current flex-shrink-0" />
                  </div>
                  <div class="text-[11px] font-condensed text-ink-black/60 flex items-center space-x-2 uppercase mt-0.5">
                    <span>Score : {{ p.score }} pts</span>
                    <span v-if="p.is_muted" class="text-game-red font-black">• Muet</span>
                  </div>
                </div>
              </div>

              <!-- Modération Master (Interdiction absolue de se cibler soi-même) -->
              <div v-if="roomStore.isMaster && p.user_id !== authStore.user?.id" class="flex items-center space-x-1.5">
                <button
                  @click="performAction(p.is_muted ? 'unmute' : 'mute', p.user_id)"
                  :title="p.is_muted ? 'Dé-muter' : 'Muter'"
                  class="p-1.5 border-2 border-ink-black rounded-lg bg-game-yellow hover:bg-[#ffe033] shadow-pop-xs active:translate-x-0.5 active:translate-y-0.5 active:shadow-none transition-none"
                >
                  <VolumeX v-if="!p.is_muted" class="w-4 h-4 text-ink-black" />
                  <Volume2 v-else class="w-4 h-4 text-game-green" />
                </button>

                <button
                  @click="performAction('kick', p.user_id)"
                  title="Expulser"
                  class="p-1.5 border-2 border-ink-black rounded-lg bg-game-red hover:bg-[#ef4444] text-board-white shadow-pop-xs active:translate-x-0.5 active:translate-y-0.5 active:shadow-none transition-none"
                >
                  <UserX class="w-4 h-4" />
                </button>

                <button
                  @click="performAction('ban', p.user_id)"
                  title="Bannir (30 min)"
                  class="p-1.5 border-2 border-ink-black rounded-lg bg-ink-black text-game-yellow shadow-pop-xs active:translate-x-0.5 active:translate-y-0.5 active:shadow-none transition-none"
                >
                  <ShieldAlert class="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
        </AppCard>

        <!-- Panneau de Configuration du Tournoi -->
        <AppCard variant="yellow" shadow="md">
          <div class="flex items-center justify-between pb-2 border-b-2 border-ink-black mb-3">
            <div class="flex items-center space-x-2">
              <Sliders class="w-4 h-4 text-ink-black" />
              <h4 class="font-display font-black text-sm uppercase text-ink-black">Règles du Tournoi Wordle</h4>
            </div>
            <span v-if="!roomStore.isMaster" class="text-[10px] font-bold uppercase text-ink-black/60">Lecture seule</span>
          </div>

          <div class="grid grid-cols-2 gap-3 text-xs">
            <div class="bg-board-white p-2.5 rounded-xl border-2 border-ink-black">
              <span class="block font-bold text-ink-black/60 uppercase text-[10px]">Longueur</span>
              <span class="font-condensed font-black text-base text-ink-black">{{ roomStore.settings.word_length }} Lettres</span>
            </div>
            <div class="bg-board-white p-2.5 rounded-xl border-2 border-ink-black">
              <span class="block font-bold text-ink-black/60 uppercase text-[10px]">Chronomètre</span>
              <span class="font-condensed font-black text-base text-ink-black">{{ roomStore.settings.round_duration }}s</span>
            </div>
            <div class="bg-board-white p-2.5 rounded-xl border-2 border-ink-black">
              <span class="block font-bold text-ink-black/60 uppercase text-[10px]">Manches</span>
              <span class="font-condensed font-black text-base text-ink-black">{{ roomStore.settings.max_rounds }} Tours</span>
            </div>
            <div class="bg-board-white p-2.5 rounded-xl border-2 border-ink-black">
              <span class="block font-bold text-ink-black/60 uppercase text-[10px]">Dictionnaire</span>
              <span class="font-condensed font-black text-base text-ink-black uppercase">{{ roomStore.settings.language }}</span>
            </div>
          </div>
        </AppCard>

      </div>

      <!-- Colonne Droite : Chat Stylisé Pop Moderniste -->
      <div class="lg:col-span-7 h-[640px]">
        <ChatPanel :on-send="sendChat" />
      </div>

    </main>

    <!-- Toast d'erreur Pop -->
    <div v-if="errorMessage" class="fixed bottom-6 right-6 z-50 bg-game-red text-board-white border-[3px] border-ink-black px-5 py-3 rounded-2xl shadow-pop-md font-display font-black uppercase text-sm flex items-center space-x-2">
      <AlertTriangle class="w-5 h-5" />
      <span>{{ errorMessage }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import {
  ArrowLeft, Copy, Play, Crown, Users, VolumeX, Volume2, UserX, ShieldAlert,
  Sliders, AlertTriangle
} from 'lucide-vue-next'
import { useRoute } from 'vue-router'
import { useRoomStore } from '~/stores/room'
import { useAuthStore } from '~/stores/auth'
import { useWebSocket } from '~/composables/useWebSocket'
import GameMascot, { type MascotName } from '~/components/ui/GameMascot.vue'
import AppButton from '~/components/ui/AppButton.vue'
import AppCard from '~/components/ui/AppCard.vue'
import AppBadge from '~/components/ui/AppBadge.vue'
import ChatPanel from '~/components/chat/ChatPanel.vue'

const route = useRoute()
const roomCode = String(route.params.code)

const roomStore = useRoomStore()
const authStore = useAuthStore()

const {
  errorMessage,
  sendChat,
  performAction,
  startGame,
  leaveRoom,
} = useWebSocket(roomCode)

onMounted(() => {
  authStore.initAuth()
})

const mascotPool: MascotName[] = ['dice', 'domino', 'card', 'knight', 'd20', 'meeple']
const getMascotName = (index: number): MascotName => {
  return mascotPool[index % mascotPool.length]
}

const copyRoomCode = async () => {
  try {
    await navigator.clipboard.writeText(roomCode)
    alert(`Code ${roomCode} copié !`)
  } catch (e) {
    // ignore
  }
}
</script>
