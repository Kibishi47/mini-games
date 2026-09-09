<template>
  <div class="min-h-screen bg-[#0B0F19] text-slate-100 flex flex-col justify-between selection:bg-indigo-500 selection:text-white">
    <!-- Navbar Desktop -->
    <header class="border-b border-white/10 bg-brand-surface/30 backdrop-blur-md sticky top-0 z-40">
      <div class="max-w-7xl mx-auto px-6 h-20 flex items-center justify-between">
        <div class="flex items-center space-x-3 cursor-pointer" @click="navigateTo('/')">
          <div class="w-11 h-11 rounded-2xl bg-gradient-to-tr from-indigo-600 to-violet-500 flex items-center justify-center shadow-lg shadow-indigo-500/30">
            <Gamepad2 class="w-6 h-6 text-white" />
          </div>
          <div>
            <span class="font-black text-2xl tracking-tight bg-gradient-to-r from-white via-indigo-100 to-indigo-400 bg-clip-text text-transparent">
              MiniGames
            </span>
            <span class="block text-[10px] uppercase font-bold tracking-widest text-indigo-400">Temps Réel Desktop</span>
          </div>
        </div>

        <!-- Profil / Statut -->
        <div class="flex items-center space-x-4">
          <div v-if="authStore.user" class="flex items-center space-x-3 glass-card px-4 py-2 rounded-2xl border border-white/10">
            <img
              :src="authStore.avatar"
              alt="Avatar"
              class="w-9 h-9 rounded-xl bg-slate-800 border border-white/10 p-0.5"
            />
            <div class="text-left">
              <div class="font-bold text-sm text-white flex items-center space-x-1.5">
                <span>{{ authStore.displayName }}</span>
                <span v-if="authStore.isGuest" class="text-[10px] px-2 py-0.5 rounded-md bg-amber-500/20 text-amber-300 font-semibold">
                  Invité
                </span>
              </div>
              <div class="text-[11px] text-slate-400">
                Wordle: {{ authStore.stats?.highest_score || 0 }} pts max
              </div>
            </div>
            
            <button
              @click="authStore.logout"
              title="Déconnexion"
              class="p-2 text-slate-400 hover:text-rose-400 hover:bg-rose-500/10 rounded-xl transition-colors ml-2"
            >
              <LogOut class="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>
    </header>

    <!-- Contenu Principal -->
    <main class="flex-1 max-w-7xl w-full mx-auto px-6 py-12 flex flex-col justify-center">
      <div class="grid grid-cols-1 lg:grid-cols-12 gap-12 items-center">
        
        <!-- Colonne Gauche : Hero Title & Présentation -->
        <div class="lg:col-span-7 space-y-8">
          <div class="inline-flex items-center space-x-2 px-4 py-1.5 rounded-full bg-indigo-500/10 border border-indigo-500/20 text-indigo-400 text-xs font-semibold uppercase tracking-wider">
            <Sparkles class="w-4 h-4" />
            <span>Moteur Temps Réel Server-Authoritative Go 1.23+</span>
          </div>

          <h1 class="text-5xl sm:text-6xl font-black tracking-tight leading-[1.15]">
            Défiez vos amis sur un <span class="bg-gradient-to-r from-indigo-400 via-purple-400 to-pink-400 bg-clip-text text-transparent">Wordle Multijoueur</span> ultra-dynamique.
          </h1>

          <p class="text-slate-400 text-lg leading-relaxed max-w-xl">
            Rejoignez une partie en 1 clic en mode invité, synchronisez vos manches avec chronomètres absolus, admirez les grilles adverses sans spoiler et partagez vos résultats viraux.
          </p>

          <!-- Badges Avantages -->
          <div class="grid grid-cols-3 gap-4 pt-4 max-w-lg">
            <div class="glass-card p-4 rounded-2xl border border-white/5 space-y-1">
              <Zap class="w-5 h-5 text-indigo-400 mb-2" />
              <div class="font-bold text-white text-sm">Grace Period 45s</div>
              <div class="text-xs text-slate-400">Reconnexion sans perte de points</div>
            </div>
            <div class="glass-card p-4 rounded-2xl border border-white/5 space-y-1">
              <EyeOff class="w-5 h-5 text-purple-400 mb-2" />
              <div class="font-bold text-white text-sm">State Masking</div>
              <div class="text-xs text-slate-400">Adversaires en direct sans triche</div>
            </div>
            <div class="glass-card p-4 rounded-2xl border border-white/5 space-y-1">
              <Trophy class="w-5 h-5 text-pink-400 mb-2" />
              <div class="font-bold text-white text-sm">Scoreboard Room</div>
              <div class="text-xs text-slate-400">Historique complet de session</div>
            </div>
          </div>
        </div>

        <!-- Colonne Droite : Cartes d'Actions (Jeu Rapide, Rejoindre, Créer) -->
        <div class="lg:col-span-5 space-y-6">
          <div class="glass-panel p-8 rounded-3xl border border-white/10 shadow-2xl relative overflow-hidden">
            <div class="absolute -top-24 -right-24 w-48 h-48 bg-indigo-500/20 rounded-full blur-3xl pointer-events-none" />

            <!-- Tabs : Rejoindre une Salle / Créer une Salle -->
            <div class="flex p-1.5 rounded-2xl bg-black/40 border border-white/5 mb-6">
              <button
                @click="activeTab = 'join'"
                :class="[
                  'flex-1 py-2.5 rounded-xl font-bold text-sm transition-all flex items-center justify-center space-x-2',
                  activeTab === 'join' ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-600/30' : 'text-slate-400 hover:text-white'
                ]"
              >
                <LogIn class="w-4 h-4" />
                <span>Rejoindre</span>
              </button>

              <button
                @click="activeTab = 'create'"
                :class="[
                  'flex-1 py-2.5 rounded-xl font-bold text-sm transition-all flex items-center justify-center space-x-2',
                  activeTab === 'create' ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-600/30' : 'text-slate-400 hover:text-white'
                ]"
              >
                <PlusCircle class="w-4 h-4" />
                <span>Créer une Salle</span>
              </button>
            </div>

            <!-- TAB : REJOINDRE -->
            <div v-if="activeTab === 'join'" class="space-y-4">
              <div>
                <label class="block text-xs font-bold uppercase tracking-wider text-slate-400 mb-2">
                  Code de la Salle (ex: ABCD-12)
                </label>
                <input
                  v-model="joinCode"
                  type="text"
                  placeholder="ABCD-12"
                  maxlength="10"
                  class="w-full bg-brand-surface/60 border border-white/10 rounded-2xl px-5 py-3.5 text-center font-black tracking-widest text-xl text-white placeholder-slate-600 focus:outline-none focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500 uppercase transition-all"
                />
              </div>

              <!-- Pseudo si non connecté -->
              <div v-if="!authStore.user">
                <label class="block text-xs font-bold uppercase tracking-wider text-slate-400 mb-2">
                  Votre Pseudo d'invité
                </label>
                <input
                  v-model="guestName"
                  type="text"
                  placeholder="Laisser vide pour aléatoire"
                  maxlength="20"
                  class="w-full bg-brand-surface/60 border border-white/10 rounded-2xl px-5 py-3 text-sm text-white placeholder-slate-600 focus:outline-none focus:border-indigo-500"
                />
              </div>

              <button
                @click="handleJoinRoom"
                :disabled="!joinCode.trim() || isLoading"
                class="w-full py-4 rounded-2xl bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 text-white font-extrabold text-base shadow-xl shadow-indigo-600/30 transition-all flex items-center justify-center space-x-2"
              >
                <Play class="w-5 h-5 fill-current" />
                <span>{{ isLoading ? 'Connexion...' : 'Rejoindre la partie' }}</span>
              </button>
            </div>

            <!-- TAB : CRÉER -->
            <div v-if="activeTab === 'create'" class="space-y-5">
              <div>
                <label class="block text-xs font-bold uppercase tracking-wider text-slate-400 mb-2">
                  Longueur des Mots
                </label>
                <div class="grid grid-cols-3 gap-2">
                  <button
                    v-for="len in [5, 6, 7]"
                    :key="len"
                    @click="newRoomSettings.word_length = len"
                    :class="[
                      'py-2.5 rounded-xl font-bold text-sm border transition-all',
                      newRoomSettings.word_length === len
                        ? 'bg-indigo-600/30 border-indigo-500 text-indigo-200'
                        : 'bg-brand-surface/40 border-white/5 text-slate-400 hover:bg-brand-surface'
                    ]"
                  >
                    {{ len }} Lettres
                  </button>
                </div>
              </div>

              <div class="grid grid-cols-2 gap-3">
                <div>
                  <label class="block text-xs font-bold uppercase tracking-wider text-slate-400 mb-2">
                    Chrono par manche
                  </label>
                  <select
                    v-model="newRoomSettings.round_duration"
                    class="w-full bg-brand-surface/60 border border-white/10 rounded-xl px-3 py-2.5 text-sm text-white focus:outline-none focus:border-indigo-500"
                  >
                    <option :value="45">45 secondes</option>
                    <option :value="60">60 secondes</option>
                    <option :value="90">90 secondes</option>
                    <option :value="120">120 secondes</option>
                  </select>
                </div>

                <div>
                  <label class="block text-xs font-bold uppercase tracking-wider text-slate-400 mb-2">
                    Nombre de manches
                  </label>
                  <select
                    v-model="newRoomSettings.max_rounds"
                    class="w-full bg-brand-surface/60 border border-white/10 rounded-xl px-3 py-2.5 text-sm text-white focus:outline-none focus:border-indigo-500"
                  >
                    <option :value="1">1 Manche</option>
                    <option :value="3">3 Manches</option>
                    <option :value="5">5 Manches</option>
                  </select>
                </div>
              </div>

              <!-- Pseudo si non connecté -->
              <div v-if="!authStore.user">
                <label class="block text-xs font-bold uppercase tracking-wider text-slate-400 mb-2">
                  Votre Pseudo
                </label>
                <input
                  v-model="guestName"
                  type="text"
                  placeholder="Laisser vide pour aléatoire"
                  maxlength="20"
                  class="w-full bg-brand-surface/60 border border-white/10 rounded-2xl px-5 py-3 text-sm text-white placeholder-slate-600 focus:outline-none focus:border-indigo-500"
                />
              </div>

              <button
                @click="handleCreateRoom"
                :disabled="isLoading"
                class="w-full py-4 rounded-2xl bg-gradient-to-r from-indigo-600 to-violet-600 hover:from-indigo-500 hover:to-violet-500 text-white font-extrabold text-base shadow-xl shadow-indigo-600/30 transition-all flex items-center justify-center space-x-2"
              >
                <Sparkles class="w-5 h-5" />
                <span>{{ isLoading ? 'Création en cours...' : 'Créer le Salon Master' }}</span>
              </button>
            </div>

            <!-- Message d'erreur -->
            <div v-if="error" class="mt-4 p-3 rounded-xl bg-rose-500/10 border border-rose-500/20 text-rose-400 text-xs text-center font-medium">
              {{ error }}
            </div>
          </div>
        </div>

      </div>
    </main>

    <!-- Footer -->
    <footer class="border-t border-white/5 py-6 text-center text-xs text-slate-500">
      MiniGames &copy; 2026 • Architecture Go + Nuxt 3 distribuée • Prêt pour VPS & Coolify
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Gamepad2, Sparkles, Zap, EyeOff, Trophy, LogIn, PlusCircle, Play, LogOut } from 'lucide-vue-next'
import { useAuthStore } from '~/stores/auth'

const router = useRouter()
const config = useRuntimeConfig()
const authStore = useAuthStore()

const activeTab = ref<'join' | 'create'>('join')
const joinCode = ref('')
const guestName = ref('')
const isLoading = ref(false)
const error = ref<string | null>(null)

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
