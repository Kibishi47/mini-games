<template>
  <div v-if="gameStore.showRoundSummary" class="fixed inset-0 bg-black/80 backdrop-blur-md flex items-center justify-center z-50 p-4 animate-in fade-in duration-300">
    <div class="glass-panel max-w-xl w-full p-6 sm:p-8 rounded-3xl border border-white/10 shadow-2xl relative flex flex-col space-y-6">
      
      <!-- Titre et Mot Secret -->
      <div class="text-center space-y-2">
        <h3 class="text-2xl sm:text-3xl font-black tracking-tight text-white">
          {{ gameStore.isGameOver ? '🏆 Partie Terminée !' : `Fin de la Manche ${gameStore.currentRound}/${gameStore.maxRounds}` }}
        </h3>
        <p class="text-slate-400 text-sm">Le mot secret à trouver était :</p>
        <div class="inline-block px-5 py-2 rounded-2xl bg-indigo-500/20 border border-indigo-500/30 text-indigo-300 font-extrabold text-2xl tracking-widest shadow-inner">
          {{ gameStore.targetWord }}
        </div>
      </div>

      <!-- Classement de la Manche & Cumul -->
      <div class="space-y-3 max-h-64 overflow-y-auto pr-1">
        <div
          v-for="(summary, index) in sortedSummaries"
          :key="summary.user_id"
          class="flex items-center justify-between p-3.5 rounded-2xl bg-white/5 border border-white/5"
        >
          <div class="flex items-center space-x-3">
            <span class="w-6 font-bold text-sm text-center" :class="getRankColor(index)">
              #{{ index + 1 }}
            </span>
            <img
              :src="summary.avatar_url || 'https://api.dicebear.com/7.x/bottts/svg?seed=' + summary.display_username"
              class="w-10 h-10 rounded-xl bg-slate-800 border border-white/10 p-0.5"
            />
            <div>
              <div class="font-bold text-sm text-white">{{ summary.display_username }}</div>
              <div class="text-xs text-slate-400">
                {{ summary.is_solved ? `Trouvé en ${summary.attempts_count} essais` : 'Non trouvé' }}
              </div>
            </div>
          </div>

          <div class="text-right">
            <div class="text-emerald-400 font-black text-base">
              +{{ summary.score_delta }} pts
            </div>
            <div class="text-xs text-slate-500 font-medium">
              Total : {{ summary.total_score }} pts
            </div>
          </div>
        </div>
      </div>

      <!-- Actions : Copier Résultat Émojis & Revanche -->
      <div class="flex flex-col sm:flex-row gap-3 pt-2">
        <button
          @click="copyEmojiGrid"
          class="flex-1 flex items-center justify-center space-x-2 py-3.5 px-4 rounded-xl bg-white/10 hover:bg-white/15 text-white font-semibold transition-all border border-white/10 shadow-lg"
        >
          <Share2 class="w-5 h-5 text-indigo-400" />
          <span>{{ copied ? 'Copié dans le presse-papier !' : 'Partager mes émojis 🟩🟨⬛' }}</span>
        </button>

        <button
          v-if="gameStore.isGameOver && isMaster"
          @click="onRematch"
          class="flex-1 flex items-center justify-center space-x-2 py-3.5 px-4 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white font-bold transition-all shadow-lg shadow-indigo-600/40 hover:scale-[1.02] active:scale-[0.98]"
        >
          <RotateCcw class="w-5 h-5" />
          <span>Lancer la Revanche 🔄</span>
        </button>
      </div>

      <div v-if="!gameStore.isGameOver" class="text-center text-xs text-slate-500 italic">
        Prochaine manche dans quelques secondes...
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { Share2, RotateCcw } from 'lucide-vue-next'
import { useGameStore } from '~/stores/game'
import { useAuthStore } from '~/stores/auth'

const props = defineProps<{
  isMaster: boolean
  onRematch: () => void
}>()

const gameStore = useGameStore()
const authStore = useAuthStore()
const copied = ref(false)

const sortedSummaries = computed(() => {
  return [...gameStore.roundSummaries].sort((a, b) => b.total_score - a.total_score)
})

const getRankColor = (idx: number) => {
  if (idx === 0) return 'text-amber-400 text-lg'
  if (idx === 1) return 'text-slate-300'
  if (idx === 2) return 'text-amber-700'
  return 'text-slate-500'
}

const copyEmojiGrid = async () => {
  const mySummary = gameStore.roundSummaries.find(s => s.user_id === authStore.user?.id)
  const grid = mySummary?.emoji_grid || '🟩🟨⬛'
  const text = `MiniGames Wordle Multijoueur\nManche ${gameStore.currentRound}/${gameStore.maxRounds}\n\n${grid}\nRejoins la partie sur MiniGames !`
  
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
