<template>
  <button
    :type="type"
    :disabled="disabled"
    :class="[
      'inline-flex items-center justify-center font-display font-black uppercase tracking-wider transition-none select-none',
      'border-[3px] border-ink-black rounded-xl active:translate-x-[2px] active:translate-y-[2px] active:shadow-none',
      'disabled:opacity-40 disabled:cursor-not-allowed disabled:active:translate-x-0 disabled:active:translate-y-0 disabled:active:shadow-pop-sm',
      variantClasses,
      sizeClasses
    ]"
    @click="$emit('click', $event)"
  >
    <slot />
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'

export type ButtonVariant = 'primary' | 'secondary' | 'danger' | 'success' | 'neutral' | 'pink'
export type ButtonSize = 'sm' | 'md' | 'lg'

const props = withDefaults(
  defineProps<{
    type?: 'button' | 'submit' | 'reset'
    variant?: ButtonVariant
    size?: ButtonSize
    disabled?: boolean
  }>(),
  {
    type: 'button',
    variant: 'primary',
    size: 'md',
    disabled: false,
  }
)

defineEmits<{
  (e: 'click', event: MouseEvent): void
}>()

const variantClasses = computed(() => {
  switch (props.variant) {
    case 'primary':
      return 'bg-game-yellow text-ink-black shadow-pop-md hover:bg-[#ffe033]'
    case 'secondary':
      return 'bg-game-blue text-board-white shadow-pop-md hover:bg-[#2563eb]'
    case 'danger':
      return 'bg-game-red text-board-white shadow-pop-md hover:bg-[#ef4444]'
    case 'success':
      return 'bg-game-green text-board-white shadow-pop-md hover:bg-[#16a34a]'
    case 'pink':
      return 'bg-game-pink text-ink-black shadow-pop-md hover:bg-[#f48eb0]'
    case 'neutral':
    default:
      return 'bg-board-white text-ink-black shadow-pop-sm hover:bg-board-cream'
  }
})

const sizeClasses = computed(() => {
  switch (props.size) {
    case 'sm':
      return 'px-3 py-1.5 text-xs'
    case 'lg':
      return 'px-6 py-4 text-lg shadow-pop-lg active:translate-x-[3px] active:translate-y-[3px]'
    case 'md':
    default:
      return 'px-4 py-2.5 text-sm'
  }
})
</script>
