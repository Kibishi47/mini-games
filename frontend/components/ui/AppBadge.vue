<template>
  <span
    :class="[
      'inline-flex items-center space-x-1.5 border-2 border-ink-black rounded-full px-3 py-0.5 text-xs font-condensed uppercase tracking-wider select-none',
      variantClasses
    ]"
  >
    <slot />
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    variant?: 'master' | 'spectator' | 'system' | 'danger' | 'success' | 'neutral'
  }>(),
  {
    variant: 'neutral',
  }
)

const variantClasses = computed(() => {
  switch (props.variant) {
    case 'master':
      return 'bg-game-blue text-board-white'
    case 'spectator':
      return 'bg-board-cream text-ink-black'
    case 'system':
      return 'bg-game-pink text-ink-black'
    case 'danger':
      return 'bg-game-red text-board-white'
    case 'success':
      return 'bg-game-green text-board-white'
    case 'neutral':
    default:
      return 'bg-game-yellow text-ink-black'
  }
})
</script>
