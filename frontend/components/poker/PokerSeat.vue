<template>
  <div
    :class="[
      'flex flex-col items-center gap-1.5 p-3 rounded-2xl border-[3px] border-ink-black relative transition-none flex-shrink-0 w-[104px]',
      seat.status === 'folded' ? 'bg-board-cream opacity-50' : 'bg-board-white',
      seat.is_turn ? 'shadow-pop-md ring-4 ring-game-yellow' : 'shadow-pop-xs'
    ]"
  >
    <!-- Badges Bouton Dealer / Blindes -->
    <div class="absolute -top-2 -right-2 flex gap-1 z-10">
      <span
        v-if="seat.is_dealer"
        title="Bouton Dealer"
        class="w-5 h-5 rounded-full bg-game-yellow border-2 border-ink-black text-[9px] font-condensed font-black flex items-center justify-center"
      >D</span>
      <span
        v-if="seat.is_sb"
        title="Petite Blinde"
        class="w-5 h-5 rounded-full bg-board-white border-2 border-ink-black text-[9px] font-condensed font-black flex items-center justify-center"
      >SB</span>
      <span
        v-if="seat.is_bb"
        title="Grosse Blinde"
        class="w-5 h-5 rounded-full bg-game-blue text-board-white border-2 border-ink-black text-[9px] font-condensed font-black flex items-center justify-center"
      >BB</span>
    </div>

    <GameMascot :name="(seat.mascot as any) || 'dice'" :mood="seat.is_turn ? 'happy' : 'idle'" size="sm" />

    <div class="text-center w-full">
      <div class="font-display font-black text-[11px] uppercase truncate">{{ seat.nickname }}</div>
      <div class="font-condensed font-black text-xs text-game-green">{{ seat.chips }} J</div>
    </div>

    <div class="flex gap-1">
      <PlayingCard v-for="(card, i) in displayCards" :key="i" :code="card" size="sm" />
    </div>

    <div v-if="seat.bet > 0" class="px-2 py-0.5 rounded-full bg-game-yellow border-2 border-ink-black text-[10px] font-condensed font-black">
      {{ seat.bet }}
    </div>

    <AppBadge v-if="seat.status === 'folded'" variant="neutral" class="text-[9px] px-2 py-0">Couché</AppBadge>
    <AppBadge v-else-if="seat.status === 'all_in'" variant="danger" class="text-[9px] px-2 py-0">Tapis !</AppBadge>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import GameMascot from '~/components/ui/GameMascot.vue'
import AppBadge from '~/components/ui/AppBadge.vue'
import PlayingCard from '~/components/poker/PlayingCard.vue'
import type { PokerSeatView } from '~/stores/poker'

const props = defineProps<{ seat: PokerSeatView }>()

// Deux cartes visibles si révélées (moi-même ou abattage), cachées si en jeu, aucune si coupé (mucked)
const displayCards = computed<(string | null)[]>(() => {
  if (props.seat.hole_cards && props.seat.hole_cards.length) {
    return props.seat.hole_cards
  }
  if (props.seat.status === 'folded') {
    return []
  }
  return [null, null]
})
</script>
