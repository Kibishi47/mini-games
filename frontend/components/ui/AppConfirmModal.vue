<template>
  <div
    v-if="isOpen"
    class="fixed inset-0 bg-ink-black/80 backdrop-blur-none flex items-center justify-center z-50 p-4"
  >
    <div
      class="bg-board-white max-w-md w-full p-6 rounded-3xl border-[4px] border-ink-black shadow-pop-lg relative flex flex-col space-y-5 animate-in fade-in zoom-in-95 duration-100"
    >
      <!-- En-tête avec icône & Titre -->
      <div class="flex items-center space-x-3 border-b-2 border-ink-black pb-3">
        <div
          :class="[
            'w-10 h-10 rounded-xl border-2 border-ink-black flex items-center justify-center shadow-pop-xs flex-shrink-0',
            options.variant === 'danger' ? 'bg-game-red text-board-white' : 'bg-game-yellow text-ink-black'
          ]"
        >
          <AlertTriangle class="w-5 h-5" />
        </div>
        <h3 class="font-display font-black text-xl uppercase tracking-tight text-ink-black truncate">
          {{ options.title }}
        </h3>
      </div>

      <!-- Corps du message -->
      <p class="font-body text-sm font-semibold text-ink-black/80 leading-relaxed">
        {{ options.message }}
      </p>

      <!-- Actions -->
      <div class="flex items-center justify-end space-x-3 pt-2">
        <AppButton
          variant="neutral"
          size="md"
          @click="handleCancel"
        >
          <span>{{ options.cancelLabel || 'ANNULER' }}</span>
        </AppButton>

        <AppButton
          :variant="options.variant === 'danger' ? 'danger' : 'primary'"
          size="md"
          @click="handleConfirm"
        >
          <span>{{ options.confirmLabel || 'CONFIRMER' }}</span>
        </AppButton>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { AlertTriangle } from 'lucide-vue-next'
import AppButton from '~/components/ui/AppButton.vue'
import { useConfirmState } from '~/composables/useConfirm'

const { isOpen, options, handleConfirm, handleCancel } = useConfirmState()
</script>
