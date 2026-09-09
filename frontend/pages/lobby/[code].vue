<template>
  <div class="min-h-screen bg-[#0B0F19] text-slate-100 flex flex-col justify-between">
    <!-- Navbar Lobby -->
    <header class="border-b border-white/10 bg-brand-surface/30 backdrop-blur-md sticky top-0 z-40">
      <div class="max-w-7xl mx-auto px-6 h-20 flex items-center justify-between">
        <div class="flex items-center space-x-4">
          <button
            @click="leaveRoom"
            class="flex items-center space-x-2 text-slate-400 hover:text-white glass-card px-4 py-2 rounded-xl border border-white/10 transition-colors"
          >
            <ArrowLeft class="w-4 h-4" />
            <span class="text-xs font-semibold">Quitter</span>
          </button>

          <div>
            <div class="flex items-center space-x-2">
              <h2 class="font-extrabold text-xl text-white">Salle {{ roomCode }}</h2>
              <button
                @click="copyRoomCode"
                title="Copier le code"
                class="p-1.5 rounded-lg bg-white/5 hover:bg-white/10 text-indigo-400 transition-colors"
              >
                <Copy class="w-4 h-4" />
              </button>
            </div>
            <span class="text-xs text-slate-400">
              {{ isConnected ? 'Connecté en direct' : 'Reconnexion en cours...' }}
            </span>
          </div>
        </div>

        <!-- Master Start Game Action (Pas de ready check !) -->
        <div class="flex items-center space-x-4">
          <div v-if="roomStore.isMaster">
            <button
              @click="startGame"
              class="px-6 py-3 rounded-2xl bg-gradient-to-r from-emerald-500 to-teal-500 hover:from-emerald-400 hover:to-teal-400 text-white font-extrabold text-sm shadow-xl shadow-emerald-500/30 hover:scale-[1.02] active:scale-[0.98] transition-all flex items-center space-x-2"
            >
              <Play class="w-4 h-4 fill-current" />
              <span>Lancer la Partie (Master)</span>
            </button>
          </div>
          <div v-else class="text-xs text-slate-400 flex items-center space-x-2 glass-card px-4 py-2.5 rounded-xl">
            <Crown class="w-4 h-4 text-amber-400" />
            <span>En attente du lancement par le Master...</span>
          </div>
        </div>
      </div>
    </header>

    <!-- Corps du Lobby -->
    <main class="flex-1 max-w-7xl w-full mx-auto px-6 py-8 grid grid-cols-1 lg:grid-cols-12 gap-8">
      
      <!-- Colonne Gauche : Liste des Joueurs & Modération Master -->
      <div class="lg:col-span-4 flex flex-col space-y-6">
        <div class="glass-panel p-5 rounded-2xl border border-white/10 flex-1 flex flex-col">
          <div class="flex items-center justify-between pb-4 border-b border-white/10">
            <div class="flex items-center space-x-2">
              <Users class="w-5 h-5 text-indigo-400" />
              <h3 class="font-bold text-white text-base">Joueurs présents</h3>
            </div>
            <span class="text-xs px-2.5 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 font-bold">
              {{ roomStore.players.length }}
            </span>
          </div>

          <div class="mt-4 space-y-2.5 flex-1 overflow-y-auto pr-1">
            <div
              v-for="p in roomStore.players"
              :key="p.user_id"
              class="flex items-center justify-between p-3 rounded-xl bg-white/5 border border-white/5 transition-all"
            >
              <div class="flex items-center space-x-3 min-w-0">
                <div class="relative">
                  <img
                    :src="p.avatar_url || 'https://api.dicebear.com/7.x/bottts/svg?seed=' + p.display_username"
                    class="w-10 h-10 rounded-xl bg-slate-800 border border-white/10 p-0.5"
                  />
                  <div
                    :class="[
                      'absolute -bottom-1 -right-1 w-3.5 h-3.5 rounded-full border-2 border-[#0B0F19]',
                      p.is_connected ? 'bg-emerald-500' : 'bg-amber-500'
                    ]"
                  />
                </div>
                
                <div class="truncate">
                  <div class="font-bold text-sm text-white flex items-center space-x-1.5 truncate">
                    <span class="truncate">{{ p.display_username }}</span>
                    <Crown v-if="p.role === 'master'" class="w-3.5 h-3.5 text-amber-400 flex-shrink-0" />
                    <Eye v-else-if="p.role === 'spectator'" class="w-3.5 h-3.5 text-indigo-400 flex-shrink-0" />
                  </div>
                  <div class="text-[10px] text-slate-400 flex items-center space-x-2">
                    <span>Score cumulé: {{ p.score }}</span>
                    <span v-if="p.is_muted" class="text-rose-400 font-semibold">• Muet</span>
                  </div>
                </div>
              </div>

              <!-- Contrôles Modération Master (Interdiction sur soi-même) -->
              <div v-if="roomStore.isMaster && p.user_id !== authStore.user?.id" class="flex items-center space-x-1">
                <button
                  @click="performAction(p.is_muted ? 'unmute' : 'mute', p.user_id)"
                  :title="p.is_muted ? 'Redonner la parole' : 'Réduire au silence'"
                  class="p-1.5 rounded-lg hover:bg-white/10 text-slate-400 hover:text-amber-400 transition-colors"
                >
                  <VolumeX v-if="!p.is_muted" class="w-4 h-4" />
                  <Volume2 v-else class="w-4 h-4 text-emerald-400" />
                </button>

                <button
                  @click="performAction('kick', p.user_id)"
                  title="Expulser de la salle"
                  class="p-1.5 rounded-lg hover:bg-rose-500/10 text-slate-400 hover:text-rose-400 transition-colors"
                >
                  <UserX class="w-4 h-4" />
                </button>

                <button
                  @click="performAction('ban', p.user_id)"
                  title="Bannir (30min)"
                  class="p-1.5 rounded-lg hover:bg-rose-500/20 text-slate-400 hover:text-rose-500 transition-colors"
                >
                  <ShieldAlert class="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Paramètres de la Salle (Modifiables en direct par le Master) -->
        <div class="glass-panel p-5 rounded-2xl border border-white/10 space-y-4">
          <div class="flex items-center justify-between pb-3 border-b border-white/10">
            <div class="flex items-center space-x-2">
              <Sliders class="w-4 h-4 text-indigo-400" />
              <h4 class="font-bold text-sm text-white">Règles du Jeu Wordle</h4>
            </div>
            <span v-if="!roomStore.isMaster" class="text-[10px] text-slate-500 italic">Lecture seule</span>
          </div>

          <div class="grid grid-cols-2 gap-3 text-xs">
            <div>
              <span class="text-slate-400 block mb-1">Longueur mot</span>
              <span class="font-bold text-white text-sm">{{ roomStore.settings.word_length }} Lettres</span>
            </div>
            <div>
              <span class="text-slate-400 block mb-1">Chrono / manche</span>
              <span class="font-bold text-white text-sm">{{ roomStore.settings.round_duration }}s</span>
            </div>
            <div>
              <span class="text-slate-400 block mb-1">Nombre de manches</span>
              <span class="font-bold text-white text-sm">{{ roomStore.settings.max_rounds }}</span>
            </div>
            <div>
              <span class="text-slate-400 block mb-1">Dictionnaire</span>
              <span class="font-bold text-white text-sm uppercase">{{ roomStore.settings.language }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Colonne Droite : Chat en Temps Réel -->
      <div class="lg:col-span-8 h-[640px]">
        <ChatPanel :on-send="sendChat" />
      </div>

    </main>

    <!-- Notification d'erreur éphémère -->
    <div v-if="errorMessage" class="fixed bottom-6 right-6 z-50 bg-rose-500/90 text-white px-5 py-3 rounded-2xl shadow-2xl backdrop-blur-md flex items-center space-x-2 border border-rose-400/30 text-sm font-semibold animate-in slide-in-from-bottom-5">
      <AlertTriangle class="w-5 h-5 text-white" />
      <span>{{ errorMessage }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import {
  ArrowLeft, Copy, Play, Crown, Eye, Users, VolumeX, Volume2, UserX, ShieldAlert,
  Sliders, AlertTriangle
} from 'lucide-vue-next'
import { useRoute } from 'vue-router'
import { useRoomStore } from '~/stores/room'
import { useAuthStore } from '~/stores/auth'
import { useWebSocket } from '~/composables/useWebSocket'
import ChatPanel from '~/components/chat/ChatPanel.vue'

const route = useRoute()
const roomCode = String(route.params.code)

const roomStore = useRoomStore()
const authStore = useAuthStore()

const {
  isConnected,
  errorMessage,
  sendChat,
  performAction,
  startGame,
  leaveRoom,
} = useWebSocket(roomCode)

onMounted(() => {
  authStore.initAuth()
})

const copyRoomCode = async () => {
  try {
    await navigator.clipboard.writeText(roomCode)
    alert(`Code ${roomCode} copié dans le presse-papier !`)
  } catch (e) {
    // ignore
  }
}
</script>
