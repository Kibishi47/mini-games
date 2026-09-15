<template>
  <div v-if="placeholder" :class="['rounded-lg border-2 border-dashed border-ink-black/25 bg-transparent flex-shrink-0', sizeClasses]" />
  <div
    v-else
    :class="['card-3d flex-shrink-0 select-none', sizeClasses, dealt ? 'card-deal-in' : '']"
    :style="dealt ? { animationDelay: `${Math.max(index, 0) * 80}ms` } : {}"
  >
    <div :class="['card-3d-inner', code ? 'is-flipped' : '']">
      <!-- Dos de carte (motif) -->
      <div class="card-face card-back-face border-[3px] border-ink-black bg-game-blue shadow-pop-xs rounded-lg">
        <div class="card-back-pattern">
          <span v-for="n in 4" :key="n" class="card-back-dot" />
        </div>
      </div>
      <!-- Face visible (rang + couleur) -->
      <div class="card-face card-front-face border-[3px] border-ink-black bg-board-white shadow-pop-xs rounded-lg flex items-center justify-center">
        <div v-if="code" :class="['flex flex-col items-center leading-none', colorClass]">
          <span :class="['font-condensed font-black', rankSizeClass]">{{ rankLabel }}</span>
          <span :class="suitSizeClass">{{ suitSymbol }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    code?: string | null // ex: "AS" (As de Pique), "TD" (10 de Carreau) — null/undefined = carte cachée
    size?: 'sm' | 'md' | 'lg'
    placeholder?: boolean // true = emplacement vide en pointillés (carte pas encore distribuée) plutôt qu'un dos de carte
    index?: number // position dans la distribution, pilote le décalage d'animation "dealing"
    dealt?: boolean // true = joue l'animation d'entrée façon distribution de croupier
  }>(),
  {
    code: null,
    size: 'md',
    placeholder: false,
    index: 0,
    dealt: false,
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

<style scoped>
.card-3d {
  perspective: 500px;
}

.card-3d-inner {
  position: relative;
  width: 100%;
  height: 100%;
  transform-style: preserve-3d;
  transition: transform 0.4s cubic-bezier(0.2, 0.85, 0.25, 1);
}

.card-3d-inner.is-flipped {
  transform: rotateY(180deg);
}

.card-face {
  position: absolute;
  inset: 0;
  backface-visibility: hidden;
  -webkit-backface-visibility: hidden;
}

.card-back-face {
  transform: rotateY(0deg);
}

.card-front-face {
  transform: rotateY(180deg);
}

.card-back-pattern {
  width: 100%;
  height: 100%;
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  grid-template-rows: repeat(2, 1fr);
  place-items: center;
  padding: 15%;
}

.card-back-dot {
  width: 35%;
  height: 35%;
  border-radius: 9999px;
  background: rgba(255, 255, 255, 0.55);
}

@keyframes cardDealIn {
  from {
    opacity: 0;
    transform: translateY(-18px) scale(0.5) rotate(-10deg);
  }
  to {
    opacity: 1;
    transform: none;
  }
}

.card-deal-in {
  animation: cardDealIn 0.32s cubic-bezier(0.2, 0.85, 0.25, 1) both;
}
</style>
