<template>
  <div
    :class="[
      'rounded-lg flex items-center justify-center select-none flex-shrink-0',
      sizeClasses,
      code ? 'border-[3px] border-ink-black bg-board-white shadow-pop-xs' :
        placeholder ? 'border-2 border-dashed border-ink-black/25 bg-transparent' :
        'border-[3px] border-ink-black bg-game-blue shadow-pop-xs'
    ]"
  >
    <div v-if="code" :class="['flex flex-col items-center leading-none', colorClass]">
      <span :class="['font-condensed font-black', rankSizeClass]">{{ rankLabel }}</span>
      <span :class="suitSizeClass">{{ suitSymbol }}</span>
    </div>
    <div v-else-if="!placeholder" class="w-2/3 h-2/3 rounded-md border-2 border-board-white/50" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    code?: string | null // ex: "AS" (As de Pique), "TD" (10 de Carreau) — null/undefined = carte cachée
    size?: 'sm' | 'md' | 'lg'
    placeholder?: boolean // true = emplacement vide en pointillés (carte pas encore distribuée) plutôt qu'un dos de carte
  }>(),
  {
    code: null,
    size: 'md',
    placeholder: false,
  }
)

const rankLabel = computed(() => {
  if (!props.code) return ''
  const r = props.code[0]
  return r === 'T' ? '10' : r
})

const suitSymbol = computed(() => {
  if (!props.code) return ''
  switch (props.code[1]) {
    case 'S': return '♠'
    case 'H': return '♥'
    case 'D': return '♦'
    case 'C': return '♣'
    default: return ''
  }
})

const colorClass = computed(() => {
  if (!props.code) return ''
  return props.code[1] === 'H' || props.code[1] === 'D' ? 'text-game-red' : 'text-ink-black'
})

const sizeClasses = computed(() => {
  switch (props.size) {
    case 'sm': return 'w-8 h-11'
    case 'lg': return 'w-14 h-20'
    default: return 'w-10 h-14'
  }
})

const rankSizeClass = computed(() => (props.size === 'lg' ? 'text-xl' : props.size === 'sm' ? 'text-xs' : 'text-sm'))
const suitSizeClass = computed(() => (props.size === 'lg' ? 'text-2xl' : props.size === 'sm' ? 'text-sm' : 'text-lg'))
</script>
