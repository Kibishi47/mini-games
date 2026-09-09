<template>
  <div class="glass-panel p-4 rounded-2xl border border-white/10 flex flex-col space-y-4">
    <div class="flex items-center justify-between pb-3 border-b border-white/10">
      <div class="flex items-center space-x-2">
        <Users class="w-5 h-5 text-indigo-400" />
        <span class="font-semibold text-sm text-white">Adversaires en direct</span>
      </div>
      <span class="text-xs px-2.5 py-0.5 rounded-full bg-white/5 text-slate-400">
        {{ opponents.length }} joueur(s)
      </span>
    </div>

    <!-- Grille compacte des adversaires -->
    <div class="grid grid-cols-1 gap-3 overflow-y-auto max-h-[500px] pr-1">
      <div
        v-for="opp in opponents"
        :key="opp.user_id"
        class="glass-card p-3 rounded-xl border border-white/5 flex flex-col space-y-2 relative overflow-hidden"
      >
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-2">
            <span class="font-medium text-xs text-slate-200 truncate max-w-[120px]">
              {{ opp.display_username }}
            </span>
          </div>
          
          <div class="flex items-center space-x-1">
            <span
              v-if="opp.is_solved"
              class="flex items-center text-[11px] font-semibold text-emerald-400 bg-emerald-500/10 px-2 py-0.5 rounded-full"
            >
              <CheckCircle2 class="w-3 h-3 mr-1" /> Trouvé
            </span>
            <span
              v-else-if="opp.is_finished"
              class="text-[11px] font-medium text-rose-400 bg-rose-500/10 px-2 py-0.5 rounded-full"
            >
              Échoué
            </span>
            <span
              v-else
              class="text-[11px] text-slate-400"
            >
              Ligne {{ (opp.masked_rows || []).length }}/{{ maxAttempts }}
            </span>
          </div>
        </div>

        <!-- Mini Tuiles Masquées (Sans lettres, uniquement les couleurs !) -->
        <div class="flex flex-col gap-1 items-center bg-black/20 p-2 rounded-lg">
          <div
            v-for="(row, rIdx) in (opp.masked_rows || [])"
            :key="rIdx"
            class="flex gap-1"
          >
            <div
              v-for="(tile, cIdx) in row"
              :key="cIdx"
              :class="[
                'w-5 h-5 rounded-md transition-colors',
                tile.status === 'correct' ? 'bg-emerald-500 shadow-sm shadow-emerald-500/50' :
                tile.status === 'present' ? 'bg-amber-400 shadow-sm shadow-amber-400/50' :
                'bg-slate-700'
              ]"
            />
          </div>
          <div v-if="!opp.masked_rows || opp.masked_rows.length === 0" class="text-[10px] text-slate-500 py-2 italic">
            En réflexion...
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Users, CheckCircle2 } from 'lucide-vue-next'
import type { OpponentProgress } from '~/stores/game'

defineProps<{
  opponents: OpponentProgress[]
  maxAttempts: number
}>()
</script>
