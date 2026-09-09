<template>
  <div v-if="isOpen" class="fixed inset-0 bg-black/80 backdrop-blur-md flex items-center justify-center z-50 p-4 animate-in fade-in duration-200">
    <div class="glass-panel max-w-md w-full p-6 sm:p-8 rounded-3xl border border-white/10 shadow-2xl relative">
      <!-- Close button -->
      <button
        @click="close"
        class="absolute top-5 right-5 p-2 text-slate-400 hover:text-white rounded-xl hover:bg-white/5 transition-colors"
      >
        <X class="w-5 h-5" />
      </button>

      <!-- Tabs : Connexion / Inscription -->
      <div class="flex p-1.5 rounded-2xl bg-black/40 border border-white/5 mb-6 mt-2">
        <button
          @click="mode = 'login'"
          :class="[
            'flex-1 py-2.5 rounded-xl font-bold text-sm transition-all flex items-center justify-center space-x-2',
            mode === 'login' ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-600/30' : 'text-slate-400 hover:text-white'
          ]"
        >
          <LogIn class="w-4 h-4" />
          <span>Connexion</span>
        </button>

        <button
          @click="mode = 'register'"
          :class="[
            'flex-1 py-2.5 rounded-xl font-bold text-sm transition-all flex items-center justify-center space-x-2',
            mode === 'register' ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-600/30' : 'text-slate-400 hover:text-white'
          ]"
        >
          <UserPlus class="w-4 h-4" />
          <span>Inscription</span>
        </button>
      </div>

      <!-- FORMULAIRE LOCAL (Login ou Register) -->
      <form @submit.prevent="handleSubmit" class="space-y-4">
        <div>
          <label class="block text-xs font-bold uppercase tracking-wider text-slate-400 mb-1.5">
            Nom d'utilisateur
          </label>
          <input
            v-model="form.username"
            type="text"
            required
            placeholder="Pseudo"
            class="w-full bg-brand-surface/70 border border-white/10 rounded-xl px-4 py-3 text-sm text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 transition-all"
          />
        </div>

        <div v-if="mode === 'register'">
          <label class="block text-xs font-bold uppercase tracking-wider text-slate-400 mb-1.5">
            Pseudo d'affichage (optionnel)
          </label>
          <input
            v-model="form.display_username"
            type="text"
            placeholder="Ex: MonSuperPseudo"
            class="w-full bg-brand-surface/70 border border-white/10 rounded-xl px-4 py-3 text-sm text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 transition-all"
          />
        </div>

        <div v-if="mode === 'register'">
          <label class="block text-xs font-bold uppercase tracking-wider text-slate-400 mb-1.5">
            Email (optionnel)
          </label>
          <input
            v-model="form.email"
            type="email"
            placeholder="email@exemple.com"
            class="w-full bg-brand-surface/70 border border-white/10 rounded-xl px-4 py-3 text-sm text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 transition-all"
          />
        </div>

        <div>
          <label class="block text-xs font-bold uppercase tracking-wider text-slate-400 mb-1.5">
            Mot de passe
          </label>
          <input
            v-model="form.password"
            type="password"
            required
            placeholder="••••••••"
            class="w-full bg-brand-surface/70 border border-white/10 rounded-xl px-4 py-3 text-sm text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 transition-all"
          />
        </div>

        <!-- Erreur -->
        <div v-if="errorMessage" class="p-3 rounded-xl bg-rose-500/10 border border-rose-500/20 text-rose-400 text-xs font-medium text-center">
          {{ errorMessage }}
        </div>

        <button
          type="submit"
          :disabled="isLoading"
          class="w-full py-3.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white font-bold text-sm shadow-xl shadow-indigo-600/30 transition-all flex items-center justify-center space-x-2 mt-2"
        >
          <span>{{ isLoading ? 'Chargement...' : mode === 'login' ? 'Se connecter' : 'Créer mon compte' }}</span>
        </button>
      </form>

      <!-- Séparateur -->
      <div class="relative flex items-center justify-center my-6">
        <div class="border-t border-white/10 w-full" />
        <span class="bg-[#111827] px-3 text-xs text-slate-500 uppercase tracking-widest absolute">ou</span>
      </div>

      <!-- Bouton Discord OAuth2 en dessous du formulaire -->
      <button
        @click="handleDiscordLogin"
        :disabled="isLoading"
        class="w-full py-3 px-4 rounded-2xl bg-[#5865F2] hover:bg-[#4752C4] disabled:opacity-50 text-white font-bold text-sm transition-all flex items-center justify-center space-x-2 shadow-lg shadow-[#5865F2]/25"
      >
        <svg class="w-5 h-5 fill-current" viewBox="0 0 24 24">
          <path d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0 12.64 12.64 0 0 0-.617-1.25.077.077 0 0 0-.079-.037A19.736 19.736 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 0 0 .031.057 19.9 19.9 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028c.462-.63.874-1.295 1.226-1.994.021-.041.001-.09-.041-.106a13.107 13.107 0 0 1-1.872-.892.077.077 0 0 1-.008-.128 10.2 10.2 0 0 0 .372-.292.074.074 0 0 1 .077-.01c3.929 1.793 8.18 1.793 12.061 0a.074.074 0 0 1 .078.01c.12.098.246.198.373.292a.077.077 0 0 1-.006.127 12.299 12.299 0 0 1-1.873.894.077.077 0 0 0-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 0 0 .084.028 19.839 19.839 0 0 0 6.002-3.03.077.077 0 0 0 .032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.028zM8.02 15.33c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.956-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.956 2.418-2.157 2.418zm7.975 0c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.955-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.946 2.418-2.157 2.418z"/>
        </svg>
        <span>Continuer avec Discord</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { X, LogIn, UserPlus } from 'lucide-vue-next'
import { useAuthStore } from '~/stores/auth'

const props = defineProps<{
  isOpen: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const authStore = useAuthStore()
const config = useRuntimeConfig()

const mode = ref<'login' | 'register'>('login')
const isLoading = ref(false)
const errorMessage = ref<string | null>(null)

const form = reactive({
  username: '',
  display_username: '',
  email: '',
  password: '',
})

const close = () => {
  errorMessage.value = null
  emit('close')
}

const handleSubmit = async () => {
  errorMessage.value = null
  isLoading.value = true
  try {
    if (mode.value === 'login') {
      await authStore.loginLocal(form.username, form.password)
    } else {
      await authStore.registerLocal(form.username, form.display_username, form.email, form.password)
    }
    close()
  } catch (err: any) {
    errorMessage.value = err.data?.error || err.message || 'Erreur lors de la connexion'
  } finally {
    isLoading.value = false
  }
}

const handleDiscordLogin = async () => {
  try {
    const res = await $fetch<{ url: string }>(`${config.public.apiUrl}/api/auth/discord/login`)
    if (res.url) {
      window.location.href = res.url
    }
  } catch (err: any) {
    errorMessage.value = err.data?.error || 'Discord OAuth non configuré'
  }
}
</script>
