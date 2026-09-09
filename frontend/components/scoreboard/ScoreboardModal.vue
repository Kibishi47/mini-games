<template>
  <div v-if="gameStore.showRoundSummary" class="fixed inset-0 bg-ink-black/80 backdrop-blur-none flex items-center justify-center z-50 p-4">
    <div class="bg-board-white max-w-2xl w-full p-6 sm:p-8 rounded-3xl border-[4px] border-ink-black shadow-pop-lg relative flex flex-col space-y-6">
      
      <!-- Titre et Mot Secret Révélé -->
      <div class="text-center space-y-2">
        <div class="inline-flex items-center space-x-2 border-2 border-ink-black px-3 py-1 rounded-full bg-game-pink font-condensed text-xs uppercase mb-1">
          <Trophy class="w-4 h-4 text-ink-black" />
          <span>{{ gameStore.isGameOver ? 'Palmarès Final' : `Fin Manche ${gameStore.currentRound}/${gameStore.maxRounds}` }}</span>
        </div>

        <h3 class="font-display font-black text-3xl sm:text-4xl uppercase tracking-tight text-ink-black">
          {{ gameStore.isGameOver ? '🏆 Victoire au Sommet !' : 'Tour Terminé !' }}
        </h3>
        
        <p class="font-display font-bold uppercase text-xs text-ink-black/60">Le mot secret était :</p>
        <div class="inline-block px-6 py-2.5 rounded-2xl bg-game-yellow border-[3px] border-ink-black text-ink-black font-condensed font-black text-3xl tracking-widest shadow-pop-sm">
          {{ gameStore.targetWord }}
        </div>
      </div>

      <!-- Piste de Score de Jeu de Société -->
      <div class="space-y-3 max-h-72 overflow-y-auto pr-1">
        <div
          v-for="(summary, index) in sortedSummaries"
          :key="summary.user_id"
          :class="[
            'flex items-center justify-between p-3.5 rounded-2xl border-[3px] border-ink-black shadow-pop-xs transition-none',
            index === 0 ? 'bg-game-yellow/30' : 'bg-board-cream'
          ]"
        >
          <!-- Rang, Mascotte et Nom -->
          <div class="flex items-center space-x-3">
            <span
              :class="[
                'w-8 h-8 rounded-xl border-2 border-ink-black font-condensed font-black text-sm flex items-center justify-center shadow-pop-xs',
                index === 0 ? 'bg-game-yellow text-ink-black' :
                index === 1 ? 'bg-board-white text-ink-black' :
                index === 2 ? 'bg-game-pink text-ink-black' : 'bg-board-cream text-ink-black/60'
              ]"
            >
              #{{ index + 1 }}
            </span>

            <GameMascot :name="getMascotName(index)" :mood="summary.is_solved ? 'happy' : 'idle'" size="md" />

            <div>
              <div class="font-display font-black text-sm uppercase text-ink-black">{{ summary.display_username }}</div>
              <div class="font-condensed text-xs uppercase text-ink-black/70">
                {{ summary.is_solved ? `Trouvé en ${summary.attempts_count} essais` : 'Non résolu' }}
              </div>
            </div>
          </div>

          <!-- Score Delta & Total sur la piste -->
          <div class="text-right">
            <div class="text-game-green font-condensed font-black text-lg">
              +{{ summary.score_delta }} pts
            </div>
            <div class="text-xs font-display font-bold uppercase text-ink-black/60">
              Total : {{ summary.total_score }} pts
            </div>
          </div>
        </div>
      </div>

      <!-- Actions : Copier Émojis & Revanche Master -->
      <div class="flex flex-col sm:flex-row gap-3 pt-2">
        <AppButton
          variant="neutral"
          size="md"
          class="flex-1"
          @click="copyEmojiGrid"
        >
          <Share2 class="w-5 h-5 mr-2" />
          <span>{{ copied ? 'Copié ! 🟩🟨⬛' : 'Partager mes émojis 🟩🟨⬛' }}</span>
        </AppButton>

        <AppButton
          v-if="gameStore.isGameOver && isMaster"
          variant="success"
          size="md"
          class="flex-1"
          @click="onRematch"
        >
          <RotateCcw class="w-5 h-5 mr-2" />
          <span>Lancer la Revanche 🔄</span>
        </AppButton>
      </div>

      <div v-if="!gameStore.isGameOver" class="text-center font-display font-bold text-xs uppercase text-ink-black/60">
        La manche suivante commence dans un instant...
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { Trophy, Share2, RotateCcw } from 'lucide-vue-next'
import { useGameStore } from '~/stores/game'
import { useAuthStore } from '~/stores/auth'
import GameMascot, { type MascotName } from '~/components/ui/GameMascot.vue'
import AppButton from '~/components/ui/AppButton.vue'

defineProps<{
  isMaster: boolean
  onRematch: () => void
}>()

const gameStore = useGameStore()
const authStore = useAuthStore()
const copied = ref(false)

const mascotPool: MascotName[] = ['dice', 'knight', 'card', 'domino', 'd20', 'meeple']
const getMascotName = (idx: number): MascotName => mascotPool[idx % mascotPool.length]

const sortedSummaries = computed(() => {
  return [...gameStore.roundSummaries].sort((a, b) => b.total_score - a.total_score)
})

const copyEmojiGrid = async () => {
  const mySummary = gameStore.roundSummaries.find(s => s.user_id === authStore.user?.id)
  const grid = mySummary?.emoji_grid || '🟩🟨⬛'
  const text = `MiniGames Wordle Multijoueur\nManche ${gameStore.currentRound}/${gameStore.maxRounds}\n\n${grid}\nFestival du Jeu de Société !`
  
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
