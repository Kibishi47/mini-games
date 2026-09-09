<template>
  <div class="min-h-screen bg-board-cream text-ink-black flex flex-col justify-between selection:bg-game-yellow selection:text-ink-black">
    <!-- Navbar Festival -->
    <header class="border-b-[4px] border-ink-black bg-board-white sticky top-0 z-40">
      <div class="max-w-7xl mx-auto px-6 h-20 flex items-center justify-between">
        
        <!-- Logo percutant & Mascotte -->
        <div class="flex items-center space-x-3 cursor-pointer" @click="navigateTo('/')">
          <GameMascot name="dice" mood="running" size="md" />
          <div>
            <div class="flex items-center space-x-2">
              <span class="font-display font-black text-3xl tracking-tight text-ink-black uppercase">
                MiniGames
              </span>
              <span class="px-2 py-0.5 rounded-full border-2 border-ink-black bg-game-yellow font-condensed text-[10px] uppercase tracking-wider">
                Festival
              </span>
            </div>
            <span class="block text-[11px] font-display font-bold uppercase tracking-widest text-game-blue">
              Jeux de Plateau & Wordle Desktop
            </span>
          </div>
        </div>

        <!-- Profil / Statut / Bouton de connexion -->
        <div class="flex items-center space-x-4">
          <!-- Connecté localement -->
          <div v-if="authStore.user && !authStore.isGuest" class="flex items-center space-x-3 border-[3px] border-ink-black bg-board-white px-4 py-2 rounded-2xl shadow-pop-xs">
            <GameMascot name="knight" mood="idle" size="sm" />
            <div class="text-left">
              <div class="font-display font-black text-sm text-ink-black uppercase">
                {{ authStore.displayName }}
              </div>
              <div class="text-[11px] font-condensed text-ink-black/60 uppercase">
                Score Max : {{ authStore.stats?.highest_score || 0 }} pts
              </div>
            </div>
            
            <button
              @click="authStore.logout"
              title="Déconnexion"
              class="p-1.5 border-2 border-ink-black bg-board-cream hover:bg-game-red hover:text-board-white rounded-xl shadow-pop-xs active:translate-x-0.5 active:translate-y-0.5 active:shadow-none transition-none ml-2"
            >
              <LogOut class="w-4 h-4" />
            </button>
          </div>

          <!-- Si invité connecté -->
          <div v-else-if="authStore.user && authStore.isGuest" class="flex items-center space-x-3">
            <div class="flex items-center space-x-2 border-[3px] border-ink-black bg-board-white px-3 py-1.5 rounded-xl shadow-pop-xs">
              <GameMascot name="domino" mood="idle" size="sm" />
              <div class="text-left">
                <span class="font-display font-black text-xs uppercase text-ink-black">{{ authStore.displayName }}</span>
                <span class="ml-1.5 border border-ink-black px-1.5 py-0.5 rounded-md bg-game-yellow font-condensed text-[9px] uppercase">
                  Invité
                </span>
              </div>
            </div>

            <AppButton
              variant="secondary"
              size="sm"
              @click="showAuthModal = true"
            >
              <UserCheck class="w-3.5 h-3.5 mr-1.5" />
              <span>Sauvegarder mon compte</span>
            </AppButton>
          </div>

          <!-- Si non connecté -->
          <AppButton
            v-else
            variant="primary"
            size="md"
            @click="showAuthModal = true"
          >
            <LogIn class="w-4 h-4 mr-2" />
            <span>Connexion / Inscription</span>
          </AppButton>
        </div>
      </div>
    </header>

    <!-- Contenu Principal -->
    <main class="flex-1 max-w-7xl w-full mx-auto px-6 py-12 flex flex-col justify-center">
      <div class="grid grid-cols-1 lg:grid-cols-12 gap-10 items-center">
        
        <!-- Colonne Gauche : Titre d'Affiche de Festival & Mascottes -->
        <div class="lg:col-span-7 space-y-6">
          <div class="inline-flex items-center space-x-2 border-[3px] border-ink-black px-4 py-1.5 rounded-full bg-game-pink font-condensed text-xs uppercase tracking-wider shadow-pop-xs">
            <Sparkles class="w-4 h-4 text-ink-black" />
            <span>Tournois Multijoueur en Direct</span>
          </div>

          <h1 class="font-display font-black text-6xl sm:text-7xl leading-[1.05] tracking-tight uppercase text-ink-black">
            L'Arène Festive du <span class="bg-game-yellow px-2 border-[4px] border-ink-black inline-block transform -rotate-1 shadow-pop-md">Wordle</span> Multijoueur !
          </h1>

          <p class="font-body font-semibold text-ink-black/80 text-lg leading-relaxed max-w-xl">
            Entrez sur le plateau en 1 clic. Affrontez vos amis manche par manche, observez leurs tuiles en temps réel sans triche et partagez vos victoires avec fierté !
          </p>

          <!-- Badges Avantages Façon Cartes de Jeu -->
          <div class="grid grid-cols-3 gap-4 pt-2 max-w-xl">
            <AppCard variant="white" shadow="sm" class="space-y-1">
              <Zap class="w-5 h-5 text-game-blue mb-1" />
              <div class="font-display font-black uppercase text-sm text-ink-black">Grace Period</div>
              <div class="text-xs font-body font-medium text-ink-black/70">45s en cas de déco</div>
            </AppCard>

            <AppCard variant="white" shadow="sm" class="space-y-1">
              <EyeOff class="w-5 h-5 text-game-red mb-1" />
              <div class="font-display font-black uppercase text-sm text-ink-black">State Masking</div>
              <div class="text-xs font-body font-medium text-ink-black/70">Direct sans spoiler</div>
            </AppCard>

            <AppCard variant="white" shadow="sm" class="space-y-1">
              <Trophy class="w-5 h-5 text-game-green mb-1" />
              <div class="font-display font-black uppercase text-sm text-ink-black">Piste Score</div>
              <div class="text-xs font-body font-medium text-ink-black/70">Cumul de session</div>
            </AppCard>
          </div>
        </div>

        <!-- Colonne Droite : Cartouche d'action Rejoindre / Créer -->
        <div class="lg:col-span-5">
          <AppCard variant="white" shadow="lg" class="p-7">
            <!-- Tabs Stylisées -->
            <div class="flex border-[3px] border-ink-black rounded-2xl bg-board-cream p-1 mb-6 shadow-pop-xs">
              <button
                @click="activeTab = 'join'"
                :class="[
                  'flex-1 py-2.5 rounded-xl font-display font-black text-sm uppercase transition-none flex items-center justify-center space-x-2',
                  activeTab === 'join' ? 'bg-game-yellow border-2 border-ink-black shadow-pop-xs text-ink-black' : 'text-ink-black/60 hover:text-ink-black'
                ]"
              >
                <LogIn class="w-4 h-4" />
                <span>Rejoindre</span>
              </button>

              <button
                @click="activeTab = 'create'"
                :class="[
                  'flex-1 py-2.5 rounded-xl font-display font-black text-sm uppercase transition-none flex items-center justify-center space-x-2',
                  activeTab === 'create' ? 'bg-game-blue border-2 border-ink-black shadow-pop-xs text-board-white' : 'text-ink-black/60 hover:text-ink-black'
                ]"
              >
                <PlusCircle class="w-4 h-4" />
                <span>Créer Salon</span>
              </button>
            </div>

            <!-- TAB : REJOINDRE -->
            <div v-if="activeTab === 'join'" class="space-y-4">
              <div>
                <label class="block text-xs font-display font-black uppercase tracking-wider text-ink-black mb-1.5">
                  Code de la Salle (ex: ABCD-12)
                </label>
                <AppInput
                  v-model="joinCode"
                  placeholder="ABCD-12"
                  maxlength="10"
                  custom-class="text-center font-condensed tracking-widest text-2xl uppercase font-black"
                />
              </div>

              <!-- Pseudo si non connecté -->
              <div v-if="!authStore.user">
                <label class="block text-xs font-display font-black uppercase tracking-wider text-ink-black mb-1.5">
                  Votre Pseudo d'invité
                </label>
                <AppInput
                  v-model="guestName"
                  placeholder="Laisser vide pour aléatoire"
                  maxlength="20"
                />
              </div>

              <AppButton
                variant="primary"
                size="lg"
                :disabled="!joinCode.trim() || isLoading"
                class="w-full mt-2"
                @click="handleJoinRoom"
              >
                <Play class="w-5 h-5 fill-current mr-2" />
                <span>{{ isLoading ? 'Connexion...' : 'Rejoindre la Partie' }}</span>
              </AppButton>
            </div>

            <!-- TAB : CRÉER -->
            <div v-if="activeTab === 'create'" class="space-y-4">
              <div>
                <label class="block text-xs font-display font-black uppercase tracking-wider text-ink-black mb-1.5">
                  Longueur des Mots
                </label>
                <div class="grid grid-cols-3 gap-2">
                  <button
                    v-for="len in [5, 6, 7]"
                    :key="len"
                    @click="newRoomSettings.word_length = len"
                    :class="[
                      'py-2 rounded-xl font-display font-black text-sm uppercase border-[3px] border-ink-black transition-none',
                      newRoomSettings.word_length === len
                        ? 'bg-game-yellow shadow-pop-xs text-ink-black'
                        : 'bg-board-cream text-ink-black/60 hover:bg-board-white'
                    ]"
                  >
                    {{ len }} Lettres
                  </button>
                </div>
              </div>

              <div class="grid grid-cols-2 gap-3">
                <div>
                  <label class="block text-xs font-display font-black uppercase tracking-wider text-ink-black mb-1.5">
                    Chrono / Manche
                  </label>
                  <select
                    v-model="newRoomSettings.round_duration"
                    class="w-full bg-board-white border-[3px] border-ink-black rounded-xl px-3 py-2 text-sm font-display font-bold uppercase focus:outline-none focus:shadow-pop-xs"
                  >
                    <option :value="45">45 secondes</option>
                    <option :value="60">60 secondes</option>
                    <option :value="90">90 secondes</option>
                    <option :value="120">120 secondes</option>
                  </select>
                </div>

                <div>
                  <label class="block text-xs font-display font-black uppercase tracking-wider text-ink-black mb-1.5">
                    Manches
                  </label>
                  <select
                    v-model="newRoomSettings.max_rounds"
                    class="w-full bg-board-white border-[3px] border-ink-black rounded-xl px-3 py-2 text-sm font-display font-bold uppercase focus:outline-none focus:shadow-pop-xs"
                  >
                    <option :value="1">1 Manche</option>
                    <option :value="3">3 Manches</option>
                    <option :value="5">5 Manches</option>
                  </select>
                </div>
              </div>

              <!-- Pseudo si non connecté -->
              <div v-if="!authStore.user">
                <label class="block text-xs font-display font-black uppercase tracking-wider text-ink-black mb-1.5">
                  Votre Pseudo
                </label>
                <AppInput
                  v-model="guestName"
                  placeholder="Laisser vide pour aléatoire"
                  maxlength="20"
                />
              </div>

              <AppButton
                variant="secondary"
                size="lg"
                :disabled="isLoading"
                class="w-full mt-2"
                @click="handleCreateRoom"
              >
                <Sparkles class="w-5 h-5 mr-2" />
                <span>{{ isLoading ? 'Création en cours...' : 'Ouvrir le Salon Master' }}</span>
              </AppButton>
            </div>

            <!-- Message d'erreur -->
            <div v-if="error" class="mt-4 p-3 rounded-xl bg-game-red/10 border-2 border-game-red text-game-red text-xs font-black text-center uppercase tracking-wide">
              {{ error }}
            </div>
          </AppCard>
        </div>

      </div>
    </main>

    <!-- Footer Festival -->
    <footer class="border-t-[3px] border-ink-black bg-board-white py-5 text-center text-xs font-display font-bold uppercase tracking-wider text-ink-black/70">
      MiniGames &copy; 2026 • Festival Pop-Moderniste de Jeux de Plateau • Go 1.23+ & Nuxt 3
    </footer>

    <!-- Modale Connexion / Inscription / Discord -->
    <AuthModal
      :is-open="showAuthModal"
      @close="showAuthModal = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Sparkles, Zap, EyeOff, Trophy, LogIn, PlusCircle, Play, LogOut, UserCheck } from 'lucide-vue-next'
import { useAuthStore } from '~/stores/auth'
import GameMascot from '~/components/ui/GameMascot.vue'
import AppCard from '~/components/ui/AppCard.vue'
import AppButton from '~/components/ui/AppButton.vue'
import AppInput from '~/components/ui/AppInput.vue'
import AuthModal from '~/components/auth/AuthModal.vue'

const router = useRouter()
const config = useRuntimeConfig()
const authStore = useAuthStore()

const activeTab = ref<'join' | 'create'>('join')
const joinCode = ref('')
const guestName = ref('')
const isLoading = ref(false)
const error = ref<string | null>(null)
const showAuthModal = ref(false)

const newRoomSettings = reactive({
  game_type: 'wordle',
  word_length: 5,
  round_duration: 60,
  max_rounds: 3,
  max_attempts: 6,
  language: 'fr',
})

onMounted(() => {
  authStore.initAuth()
})

const ensureAuth = async () => {
  if (!authStore.user || !authStore.token) {
    await authStore.loginGuest(guestName.value)
  }
}

const handleJoinRoom = async () => {
  error.value = null
  isLoading.value = true
  try {
    await ensureAuth()
    const cleanCode = joinCode.value.toUpperCase().trim()
    const res = await $fetch<any>(`${config.public.apiUrl}/api/rooms/${cleanCode}/join`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authStore.token}` },
    })

    if (res.room.status === 'in_game') {
      router.push(`/game/${cleanCode}`)
    } else {
      router.push(`/lobby/${cleanCode}`)
    }
  } catch (err: any) {
    error.value = err.data?.error || 'Impossible de rejoindre la salle (code invalide ou salle fermée)'
  } finally {
    isLoading.value = false
  }
}

const handleCreateRoom = async () => {
  error.value = null
  isLoading.value = true
  try {
    await ensureAuth()
    const res = await $fetch<any>(`${config.public.apiUrl}/api/rooms`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authStore.token}` },
      body: { settings: newRoomSettings },
    })

    router.push(`/lobby/${res.code}`)
  } catch (err: any) {
    error.value = err.data?.error || 'Erreur lors de la création de la salle'
  } finally {
    isLoading.value = false
  }
}
</script>
