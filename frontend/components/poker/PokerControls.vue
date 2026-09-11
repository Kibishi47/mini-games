<template>
  <div class="border-[3px] border-ink-black bg-board-white rounded-2xl p-4 shadow-pop-md flex flex-col gap-3">
    <div v-if="!isMyTurn" class="text-center font-display font-black text-xs uppercase text-ink-black/50 py-3">
      En attente de votre tour...
    </div>

    <template v-else>
      <div class="flex items-center justify-between text-[11px] font-condensed font-black uppercase text-ink-black/70">
        <span>{{ callAmount > 0 ? `À suivre : ${callAmount} jetons` : 'Aucune mise à suivre' }}</span>
        <span :class="remainingSeconds <= 5 ? 'text-game-red animate-pulse' : ''">⏱ {{ remainingSeconds }}s</span>
      </div>

      <div class="flex flex-wrap gap-2">
        <AppButton variant="danger" size="md" @click="emitAction('fold')">
          Se Coucher
        </AppButton>
        <AppButton v-if="callAmount <= 0" variant="neutral" size="md" @click="emitAction('check')">
          Checker
        </AppButton>
        <AppButton v-else variant="secondary" size="md" :disabled="callAmount >= myChips" @click="emitAction('call')">
          Suivre ({{ Math.min(callAmount, myChips) }})
        </AppButton>
        <AppButton v-if="myChips > callAmount" variant="success" size="md" @click="toggleRaiseSlider">
          {{ currentBet > 0 ? 'Relancer' : 'Miser' }}
        </AppButton>
        <AppButton variant="primary" size="md" @click="emitAction('all_in')">
          Tapis ({{ myChips + myCommitted }})
        </AppButton>
      </div>

      <div v-if="showRaiseSlider" class="flex flex-wrap items-center gap-3 pt-3 border-t-2 border-ink-black">
        <input
          type="range"
          :min="minRaiseTarget"
          :max="maxRaiseTarget"
          :step="Math.max(bigBlind, 1)"
          v-model.number="raiseAmount"
          class="flex-1 min-w-[120px]"
        />
        <AppInput v-model.number="raiseAmount" type="number" class="w-24 text-center" />
        <AppButton variant="success" size="sm" @click="confirmRaise">
          {{ currentBet > 0 ? 'Relancer à' : 'Miser' }} {{ raiseAmount }}
        </AppButton>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, onMounted, onUnmounted } from 'vue'
import AppButton from '~/components/ui/AppButton.vue'
import AppInput from '~/components/ui/AppInput.vue'

const props = defineProps<{
  isMyTurn: boolean
  callAmount: number
  currentBet: number
  minRaise: number
  bigBlind: number
  myChips: number
  myCommitted: number
  actionEndsAt?: string
}>()

const emit = defineEmits<{
  (e: 'action', payload: { action: string; amount: number }): void
}>()

const showRaiseSlider = ref(false)
const raiseAmount = ref(0)

const minRaiseTarget = computed(() => {
  const base = props.currentBet > 0 ? props.currentBet + Math.max(props.minRaise, props.bigBlind) : props.bigBlind
  return Math.min(base, maxRaiseTarget.value)
})

const maxRaiseTarget = computed(() => props.myChips + props.myCommitted)

const toggleRaiseSlider = () => {
  showRaiseSlider.value = !showRaiseSlider.value
  if (showRaiseSlider.value) {
    raiseAmount.value = minRaiseTarget.value
  }
}

const emitAction = (action: string, amount = 0) => {
  showRaiseSlider.value = false
  emit('action', { action, amount })
}

const confirmRaise = () => {
  const action = props.currentBet > 0 ? 'raise' : 'bet'
  emitAction(action, raiseAmount.value)
}

watch(() => props.isMyTurn, (val) => {
  if (!val) showRaiseSlider.value = false
})

// Décompte visuel du temps restant pour agir, basé sur l'horodatage absolu du serveur
const remainingSeconds = ref(20)
let interval: any = null

const updateCountdown = () => {
  if (!props.actionEndsAt) return
  const diff = Math.max(0, Math.round((new Date(props.actionEndsAt).getTime() - Date.now()) / 1000))
  remainingSeconds.value = diff
}

onMounted(() => {
  updateCountdown()
  interval = setInterval(updateCountdown, 1000)
})

onUnmounted(() => {
  if (interval) clearInterval(interval)
})

watch(() => props.actionEndsAt, updateCountdown)
</script>
