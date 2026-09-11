<template>
  <div class="flex flex-col space-y-4">
    <!-- Sièges adverses -->
    <div v-if="opponentSeats.length" class="flex gap-3 overflow-x-auto pb-1">
      <PokerSeat v-for="seat in opponentSeats" :key="seat.user_id" :seat="seat" />
    </div>

    <!-- Plateau central : cartes communes, rue en cours et pot -->
    <div class="bg-game-green/10 border-[3px] border-ink-black rounded-3xl p-5 flex flex-col items-center gap-3">
      <div class="font-condensed font-black text-[11px] uppercase text-ink-black/60 tracking-wider">
        Main #{{ table?.hand_num || 0 }} — {{ streetLabel }}
      </div>
      <div class="flex gap-2 min-h-[56px] items-center">
        <PlayingCard v-for="(c, i) in communityDisplay" :key="i" :code="c" :placeholder="!c" size="md" />
        <span v-if="!table" class="font-body text-xs text-ink-black/50">En attente du début de la main...</span>
      </div>
      <div class="px-4 py-1.5 rounded-full bg-game-yellow border-2 border-ink-black font-display font-black text-sm uppercase shadow-pop-xs">
        Pot : {{ table?.pot || 0 }} jetons
      </div>
      <div class="text-[10px] font-condensed uppercase text-ink-black/50">
        Blindes {{ table?.small_blind || 0 }} / {{ table?.big_blind || 0 }}
      </div>
    </div>

    <!-- Mon siège -->
    <div v-if="mySeat" class="flex justify-center">
      <PokerSeat :seat="mySeat" />
    </div>

    <!-- Contrôles d'action (visibles seulement pour le joueur actif) -->
    <PokerControls
      v-if="mySeat && mySeat.status === 'active' && !table?.is_completed"
      :is-my-turn="isMyTurn"
      :call-amount="table?.call_amount || 0"
      :current-bet="table?.current_bet || 0"
      :min-raise="table?.min_raise || 0"
      :big-blind="table?.big_blind || 0"
      :my-chips="mySeat.chips"
      :my-committed="mySeat.bet"
      :action-ends-at="table?.action_ends_at"
      @action="onAction"
    />

    <!-- Bandeau de résultat de main (abattage ou victoire par tapis non contesté) -->
    <div
      v-if="pokerStore.showHandEnd && pokerStore.lastHandEnd"
      class="border-[3px] border-ink-black bg-board-white rounded-2xl p-4 shadow-pop-md space-y-3"
    >
      <div class="flex items-center justify-between border-b-2 border-ink-black pb-2">
        <span class="font-display font-black text-sm uppercase flex items-center gap-2">
          <Trophy class="w-4 h-4 text-game-yellow fill-game-yellow" />
          Résultat de la Main #{{ pokerStore.lastHandEnd.hand_num }}
        </span>
        <button @click="pokerStore.dismissHandEnd()" class="p-1 hover:text-game-red">
          <X class="w-4 h-4" />
        </button>
      </div>

      <div v-if="pokerStore.lastHandEnd.uncontested" class="text-center font-body text-sm">
        <strong>{{ pokerStore.lastHandEnd.winner_name }}</strong> remporte <strong>{{ pokerStore.lastHandEnd.amount }} jetons</strong> — tous les autres joueurs se sont couchés.
      </div>

      <div v-else class="space-y-2">
        <div
          v-for="p in (pokerStore.lastHandEnd.players || []).filter(p => !p.folded)"
          :key="p.user_id"
          class="flex items-center justify-between p-2 rounded-xl border-2 border-ink-black bg-board-cream"
        >
          <div class="flex items-center gap-2">
            <div class="flex gap-1">
              <PlayingCard v-for="(c, i) in p.hole_cards || []" :key="i" :code="c" size="sm" />
            </div>
            <div>
              <div class="font-display font-black text-xs uppercase">{{ p.nickname }}</div>
              <div class="text-[10px] font-condensed uppercase text-ink-black/60">{{ p.category }}</div>
            </div>
          </div>
          <div :class="['font-condensed font-black text-sm', p.winnings > 0 ? 'text-game-green' : 'text-ink-black/40']">
            {{ p.winnings > 0 ? `+${p.winnings}` : '—' }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Trophy, X } from 'lucide-vue-next'
import { usePokerStore } from '~/stores/poker'
import { useRoomStore } from '~/stores/room'
import PokerSeat from '~/components/poker/PokerSeat.vue'
import PokerControls from '~/components/poker/PokerControls.vue'
import PlayingCard from '~/components/poker/PlayingCard.vue'

const props = defineProps<{
  onAction: (action: string, amount?: number) => void
}>()

const pokerStore = usePokerStore()
const roomStore = useRoomStore()

const table = computed(() => pokerStore.table)

const myUserId = computed(() => roomStore.me?.id)

const mySeat = computed(() => table.value?.seats.find(s => s.user_id === myUserId.value) || null)

const opponentSeats = computed(() => (table.value?.seats || []).filter(s => s.user_id !== myUserId.value))

const isMyTurn = computed(() => !!mySeat.value && table.value?.to_act === myUserId.value && !table.value?.is_completed)

const communityDisplay = computed<(string | null)[]>(() => {
  const community = table.value?.community || []
  const padded = [...community]
  while (padded.length < 5) padded.push(null as any)
  return padded
})

const streetLabel = computed(() => {
  switch (table.value?.street) {
    case 'preflop': return 'Pré-Flop'
    case 'flop': return 'Flop'
    case 'turn': return 'Turn'
    case 'river': return 'River'
    case 'showdown': return 'Abattage'
    default: return 'En attente'
  }
})

const onAction = (payload: { action: string; amount: number }) => {
  props.onAction(payload.action, payload.amount)
}
</script>
