<template>
  <div
    :class="[
      'border-[3px] border-ink-black rounded-2xl p-5 select-none relative',
      variantClasses,
      shadowClass
    ]"
  >
    <slot />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    variant?: 'white' | 'cream' | 'yellow' | 'blue' | 'pink'
    shadow?: 'sm' | 'md' | 'lg' | 'none'
  }>(),
  {
    variant: 'white',
    shadow: 'md',
  }
)

const variantClasses = computed(() => {
  switch (props.variant) {
    case 'cream':
      return 'bg-board-cream text-ink-black'
    case 'yellow':
      return 'bg-game-yellow text-ink-black'
    case 'blue':
      return 'bg-game-blue text-board-white'
    case 'pink':
      return 'bg-game-pink text-ink-black'
    case 'white':
    default:
      return 'bg-board-white text-ink-black'
  }
})

const shadowClass = computed(() => {
  switch (props.shadow) {
    case 'sm':
      return 'shadow-pop-sm'
    case 'lg':
      return 'shadow-pop-lg'
    case 'none':
      return ''
    case 'md':
    default:
      return 'shadow-pop-md'
  }
})
</script>
