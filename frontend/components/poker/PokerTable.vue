<template>
  <div class="flex flex-col space-y-4">
    <!-- Scène de table : reset complet (et rejoue les animations de distribution) à chaque nouvelle main -->
    <div :key="table?.hand_num || 0" class="poker-scene">
      <div class="table-ring relative">
        <!-- Feutre (plus petit que la zone totale : laisse la place aux sièges tout autour, sans déborder) -->
        <div class="felt absolute border-[4px] border-ink-black bg-game-green flex flex-col items-center justify-center gap-2.5 px-4">
          <div class="font-condensed font-black text-[11px] uppercase text-board-white/80 tracking-wider">
            Main #{{ table?.hand_num || 0 }} — {{ streetLabel }}
          </div>
          <div class="flex gap-2 min-h-[56px] items-center">
            <PlayingCard
              v-for="(c, i) in communityDisplay"
              :key="i"
              :code="c"
              :placeholder="!c"
              :index="i"
              :dealt="true"
              size="md"
            />
            <span v-if="!table" class="font-body text-xs text-board-white/70">En attente du début de la main...</span>
          </div>
          <PokerChipStack v-if="table?.pot" :amount="table.pot" size="lg" class="pot-chip-pop" />
          <div class="text-[10px] font-condensed uppercase text-board-white/70">
            Blindes {{ table?.small_blind || 0 }} / {{ table?.big_blind || 0 }}
          </div>

          <!-- Éclats de célébration au dénouement d'une main -->
          <div v-if="pokerStore.showHandEnd" class="confetti-burst" aria-hidden="true">
            <span
              v-for="(c, i) in confettiPieces"
              :key="i"
              class="confetti-piece"
              :style="{ '--tx': c.tx, '--ty': c.ty, '--rot': c.rot, background: c.color, animationDelay: `${i * 15}ms` }"
            />
          </div>
        </div>

        <!-- Sièges répartis tout autour de la table, en dehors du feutre -->
        <div
          v-for="sp in seatPositions"
          :key="sp.seat.user_id"
          class="seat-slot absolute"
          :style="sp.style"
        >
          <PokerSeat
            :seat="sp.seat"
            :is-me="sp.seat.user_id === myUserId"
            :is-winner="winnerIds.has(sp.seat.user_id)"
            :action-ends-at="table?.action_ends_at"
            :seat-index="sp.seatIndex"
            :dealt="true"
          />
        </div>
      </div>
    </div>

    <!-- Contrôles d'action (visibles seulement pour le joueur actif) -->
    <PokerControls
      v-if="mySeat && mySeat.status === 'active' && !table?.is_completed"
      :is-my-turn="isMyTurn"
      :call-amount="table?.call_amount || 0"
      :current-bet="table?.current_bet || 0"
      :min-raise="table?.min_raise || 0"
      :big-blind="table?.big_blind || 0"
      :pot="table?.pot || 0"
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
import PokerChipStack from '~/components/poker/PokerChipStack.vue'
import PlayingCard from '~/components/poker/PlayingCard.vue'

const props = defineProps<{
  onAction: (action: string, amount?: number) => void
}>()

const pokerStore = usePokerStore()
const roomStore = useRoomStore()

const table = computed(() => pokerStore.table)

const myUserId = computed(() => roomStore.me?.id)

const mySeat = computed(() => table.value?.seats.find(s => s.user_id === myUserId.value) || null)

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

// Réordonne les sièges pour que "moi" soit toujours affiché en bas de la table, quel que soit l'ordre serveur
const rotatedSeats = computed(() => {
  const seats = table.value?.seats || []
  if (!seats.length) return []
  const meIdx = seats.findIndex(s => s.user_id === myUserId.value)
  if (meIdx <= 0) return seats
  return [...seats.slice(meIdx), ...seats.slice(0, meIdx)]
})

// Positionnement trigonométrique des sièges en anneau AUTOUR du feutre (moi = 90°, plein bas).
// Le rayon reste net en dehors du feutre (cf. .felt en CSS) : les sièges ne chevauchent jamais les cartes/le pot,
// et la marge verticale réservée par .poker-scene absorbe le léger débordement de la carte de siège elle-même.
const seatPositions = computed(() => {
  const seats = rotatedSeats.value
  const n = seats.length
  if (!n) return []
  const rx = 46
  const ry = 43
  return seats.map((seat, i) => {
    const angleDeg = 90 + (360 / n) * i
    const angleRad = (angleDeg * Math.PI) / 180
    const left = 50 + rx * Math.cos(angleRad)
    const top = 50 + ry * Math.sin(angleRad)
    return {
      seat,
      seatIndex: i,
      style: {
        left: `${left}%`,
        top: `${top}%`,
        transform: 'translate(-50%, -50%)',
      },
    }
  })
})

// Identifiants des gagnants de la dernière main, pour l'effet de halo doré sur leur siège
const winnerIds = computed<Set<string>>(() => {
  if (!pokerStore.showHandEnd || !pokerStore.lastHandEnd) return new Set()
  const r = pokerStore.lastHandEnd
  if (r.uncontested && r.winner_id) return new Set([r.winner_id])
  return new Set((r.players || []).filter(p => p.winnings > 0).map(p => p.user_id))
})

const CONFETTI_COLORS = ['#FFD300', '#1D4ED8', '#E63228', '#158A44', '#F27A9B']
const confettiPieces = computed(() => {
  return Array.from({ length: 14 }, (_, i) => {
    const angle = (360 / 14) * i + (i % 2 === 0 ? 6 : -6)
    const rad = (angle * Math.PI) / 180
    const dist = 60 + (i % 3) * 18
    return {
      tx: `${Math.cos(rad) * dist}px`,
      ty: `${Math.sin(rad) * dist}px`,
      rot: `${(i * 47) % 360}deg`,
      color: CONFETTI_COLORS[i % CONFETTI_COLORS.length],
    }
  })
})

const onAction = (payload: { action: string; amount: number }) => {
  props.onAction(payload.action, payload.amount)
}
</script>

<style scoped>
/* Marge verticale généreuse : absorbe le débordement naturel des sièges au-delà de l'anneau
   sans jamais chevaucher le header au-dessus ou les contrôles en dessous. */
.poker-scene {
  width: 100%;
  padding-block: 84px;
}

@media (max-width: 640px) {
  .poker-scene {
    padding-block: 92px;
  }
}

.table-ring {
  width: 100%;
  aspect-ratio: 15 / 9;
  min-height: 300px;
}

@media (max-width: 640px) {
  .table-ring {
    aspect-ratio: 3 / 4;
    min-height: 420px;
  }
}

.felt {
  inset: 24% 13%;
  border-radius: 50% / 44%;
  box-shadow: 6px 6px 0 0 #121212;
}

@media (max-width: 640px) {
  .felt {
    inset: 15% 6%;
    border-radius: 30% / 20%;
  }
}

.seat-slot {
  z-index: 5;
}

.pot-chip-pop {
  animation: potPop 0.3s cubic-bezier(0.2, 0.85, 0.25, 1) both;
}

@keyframes potPop {
  from { opacity: 0; transform: scale(0.6); }
  to { opacity: 1; transform: none; }
}

.confetti-burst {
  position: absolute;
  inset: 0;
  pointer-events: none;
  display: flex;
  align-items: center;
  justify-content: center;
}

.confetti-piece {
  position: absolute;
  width: 8px;
  height: 8px;
  border: 2px solid #121212;
  animation: confettiBurst 0.8s cubic-bezier(0.15, 0.8, 0.3, 1) both;
}

@keyframes confettiBurst {
  0% {
    opacity: 1;
    transform: translate(0, 0) rotate(0deg) scale(1);
  }
  100% {
    opacity: 0;
    transform: translate(var(--tx), var(--ty)) rotate(var(--rot)) scale(0.6);
  }
}
</style>
