<template>
  <!-- Écran de chargement / Reconnexion Pop Moderniste (élimine tout flash de profil par défaut au F5) -->
  <div v-if="isInitializing" class="min-h-screen bg-board-cream text-ink-black flex flex-col items-center justify-center p-6 select-none">
    <div class="bg-board-white p-8 rounded-3xl border-[4px] border-ink-black shadow-pop-lg flex flex-col items-center text-center max-w-sm w-full space-y-4">
      <GameMascot :name="profileStore.mascot || 'dice'" mood="running" size="lg" class="animate-bounce" />
      <div>
        <h2 class="font-display font-black text-2xl uppercase tracking-tight text-ink-black">
          Reconnexion à la salle…
        </h2>
        <p class="font-condensed text-xs uppercase text-game-blue font-bold tracking-wider mt-1">
          Salle {{ roomCode }}
        </p>
      </div>
      <div class="w-full h-3 bg-board-cream rounded-full border-2 border-ink-black overflow-hidden p-0.5">
        <div class="h-full bg-game-yellow rounded-full animate-pulse w-full" />
      </div>
    </div>
  </div>

  <div v-else class="min-h-screen bg-board-cream text-ink-black flex flex-col justify-between selection:bg-game-yellow selection:text-ink-black">
    <!-- Navbar Salle Pop Moderniste (3 blocs équilibrés) -->
    <header class="border-b-[4px] border-ink-black bg-board-white sticky top-0 z-40">
      <div class="max-w-7xl mx-auto px-6 h-20 py-3 flex items-center justify-between">
        
        <!-- Bloc Gauche : Logo / Titre MINIGAMES + Badge Pop FESTIVAL -->
        <div class="flex items-center space-x-3 cursor-pointer" @click="navigateTo('/')">
          <GameMascot name="dice" mood="running" size="sm" />
          <span class="font-display font-black text-2xl tracking-tight text-ink-black uppercase">
            MiniGames
          </span>
          <AppBadge variant="system" class="hidden sm:inline-flex text-[10px] font-black">
            Festival
          </AppBadge>
        </div>

        <!-- Bloc Centre : Cartouche Code de Salle cliquable (avec feedback visuel 'Copié !') + Statut Manche si en jeu -->
        <div class="flex items-center space-x-3">
          <button
            @click="copyRoomCode"
            type="button"
            class="flex items-center space-x-2 border-2 border-ink-black bg-board-cream hover:bg-board-white px-3.5 py-1.5 rounded-xl shadow-pop-xs active:translate-x-[1px] active:translate-y-[1px] active:shadow-none transition-none cursor-pointer"
            :title="isCopied ? 'Code copié dans le presse-papier !' : 'Cliquer pour copier le code de la salle'"
          >
            <span class="font-condensed font-black text-sm tracking-widest uppercase text-game-blue">
              {{ roomCode }}
            </span>
            <span v-if="isCopied" class="flex items-center text-xs font-condensed font-black uppercase text-game-green gap-1">
              <Check class="w-3.5 h-3.5 stroke-[3]" />
              <span>Copié !</span>
            </span>
            <Copy v-else class="w-3.5 h-3.5 text-ink-black/60" />
          </button>

          <!-- Si en jeu : Statut Manche / Chrono + Bouton Master Arrêter la partie -->
          <template v-if="currentView === 'game'">
            <div class="hidden md:flex items-center space-x-2">
              <div class="border-2 border-ink-black px-2.5 py-1 rounded-xl bg-game-yellow font-condensed font-black text-xs uppercase shadow-pop-xs">
                Manche {{ gameStore.currentRound }}/{{ gameStore.maxRounds }}
              </div>
              <div
                :class="[
                  'flex items-center space-x-1 border-2 border-ink-black px-2.5 py-1 rounded-xl font-condensed font-black text-xs uppercase shadow-pop-xs',
                  remainingSeconds <= 10 ? 'bg-game-red text-board-white animate-pulse' : 'bg-board-white text-ink-black'
                ]"
              >
                <Clock class="w-3.5 h-3.5" />
                <span>{{ remainingSeconds }}s</span>
              </div>
              <AppButton
                v-if="roomStore.isMaster"
                variant="danger"
                size="sm"
                class="hidden lg:inline-flex text-xs px-2.5 py-1"
                @click="confirmStopGame"
                title="Arrêter la partie et retourner au lobby"
              >
                <Square class="w-3.5 h-3.5 mr-1 fill-current" />
                <span>Arrêter</span>
              </AppButton>
            </div>
          </template>
        </div>

        <!-- Bloc Droit : Profil Joueur Connecté (Mascotte + Pseudo + Badge MASTER) & Bouton Quitter -->
        <div class="flex items-center space-x-3">
          <div class="flex items-center space-x-2 border-2 border-ink-black bg-board-white px-3 py-1.5 rounded-xl shadow-pop-xs">
            <ClientOnly>
              <GameMascot v-if="profileStore.isReady && profileStore.mascot" :name="profileStore.mascot" mood="idle" size="sm" />
              <div v-else class="w-8 h-8 opacity-0" />
              <template #fallback>
                <div class="w-8 h-8 opacity-0" />
              </template>
            </ClientOnly>
            <div class="text-left">
              <div class="font-display font-black text-xs uppercase text-ink-black leading-tight flex items-center gap-1">
                <span>{{ profileStore.nickname }}</span>
              </div>
              <span v-if="roomStore.isMaster" class="text-[9px] font-condensed uppercase text-game-blue font-black flex items-center gap-0.5">
                <Crown class="w-2.5 h-2.5 text-game-yellow fill-game-yellow" />
                <span>Master</span>
              </span>
              <span v-else-if="roomStore.me?.is_spectator" class="text-[9px] font-condensed uppercase text-ink-black/60 font-black block">
                Spectateur
              </span>
            </div>
          </div>

          <button
            @click="leaveRoom"
            title="Quitter la salle"
            class="p-2 border-2 border-ink-black bg-board-cream hover:bg-game-red hover:text-board-white rounded-xl shadow-pop-xs active:translate-x-[2px] active:translate-y-[2px] active:shadow-none transition-none cursor-pointer"
          >
            <LogOut class="w-4 h-4" />
          </button>
        </div>
      </div>
    </header>

    <!-- Message d'erreur éphémère -->
    <div v-if="errorMessage" class="fixed top-20 left-1/2 -translate-x-1/2 z-50 bg-game-red text-board-white font-display font-black text-sm px-6 py-2.5 rounded-2xl border-[3px] border-ink-black shadow-pop-md flex items-center space-x-2">
      <AlertTriangle class="w-5 h-5" />
      <span>{{ errorMessage }}</span>
    </div>

    <!-- Corps Principal Dynamique (Lobby vs In Game) -->
    <main class="max-w-7xl mx-auto px-6 py-6 flex-1 w-full">
      <!-- 1. VUE LOBBY (En attente du lancement par le Master) -->
      <div v-if="currentView === 'lobby'" class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
        <div class="lg:col-span-8 space-y-6">
          <!-- Bannière Lobby Pop -->
          <AppCard variant="white" shadow="lg" class="space-y-4">
            <div class="flex items-center justify-between border-b-2 border-ink-black pb-4">
              <div class="flex items-center space-x-3">
                <GameMascot :name="masterPlayer?.mascot as any || 'meeple'" mood="happy" size="lg" />
                <div>
                  <h1 class="font-display font-black text-2xl sm:text-3xl uppercase tracking-tight text-ink-black">
                    Salon de Jeu
                  </h1>
                  <p class="font-body text-xs text-ink-black/70">
                    Le Master peut configurer les options et lancer la partie dès qu'il le souhaite !
                  </p>
                </div>
              </div>

              <div class="text-right">
                <AppBadge variant="master">
                  {{ roomStore.players.length }} Joueur(s)
                </AppBadge>
              </div>
            </div>

            <!-- Paramètres de la Salle (Modifiables par le Master) -->
            <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 pt-2">
              <!-- Longueur du Mot -->
              <div class="p-3 bg-board-cream rounded-xl border-2 border-ink-black">
                <label class="block font-display font-bold text-[10px] uppercase text-ink-black/60 mb-1">
                  Longueur du Mot
                </label>
                <select
                  :disabled="!roomStore.isMaster"
                  :value="roomStore.currentRoom?.settings?.word_length || 5"
                  @change="onWordLengthChange($event)"
                  class="w-full bg-board-white border-2 border-ink-black rounded-lg px-2 py-1 font-condensed font-black text-sm uppercase disabled:opacity-60"
                >
                  <option :value="3">3 Lettres</option>
                  <option :value="4">4 Lettres</option>
                  <option :value="5">5 Lettres (Standard)</option>
                  <option :value="6">6 Lettres</option>
                  <option :value="7">7 Lettres</option>
                  <option :value="8">8 Lettres (Expert)</option>
                </select>
              </div>

              <!-- Durée de Manche -->
              <div class="p-3 bg-board-cream rounded-xl border-2 border-ink-black">
                <label class="block font-display font-bold text-[10px] uppercase text-ink-black/60 mb-1">
                  Chrono par Manche
                </label>
                <select
                  :disabled="!roomStore.isMaster"
                  :value="roomStore.currentRoom?.settings?.round_duration || 60"
                  @change="onDurationChange($event)"
                  class="w-full bg-board-white border-2 border-ink-black rounded-lg px-2 py-1 font-condensed font-black text-sm uppercase disabled:opacity-60"
                >
                  <option :value="30">30 secondes (Blitz)</option>
                  <option :value="60">60 secondes</option>
                  <option :value="90">90 secondes</option>
                  <option :value="120">2 minutes</option>
                </select>
              </div>

              <!-- Nombre de Manches -->
              <div class="p-3 bg-board-cream rounded-xl border-2 border-ink-black">
                <label class="block font-display font-bold text-[10px] uppercase text-ink-black/60 mb-1">
                  Nombre de Manches
                </label>
                <select
                  :disabled="!roomStore.isMaster"
                  :value="roomStore.currentRoom?.settings?.max_rounds || 3"
                  @change="onMaxRoundsChange($event)"
                  class="w-full bg-board-white border-2 border-ink-black rounded-lg px-2 py-1 font-condensed font-black text-sm uppercase disabled:opacity-60"
                >
                  <option :value="1">1 Manche</option>
                  <option :value="3">3 Manches (Standard)</option>
                  <option :value="5">5 Manches</option>
                  <option :value="7">7 Manches</option>
                </select>
              </div>

              <!-- Essais Max -->
              <div class="p-3 bg-board-cream rounded-xl border-2 border-ink-black">
                <label class="block font-display font-bold text-[10px] uppercase text-ink-black/60 mb-1">
                  Essais Autorisés
                </label>
                <select
                  :disabled="!roomStore.isMaster"
                  :value="roomStore.currentRoom?.settings?.max_attempts || 6"
                  @change="onMaxAttemptsChange($event)"
                  class="w-full bg-board-white border-2 border-ink-black rounded-lg px-2 py-1 font-condensed font-black text-sm uppercase disabled:opacity-60"
                >
                  <option :value="5">5 Essais</option>
                  <option :value="6">6 Essais (Standard)</option>
                  <option :value="7">7 Essais</option>
                </select>
              </div>
            </div>

            <!-- Bouton Lancer la Partie (Master) -->
            <div class="pt-4 border-t-2 border-ink-black">
              <AppButton
                v-if="roomStore.isMaster"
                variant="primary"
                size="lg"
                class="w-full text-lg"
                @click="handleStartGame"
              >
                <Play class="w-6 h-6 mr-3 fill-current" />
                <span>Lancer la Partie de Wordle !</span>
              </AppButton>
              <div v-else class="text-center p-3 bg-board-cream border-2 border-ink-black rounded-xl font-display font-bold text-sm uppercase text-ink-black/70 flex items-center justify-center gap-2">
                <Clock class="w-4 h-4 text-ink-black/60" />
                <span>En attente que le Master (<strong>{{ masterPlayer?.nickname }}</strong>) lance la partie...</span>
              </div>
            </div>
          </AppCard>

          <!-- Liste des Joueurs dans le Lobby -->
          <AppCard variant="white" shadow="md">
            <h2 class="font-display font-black text-lg uppercase tracking-wider mb-4 border-b-2 border-ink-black pb-2 flex flex-wrap items-center justify-between gap-2">
              <span>Joueurs Présents ({{ roomStore.players.length }})</span>
              <div class="flex items-center space-x-2">
                <AppButton
                  v-if="roomStore.isMaster"
                  variant="neutral"
                  size="sm"
                  class="text-xs"
                  @click="confirmResetScores"
                  title="Remettre à zéro les scores de tous les joueurs"
                >
                  <RotateCcw class="w-3.5 h-3.5 mr-1" />
                  <span>Réinitialiser les scores</span>
                </AppButton>
              </div>
            </h2>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div
                v-for="p in roomStore.players"
                :key="p.id"
                class="p-3 rounded-xl border-2 border-ink-black bg-board-cream flex items-center justify-between shadow-pop-xs"
              >
                <div class="flex items-center space-x-3">
                  <GameMascot :name="(p.mascot as any) || 'dice'" mood="idle" size="sm" />
                  <div>
                    <div class="font-display font-black text-sm uppercase text-ink-black flex items-center space-x-1.5">
                      <span>{{ p.nickname }}</span>
                      <Crown v-if="p.is_master" class="w-4 h-4 text-game-yellow fill-game-yellow inline-block" title="Master" />
                      <AppBadge v-if="p.location === 'in_game'" variant="neutral" class="ml-1 text-[10px] px-2 py-0">
                        En jeu
                      </AppBadge>
                    </div>
                    <span class="text-[10px] font-condensed uppercase text-ink-black/60">
                      Score session : {{ p.score }} pts
                    </span>
                  </div>
                </div>

                <!-- Contrôles de modération pour le Master -->
                <div v-if="roomStore.isMaster && !p.is_master" class="flex items-center space-x-1">
                  <button
                    @click="mutePlayer(p.id, !p.is_muted)"
                    :title="p.is_muted ? 'Débloquer parole' : 'Rendre muet'"
                    class="p-1 border border-ink-black rounded-lg bg-board-white hover:bg-game-yellow text-xs"
                  >
                    <VolumeX v-if="p.is_muted" class="w-3.5 h-3.5 text-game-red" />
                    <Volume2 v-else class="w-3.5 h-3.5" />
                  </button>
                  <button
                    @click="confirmKickPlayer(p.id)"
                    title="Expulser"
                    class="p-1 border border-ink-black rounded-lg bg-board-white hover:bg-game-red hover:text-board-white text-xs"
                  >
                    <UserMinus class="w-3.5 h-3.5" />
                  </button>
                  <button
                    @click="confirmBanPlayer(p.id)"
                    title="Bannir de la salle"
                    class="p-1 border border-ink-black rounded-lg bg-board-white hover:bg-game-red hover:text-board-white text-xs"
                  >
                    <Ban class="w-3.5 h-3.5 text-game-red" />
                  </button>
                </div>
              </div>
            </div>
          </AppCard>
        </div>

        <!-- Chat du Lobby (Hauteur uniforme identique au jeu : 460px) -->
        <div class="lg:col-span-4">
          <ChatPanel :on-send="sendChatMessage" height="h-[460px]" />
        </div>
      </div>

      <!-- 2. VUE EN JEU (Architecture 2 Colonnes : Plateau à gauche, Chat + Adversaires à droite) -->
      <div v-else-if="currentView === 'game'" class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
        <!-- Colonne Gauche (8 cols) : Plateau Wordle et Clavier -->
        <div class="lg:col-span-8 space-y-6">
          <AppCard variant="white" shadow="lg" class="p-6">
            <!-- Mode Spectateur Banner -->
            <div v-if="roomStore.me?.is_spectator" class="mb-4 p-3 bg-game-pink border-2 border-ink-black rounded-xl text-center font-display font-black text-xs uppercase shadow-pop-xs flex items-center justify-center gap-2">
              <AlertTriangle class="w-4 h-4 text-ink-black" />
              <span>Vous observez la manche en cours. Vous participerez activement à la manche suivante !</span>
            </div>

            <!-- Message de réussite personnelle -->
            <div v-if="gameStore.isSolved" class="mb-4 p-3 bg-game-green text-board-white border-2 border-ink-black rounded-xl text-center font-display font-black text-sm uppercase shadow-pop-xs flex items-center justify-center gap-2">
              <Trophy class="w-5 h-5 text-board-white" />
              <span>Bravo ! Vous avez trouvé le mot en {{ gameStore.myAttempts.length }} coup(s) (+{{ gameStore.myRoundScore }} pts) !</span>
            </div>

            <!-- Message d'échec personnel -->
            <div v-else-if="gameStore.isFinished" class="mb-4 p-3 bg-game-red text-board-white border-2 border-ink-black rounded-xl text-center font-display font-black text-sm uppercase shadow-pop-xs flex items-center justify-center gap-2">
              <AlertTriangle class="w-5 h-5 text-board-white" />
              <span>Échec pour cette manche ! En attente des autres joueurs...</span>
            </div>

            <!-- Plateau et Clavier Wordle -->
            <WordleGrid :on-submit="submitGuess" />

            <!-- Bandeau persistant quand la partie est terminée et que la modale est masquée -->
            <div
              v-if="gameStore.isGameOver && !gameStore.showRoundSummary"
              class="mt-6 p-4 bg-game-yellow/20 border-2 border-ink-black rounded-2xl flex items-center justify-between shadow-pop-xs"
            >
              <div class="flex items-center space-x-2">
                <Trophy class="w-5 h-5 text-ink-black" />
                <span class="font-display font-black text-sm uppercase text-ink-black">Partie terminée !</span>
              </div>
              <AppButton
                variant="primary"
                size="sm"
                @click="handleLocalReturnToLobby"
              >
                <RotateCcw class="w-4 h-4 mr-1.5" />
                <span>Retourner au Lobby</span>
              </AppButton>
            </div>
          </AppCard>
        </div>

        <!-- Colonne Droite (4 cols) : 1. Chat (hauteur identique 460px) + 2. Adversaires en direct en dessous -->
        <div class="lg:col-span-4 space-y-6">
          <!-- 1. Chat Gazette du Festival (Hauteur uniforme fixe 460px identique au lobby) -->
          <ChatPanel :on-send="sendChatMessage" height="h-[460px]" />

          <!-- 2. Adversaires en Direct repositionnés sous le chat -->
          <OpponentPreview
            :opponents="gameStore.opponents"
            :max-attempts="gameStore.maxAttempts"
          />
        </div>
      </div>
    </main>

    <!-- Modal de Fin de Manche / Fin de Partie avec Scoreboard -->
    <ScoreboardModal
      :is-master="roomStore.isMaster"
      :on-rematch="rematch"
      :on-return-lobby="handleLocalReturnToLobby"
      :on-individual-return-lobby="handleLocalReturnToLobby"
      :on-next-round="nextRound"
      @return-to-lobby="handleLocalReturnToLobby"
    />

    <!-- Modale de Confirmation Pop Moderniste -->
    <AppConfirmModal />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, onMounted, onUnmounted } from 'vue'
import { Copy, Check, Crown, Clock, LogOut, Play, Volume2, VolumeX, UserMinus, Ban, AlertTriangle, Square, RotateCcw } from 'lucide-vue-next'
import { useRoute, useRouter } from 'vue-router'
import { useProfileStore } from '~/stores/profile'
import { useRoomStore } from '~/stores/room'
import { useGameStore } from '~/stores/game'
import { useWebSocket } from '~/composables/useWebSocket'
import { useConfirm } from '~/composables/useConfirm'

import AppCard from '~/components/ui/AppCard.vue'
import AppButton from '~/components/ui/AppButton.vue'
import AppBadge from '~/components/ui/AppBadge.vue'
import GameMascot from '~/components/ui/GameMascot.vue'
import ChatPanel from '~/components/chat/ChatPanel.vue'
import WordleGrid from '~/components/wordle/WordleGrid.vue'
import OpponentPreview from '~/components/wordle/OpponentPreview.vue'
import ScoreboardModal from '~/components/scoreboard/ScoreboardModal.vue'
import AppConfirmModal from '~/components/ui/AppConfirmModal.vue'

const route = useRoute()
const router = useRouter()
const roomCode = computed(() => String(route.params.code || '').toUpperCase())

const profileStore = useProfileStore()
const roomStore = useRoomStore()
const gameStore = useGameStore()
const { confirm } = useConfirm()

// Initialisation immédiate du profil avant la connexion WS
profileStore.initProfile()

// Initialisation WebSocket
const {
  isConnected,
  isInitializing,
  errorMessage,
  sendChatMessage,
  updateSettings,
  startGame,
  stopGame,
  returnToLobby,
  returnLobby: wsReturnLobby,
  resetScores,
  leaveRoom: wsLeaveRoom,
  nextRound,
  submitGuess,
  kickPlayer: wsKickPlayer,
  banPlayer: wsBanPlayer,
  mutePlayer,
  rematch,
} = useWebSocket(roomCode.value)

// Vue active : 'lobby' ou 'game' (pilotage individuel)
const currentView = ref<'lobby' | 'game'>('lobby')

// Retour au lobby propre et strictement individuel : mise à jour locale immédiate et émission player:return_lobby
const handleLocalReturnToLobby = () => {
  currentView.value = 'lobby'
  gameStore.returnToLobbyView()
  wsReturnLobby()
}

// Synchronisation de la vue avec les transitions globales de partie
watch(
  () => [roomStore.currentRoom?.status, roomStore.currentRoom?.round_state, roomStore.me?.location],
  ([status, roundState, myLocation]) => {
    if (status === 'in_game' && roundState === 'playing') {
      currentView.value = 'game'
    } else if (status === 'in_lobby') {
      currentView.value = 'lobby'
      gameStore.returnToLobbyView()
    } else if (status === 'in_game' && roundState === 'game_over') {
      // Lors d'une reconnexion / F5 en game_over : vérifier la localisation individuelle
      if (myLocation === 'lobby') {
        currentView.value = 'lobby'
        gameStore.returnToLobbyView()
      } else {
        currentView.value = 'game'
      }
    }
  },
  { immediate: true }
)

const confirmStopGame = async () => {
  const ok = await confirm({
    title: 'ARRÊTER LA PARTIE ?',
    message: 'Voulez-vous vraiment arrêter la partie en cours et ramener tous les joueurs au lobby ?',
    confirmLabel: 'ARRÊTER',
    cancelLabel: 'CONTINUER',
    variant: 'danger',
  })
  if (ok) {
    stopGame()
  }
}

const confirmResetScores = async () => {
  const ok = await confirm({
    title: 'RÉINITIALISER LES SCORES ?',
    message: 'Tous les scores cumulés de la session seront remis à zéro.',
    confirmLabel: 'RÉINITIALISER',
    cancelLabel: 'ANNULER',
    variant: 'danger',
  })
  if (ok) {
    resetScores()
  }
}

const confirmKickPlayer = async (playerId: string) => {
  const player = roomStore.players.find(p => p.id === playerId)
  const name = player?.nickname || 'ce joueur'
  const ok = await confirm({
    title: 'EXPULSER LE JOUEUR ?',
    message: `Voulez-vous expulser ${name} de la salle ? Il pourra revenir s'il possède le code.`,
    confirmLabel: 'EXPULSER',
    cancelLabel: 'ANNULER',
    variant: 'danger',
  })
  if (ok) {
    wsKickPlayer(playerId)
  }
}

const confirmBanPlayer = async (playerId: string) => {
  const player = roomStore.players.find(p => p.id === playerId)
  const name = player?.nickname || 'ce joueur'
  const ok = await confirm({
    title: 'BANNIR LE JOUEUR ?',
    message: `Voulez-vous bannir définitivement ${name} de cette salle ? Tout accès lui sera refusé.`,
    confirmLabel: 'BANNIR',
    cancelLabel: 'ANNULER',
    variant: 'danger',
  })
  if (ok) {
    wsBanPlayer(playerId)
  }
}

const handleStartGame = async () => {
  const playersStillInGame = roomStore.players.filter(p => p.location === 'in_game')
  if (playersStillInGame.length > 0) {
    const names = playersStillInGame.map(p => p.nickname).join(', ')
    const ok = await confirm({
      title: 'LANCER LA PARTIE ?',
      message: `Certains joueurs sont encore sur l'écran des scores (${names}). Voulez-vous lancer quand même la partie ?`,
      confirmLabel: 'LANCER',
      cancelLabel: 'ATTENDRE',
      variant: 'primary',
    })
    if (!ok) {
      return
    }
  }
  startGame()
}

const masterPlayer = computed(() => roomStore.masterPlayer)

// Compte à rebours absolu
const remainingSeconds = ref(60)
let timerInterval: any = null

const updateCountdown = () => {
  if (gameStore.endsAt) {
    const diff = Math.max(0, Math.floor((gameStore.endsAt.getTime() - Date.now()) / 1000))
    remainingSeconds.value = diff
  }
}

onMounted(() => {
  profileStore.initProfile()
  timerInterval = setInterval(updateCountdown, 1000)
})

onUnmounted(() => {
  if (timerInterval) clearInterval(timerInterval)
})

const isCopied = ref(false)
let copyTimeout: any = null

const copyRoomCode = async () => {
  try {
    await navigator.clipboard.writeText(roomCode.value)
    isCopied.value = true
    if (copyTimeout) clearTimeout(copyTimeout)
    copyTimeout = setTimeout(() => {
      isCopied.value = false
    }, 1500)
  } catch (e) {
    // ignore
  }
}

const leaveRoom = async () => {
  if (roomStore.currentRoom?.status === 'in_game') {
    const ok = await confirm({
      title: 'QUITTER LA PARTIE ?',
      message: 'Voulez-vous vraiment quitter la partie en cours et abandonner la salle ?',
      confirmLabel: 'QUITTER',
      cancelLabel: 'RESTER',
      variant: 'danger',
    })
    if (!ok) return
  }

  wsLeaveRoom()
  profileStore.clearSession()
  router.push('/')
}

// Mise à jour des paramètres par le Master
const onWordLengthChange = (e: Event) => {
  const val = parseInt((e.target as HTMLSelectElement).value, 10)
  updateSettings({
    ...(roomStore.currentRoom?.settings || {}),
    word_length: val,
  })
}

const onDurationChange = (e: Event) => {
  const val = parseInt((e.target as HTMLSelectElement).value, 10)
  updateSettings({
    ...(roomStore.currentRoom?.settings || {}),
    round_duration: val,
  })
}

const onMaxRoundsChange = (e: Event) => {
  const val = parseInt((e.target as HTMLSelectElement).value, 10)
  updateSettings({
    ...(roomStore.currentRoom?.settings || {}),
    max_rounds: val,
  })
}

const onMaxAttemptsChange = (e: Event) => {
  const val = parseInt((e.target as HTMLSelectElement).value, 10)
  updateSettings({
    ...(roomStore.currentRoom?.settings || {}),
    max_attempts: val,
  })
}
</script>
