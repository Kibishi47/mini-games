<template>
  <div v-if="amount > 0" :class="['flex items-center gap-1.5', size === 'lg' ? 'gap-2' : '']">
    <div class="relative flex-shrink-0" :style="{ width: `${chipPx}px`, height: `${chipPx + stackCount * chipOffset}px` }">
      <div
        v-for="n in stackCount"
        :key="n"
        :class="['chip-disc absolute left-0 rounded-full border-2 border-ink-black', chipColorClass]"
        :style="{
          width: `${chipPx}px`,
          height: `${chipPx}px`,
          bottom: `${(n - 1) * chipOffset}px`,
        }"
      >
        <span class="chip-ring" />
      </div>
    </div>
    <span :class="['font-condensed font-black uppercase', labelSizeClass]">{{ amount }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    amount: number
    size?: 'sm' | 'md' | 'lg'
  }>(),
  {
    size: 'md',
  }
)

const chipPx = computed(() => (props.size === 'lg' ? 22 : props.size === 'sm' ? 12 : 16))
const chipOffset = computed(() => Math.round(chipPx.value * 0.22))
const labelSizeClass = computed(() => (props.size === 'lg' ? 'text-sm' : props.size === 'sm' ? 'text-[10px]' : 'text-xs'))

// Nombre de jetons visuels empilés selon la valeur (purement décoratif, plafonné pour rester lisible)
const stackCount = computed(() => {
  const a = props.amount
  if (a <= 0) return 0
  if (a < 50) return 1
  if (a < 200) return 2
  if (a < 600) return 3
  return 4
})

// Couleur du jeton façon casino, indexée sur la dénomination (blanc/rouge/bleu/vert/noir)
const chipColorClass = computed(() => {
  const a = props.amount
  if (a < 50) return 'bg-board-white'
  if (a < 200) return 'bg-game-red'
  if (a < 600) return 'bg-game-blue'
  if (a < 2000) return 'bg-game-green'
  return 'bg-ink-black'
})
</script>

<style scoped>
.chip-disc {
  box-shadow: 0 1px 0 rgba(18, 18, 18, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
}

.chip-ring {
  width: 60%;
  height: 60%;
  border-radius: 9999px;
  border: 2px dashed rgba(255, 255, 255, 0.55);
}
</style>
