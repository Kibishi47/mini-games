<template>
  <AppCard variant="white" shadow="md" class="flex flex-col space-y-3">
    <div class="flex items-center justify-between pb-2 border-b-2 border-ink-black">
      <div class="flex items-center space-x-2">
        <Users class="w-4 h-4 text-game-blue" />
        <span class="font-display font-black text-xs uppercase text-ink-black tracking-wider">Adversaires en Direct</span>
      </div>
      <AppBadge variant="master">{{ opponents.length }}</AppBadge>
    </div>

    <!-- Mini Panneaux Latéraux des Concurrents -->
    <div class="grid grid-cols-1 gap-2.5 overflow-y-auto max-h-[380px] pr-1">
      <div
        v-for="opp in opponents"
        :key="opp.user_id"
        class="border-2 border-ink-black rounded-xl p-2.5 bg-board-cream flex flex-col space-y-1.5 shadow-pop-xs"
      >
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-2">
            <GameMascot :name="(opp.mascot as any) || 'dice'" mood="idle" size="sm" />
            <span class="font-display font-black text-xs uppercase text-ink-black truncate max-w-[120px]">
              {{ opp.nickname }}
            </span>
          </div>
          
          <div>
            <AppBadge v-if="opp.is_solved" variant="success">Trouvé !</AppBadge>
            <AppBadge v-else-if="opp.is_finished" variant="danger">Échoué</AppBadge>
            <span v-else class="font-condensed font-bold text-[10px] uppercase text-ink-black/60">
              Essai {{ (opp.masked_rows || []).length }}/{{ maxAttempts }}
            </span>
          </div>
        </div>

        <!-- Mini Tuiles Masquées (State Masking Strict : zéro lettre, seulement les couleurs) -->
        <div class="flex flex-col gap-1 items-center bg-board-white p-2 rounded-lg border border-ink-black">
          <div
            v-for="(row, rIdx) in (opp.masked_rows || [])"
            :key="rIdx"
            class="flex gap-1"
          >
            <div
              v-for="(tile, cIdx) in row"
              :key="cIdx"
              :class="[
                'w-4 h-4 rounded-md border border-ink-black',
                tile.status === 'correct' ? 'bg-game-green' :
                tile.status === 'present' ? 'bg-game-yellow' :
                'pattern-hatch'
              ]"
            />
          </div>
          <div v-if="!opp.masked_rows || opp.masked_rows.length === 0" class="text-[10px] font-bold text-ink-black/40 py-1 uppercase">
            En réflexion...
          </div>
        </div>
      </div>
    </div>
  </AppCard>
</template>

<script setup lang="ts">
import { Users } from 'lucide-vue-next'
import AppCard from '~/components/ui/AppCard.vue'
import AppBadge from '~/components/ui/AppBadge.vue'
import GameMascot from '~/components/ui/GameMascot.vue'
import type { OpponentProgress } from '~/stores/game'

defineProps<{
  opponents: OpponentProgress[]
  maxAttempts: number
}>()
</script>
