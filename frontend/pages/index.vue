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
              Wordle Multijoueur Desktop (3 à 8 Lettres)
            </span>
          </div>
        </div>

        <!-- Profil Invité en direct -->
        <div class="flex items-center space-x-3 border-[3px] border-ink-black bg-board-white px-4 py-2 rounded-2xl shadow-pop-xs">
          <GameMascot :name="profileStore.mascot" mood="idle" size="sm" />
          <div class="text-left">
            <div class="font-display font-black text-sm text-ink-black uppercase leading-tight">
              {{ profileStore.nickname }}
            </div>
          </div>
        </div>
      </div>
    </header>
    <!-- Message d'alerte Pop Moderniste si salle fermée ou expirée -->
    <div v-if="closedRoomAlert" class="max-w-2xl mx-auto mt-6 px-6 w-full">
      <div class="bg-game-yellow text-ink-black font-display font-black text-sm px-6 py-3.5 rounded-2xl border-[3.5px] border-ink-black shadow-pop-md flex items-center justify-between">
        <div class="flex items-center space-x-3">
          <AlertTriangle class="w-6 h-6 text-game-red flex-shrink-0" />
          <span>Cette room n'existe plus ou a été fermée pour inactivité.</span>
        </div>
        <button @click="closedRoomAlert = false" class="text-ink-black hover:text-game-red font-black text-base ml-4">
          ✕
        </button>
      </div>
    </div>

    <!-- Corps de la page d'accueil -->
    <main class="max-w-7xl mx-auto px-6 py-10 flex-1 flex flex-col items-center justify-center w-full">
      <div class="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start w-full max-w-6xl">
        
        <!-- Colonne Gauche : Personnalisation Invité -->
        <div class="lg:col-span-5 space-y-6">
          <AppCard variant="white" shadow="lg">
            <div class="flex items-center space-x-3 pb-4 mb-4 border-b-2 border-ink-black">
              <UserCircle class="w-6 h-6 text-game-blue" />
              <h2 class="font-display font-black text-xl uppercase tracking-wide">Mon Profil Festival</h2>
            </div>

            <!-- Pseudo -->
            <div class="space-y-2 mb-6">
              <label class="block font-display font-bold text-xs uppercase tracking-wider text-ink-black/70">
                Ton Pseudo de Joueur
              </label>
              <AppInput
                v-model="inputNickname"
                maxlength="15"
                placeholder="Ex: WordleKing"
                @input="updateProfile"
              />
            </div>

            <!-- Choix de Mascotte -->
            <div class="space-y-3 mb-6">
              <label class="block font-display font-bold text-xs uppercase tracking-wider text-ink-black/70">
                Choisis ta Mascotte Fétiche
              </label>
              <div class="grid grid-cols-3 gap-3">
                <button
                  v-for="m in MASCOTS"
                  :key="m.id"
                  @click="selectMascot(m.id)"
                  :class="[
                    'flex flex-col items-center justify-center p-3 rounded-2xl border-[3px] border-ink-black transition-none select-none relative',
                    profileStore.mascot === m.id
                      ? 'bg-game-yellow shadow-pop-sm scale-105 z-10'
                      : 'bg-board-cream hover:bg-board-white shadow-pop-xs'
                  ]"
                >
                  <GameMascot :name="m.id" :mood="profileStore.mascot === m.id ? 'happy' : 'idle'" size="md" />
                  <span class="font-display font-black text-[11px] uppercase mt-2 text-ink-black text-center truncate w-full">
                    {{ m.id }}
                  </span>
                </button>
              </div>
            </div>

            <!-- Mascotte active en vitrine -->
            <div class="p-4 rounded-2xl border-2 border-ink-black bg-board-cream flex items-center space-x-4">
              <GameMascot :name="profileStore.mascot" mood="running" size="lg" />
              <div>
                <span class="font-display font-black text-base uppercase text-ink-black block">
                  {{ activeMascotInfo?.name }}
                </span>
                <span class="text-xs font-body text-ink-black/70">
                  {{ activeMascotInfo?.desc }}
                </span>
              </div>
            </div>
          </AppCard>
        </div>

        <!-- Colonne Droite : Lancer ou Rejoindre une Partie -->
        <div class="lg:col-span-7 space-y-6">
          <!-- Créer une Salle -->
          <AppCard variant="yellow" shadow="lg" class="space-y-4">
            <div class="flex items-center justify-between">
              <div class="flex items-center space-x-3">
                <Sparkles class="w-7 h-7 text-ink-black" />
                <h2 class="font-display font-black text-2xl uppercase tracking-tight">Créer une Room Privée</h2>
              </div>
              <AppBadge variant="master">Gratuit & Instantané</AppBadge>
            </div>

            <p class="font-body text-sm text-ink-black/80">
              Invite tes amis dans ton salon de jeu de société personnalisé. Tu seras désigné <strong>Master</strong> pour lancer les manches de 3 à 8 lettres à ta convenance.
            </p>

            <AppButton
              variant="secondary"
              size="lg"
              class="w-full"
              :disabled="isLoading"
              @click="handleCreateRoom"
            >
              <Play class="w-6 h-6 mr-3 fill-current" />
              <span>Créer la Salle & Devenir Master</span>
            </AppButton>
          </AppCard>

          <!-- Rejoindre une Salle Existante -->
          <AppCard variant="white" shadow="lg" class="space-y-4">
            <div class="flex items-center space-x-3 pb-3 border-b-2 border-ink-black">
              <LogIn class="w-6 h-6 text-game-blue" />
              <h2 class="font-display font-black text-xl uppercase tracking-wide">Rejoindre une Room</h2>
            </div>

            <p class="font-body text-sm text-ink-black/80">
              Un ami t'a partagé un code court (ex: <code>ABCD-12</code>) ? Tape-le ci-dessous pour entrer dans la salle !
            </p>

            <form @submit.prevent="handleJoinRoom" class="space-y-3">
              <div class="flex gap-3">
                <AppInput
                  v-model="inputRoomCode"
                  placeholder="CODE (ex: ABCD-12)"
                  maxlength="7"
                  class="uppercase text-center font-condensed font-black tracking-widest text-lg"
                />
                <AppButton
                  type="submit"
                  variant="primary"
                  size="md"
                  :disabled="isLoading || !inputRoomCode.trim()"
                >
                  <span>Entrer</span>
                </AppButton>
              </div>
            </form>

            <div v-if="errorMessage" class="p-3 bg-game-red/10 border-2 border-game-red rounded-xl text-game-red font-display font-bold text-xs uppercase text-center">
              ⚠️ {{ errorMessage }}
            </div>
          </AppCard>
        </div>

      </div>
    </main>

    <!-- Footer Pop Festival -->
    <footer class="border-t-[3px] border-ink-black bg-board-white py-4 px-6 text-center text-xs font-display font-bold uppercase tracking-wider text-ink-black/60">
      MiniGames — 100% In-Memory & Redis 7 — Moteur Wordle Server-Authoritative Pop Moderniste
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { UserCircle, Sparkles, Play, LogIn, AlertTriangle } from 'lucide-vue-next'
import { useRoute, useRouter } from 'vue-router'
import { useProfileStore, MASCOTS } from '~/stores/profile'
import type { MascotName } from '~/components/ui/GameMascot.vue'
import AppCard from '~/components/ui/AppCard.vue'
import AppButton from '~/components/ui/AppButton.vue'
import AppBadge from '~/components/ui/AppBadge.vue'
import AppInput from '~/components/ui/AppInput.vue'
import GameMascot from '~/components/ui/GameMascot.vue'

const profileStore = useProfileStore()
const config = useRuntimeConfig()
const route = useRoute()
const router = useRouter()

const inputNickname = ref('')
const inputRoomCode = ref('')
const isLoading = ref(false)
const errorMessage = ref('')
const closedRoomAlert = ref(false)

onMounted(() => {
  profileStore.initProfile()
  inputNickname.value = profileStore.nickname
  if (route.query.error === 'room_closed') {
    closedRoomAlert.value = true
  }
})

const activeMascotInfo = computed(() => {
  return MASCOTS.find(m => m.id === profileStore.mascot) || MASCOTS[0]
})

const selectMascot = (id: MascotName) => {
  profileStore.setProfile(inputNickname.value, id)
}

const updateProfile = () => {
  profileStore.setProfile(inputNickname.value, profileStore.mascot)
}

const handleCreateRoom = async () => {
  isLoading.value = true
  errorMessage.value = ''
  updateProfile()

  try {
    const apiBase = config.public.apiUrl || 'http://localhost:8080'
    const res: any = await $fetch(`${apiBase}/api/rooms`, {
      method: 'POST',
      body: {
        nickname: profileStore.nickname,
        mascot: profileStore.mascot,
        color: profileStore.color,
      },
    })

    if (res?.room?.code && res?.session_token) {
      profileStore.setSession(res.session_token, res.room.code)
      router.push(`/room/${res.room.code}`)
    }
  } catch (err: any) {
    errorMessage.value = err?.data?.error || 'Erreur lors de la création de la salle'
  } finally {
    isLoading.value = false
  }
}

const handleJoinRoom = async () => {
  const code = inputRoomCode.value.trim().toUpperCase()
  if (!code) return

  isLoading.value = true
  errorMessage.value = ''
  updateProfile()

  try {
    const apiBase = config.public.apiUrl || 'http://localhost:8080'
    const res: any = await $fetch(`${apiBase}/api/rooms/${code}/join`, {
      method: 'POST',
      body: {
        nickname: profileStore.nickname,
        mascot: profileStore.mascot,
        color: profileStore.color,
      },
    })

    if (res?.room?.code && res?.session_token) {
      profileStore.setSession(res.session_token, res.room.code)
      router.push(`/room/${res.room.code}`)
    }
  } catch (err: any) {
    errorMessage.value = err?.data?.error || 'Salle introuvable ou fermée'
  } finally {
    isLoading.value = false
  }
}
</script>
