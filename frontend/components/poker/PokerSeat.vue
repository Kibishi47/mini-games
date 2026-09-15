<template>
  <div
    :class="[
      'seat-pop flex flex-col items-center gap-1 p-2 rounded-2xl border-[3px] border-ink-black relative flex-shrink-0 w-[84px]',
      seat.status === 'folded' ? 'seat-folded bg-board-cream' : 'bg-board-white',
      seat.is_turn ? 'seat-active-ring shadow-pop-md' : 'shadow-pop-xs',
      isWinner ? 'seat-winner-glow' : ''
    ]"
    :style="dealt ? { animationDelay: `${seatIndex * 70}ms` } : {}"
  >
    <!-- Rangée de badges (Dealer / Blindes) en flux normal, jamais superposée à l'avatar -->
    <div v-if="seat.is_dealer || seat.is_sb || seat.is_bb" class="flex items-center gap-1">
      <span
        v-if="seat.is_dealer"
        title="Bouton Dealer"
        class="w-4 h-4 rounded-full bg-game-yellow border-2 border-ink-black text-[8px] font-condensed font-black flex items-center justify-center leading-none"
      >D</span>
      <span
        v-if="seat.is_sb"
        title="Petite Blinde"
        class="w-4 h-4 rounded-full bg-board-white border-2 border-ink-black text-[8px] font-condensed font-black flex items-center justify-center leading-none"
      >SB</span>
      <span
        v-if="seat.is_bb"
        title="Grosse Blinde"
        class="w-4 h-4 rounded-full bg-game-blue text-board-white border-2 border-ink-black text-[8px] font-condensed font-black flex items-center justify-center leading-none"
      >BB</span>
    </div>

    <!-- Anneau de décompte du temps d'action, dessiné autour de la mascotte -->
    <div class="relative">
      <svg v-if="seat.is_turn && actionEndsAt" class="turn-ring" viewBox="0 0 40 40">
        <circle cx="20" cy="20" r="17" fill="none" stroke="#12121233" stroke-width="3" />
        <circle
          cx="20" cy="20" r="17" fill="none" stroke="#121212" stroke-width="3"
          stroke-linecap="round"
          :stroke-dasharray="107"
          :stroke-dashoffset="107 * (1 - ringProgress)"
          class="turn-ring-progress"
        />
      </svg>
      <GameMascot :name="(seat.mascot as any) || 'dice'" :mood="seat.is_turn ? 'happy' : 'idle'" size="sm" />
    </div>

    <div class="text-center w-full">
      <div class="font-display font-black text-[11px] uppercase truncate">{{ seat.nickname }}</div>
      <div class="font-condensed font-black text-xs text-game-green">{{ seat.chips }} J</div>
      <span v-if="isMe" class="inline-block mt-0.5 px-1.5 py-0 rounded-full bg-game-blue text-board-white text-[8px] font-condensed font-black uppercase">Vous</span>
    </div>

    <div class="flex gap-1">
      <PlayingCard v-for="(card, i) in displayCards" :key="i" :code="card" size="sm" :index="i" :dealt="dealt" />
    </div>

    <PokerChipStack v-if="seat.bet > 0" :amount="seat.bet" size="sm" class="bet-chip-pop" />

    <AppBadge v-if="seat.status === 'folded'" variant="neutral" class="text-[9px] px-2 py-0">Couché</AppBadge>
    <AppBadge v-else-if="seat.status === 'all_in'" variant="danger" class="text-[9px] px-2 py-0 all-in-pulse">Tapis !</AppBadge>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted, watch } from 'vue'
import GameMascot from '~/components/ui/GameMascot.vue'
import AppBadge from '~/components/ui/AppBadge.vue'
import PlayingCard from '~/components/poker/PlayingCard.vue'
import PokerChipStack from '~/components/poker/PokerChipStack.vue'
import type { PokerSeatView } from '~/stores/poker'

const props = withDefaults(
  defineProps<{
    seat: PokerSeatView
    isMe?: boolean
    isWinner?: boolean
    actionEndsAt?: string
    seatIndex?: number
    dealt?: boolean
  }>(),
  {
    isMe: false,
    isWinner: false,
    seatIndex: 0,
    dealt: false,
  }
)

// Deux cartes visibles si révélées (moi-même ou abattage), cachées si en jeu, aucune si couché (mucked)
const displayCards = computed<(string | null)[]>(() => {
  if (props.seat.hole_cards && props.seat.hole_cards.length) {
    return props.seat.hole_cards
  }
  if (props.seat.status === 'folded') {
    return []
  }
  return [null, null]
})

// Progression du cercle de décompte (1 = plein temps restant, 0 = expiré) basée sur l'horodatage absolu serveur
const ringProgress = ref(1)
let ringInterval: any = null
const ACTION_TIMEOUT_MS = 20000

const updateRing = () => {
  if (!props.actionEndsAt) return
  const remaining = new Date(props.actionEndsAt).getTime() - Date.now()
  ringProgress.value = Math.max(0, Math.min(1, remaining / ACTION_TIMEOUT_MS))
}

onMounted(() => {
  updateRing()
  ringInterval = setInterval(updateRing, 250)
})
onUnmounted(() => {
  if (ringInterval) clearInterval(ringInterval)
})
watch(() => props.actionEndsAt, updateRing)
</script>

<style scoped>
.seat-active-ring {
  outline: 3px solid #ffd300;
  outline-offset: 1px;
}

.seat-folded {
  opacity: 0.5;
  transform: rotate(-2deg) scale(0.94);
}

.turn-ring {
  position: absolute;
  top: -4px;
  left: 50%;
  transform: translateX(-50%);
  width: 40px;
  height: 40px;
  pointer-events: none;
}

.turn-ring-progress {
  transition: stroke-dashoffset 0.25s linear;
  transform: rotate(-90deg);
  transform-origin: 20px 20px;
}

@keyframes allInPulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.12); }
}

.all-in-pulse {
  animation: allInPulse 0.7s ease-in-out infinite;
}

@keyframes seatWinnerGlow {
  0%, 100% { outline-color: #ffd300; }
  50% { outline-color: #158a44; }
}

.seat-winner-glow {
  outline: 3px solid #ffd300;
  outline-offset: 2px;
  animation: seatWinnerGlow 0.9s ease-in-out infinite, seatWinnerBounce 0.5s ease-out;
}

@keyframes seatWinnerBounce {
  0% { transform: scale(1); }
  40% { transform: scale(1.12); }
  100% { transform: scale(1); }
}

.seat-pop {
  transition: outline-color 0.2s ease;
}

.bet-chip-pop {
  animation: betChipPop 0.25s cubic-bezier(0.2, 0.85, 0.25, 1) both;
}

@keyframes betChipPop {
  from { opacity: 0; transform: translateY(-6px) scale(0.6); }
  to { opacity: 1; transform: none; }
}
</style>
