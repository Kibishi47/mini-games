<template>
  <div class="bg-board-white border-[3.5px] border-ink-black rounded-2xl shadow-pop-md flex flex-col h-full overflow-hidden select-none">
    
    <!-- En-tête du Chat Pop Moderniste -->
    <div class="px-5 py-4 border-b-[3px] border-ink-black flex items-center justify-between bg-game-yellow">
      <div class="flex items-center space-x-2.5">
        <MessageSquare class="w-5 h-5 text-ink-black" />
        <span class="font-display font-black text-ink-black uppercase text-base tracking-wider">
          Gazette du Festival
        </span>
      </div>
      <AppBadge variant="master">
        {{ roomStore.chatMessages.length }} msg
      </AppBadge>
    </div>

    <!-- Liste des Messages & Événements Système -->
    <div ref="messagesContainer" @scroll="handleScroll" class="flex-1 p-4 overflow-y-auto space-y-3 bg-board-cream">
      <div
        v-for="msg in roomStore.chatMessages"
        :key="msg.id"
      >
        <!-- Message Système : Pilule bicolore centrée sans émoji -->
        <div v-if="msg.is_system" class="flex items-center justify-center my-2">
          <span class="inline-flex items-center gap-1.5 text-xs font-condensed uppercase px-3 py-1 rounded-full border-2 border-ink-black bg-game-pink text-ink-black shadow-pop-xs">
            <Megaphone class="w-3.5 h-3.5 text-ink-black" />
            <span>{{ msg.content }}</span>
          </span>
        </div>

        <!-- Message Utilisateur Local (aligné à droite) -->
        <div v-else-if="msg.sender_id === roomStore.me?.id" class="flex items-end justify-end space-x-2">
          <div class="max-w-[85%] text-right">
            <div class="flex items-baseline justify-end space-x-2">
              <span class="text-[10px] font-condensed text-ink-black/50">{{ formatTime(msg.created_at) }}</span>
              <span class="font-display font-black text-xs uppercase text-game-yellow">Moi</span>
            </div>
            <div class="mt-1 inline-block bg-game-yellow/20 border-2 border-ink-black rounded-2xl rounded-tr-none px-3.5 py-2 shadow-pop-xs text-sm font-body font-semibold text-ink-black break-words text-left">
              {{ msg.content }}
            </div>
          </div>
          <GameMascot :name="(msg.mascot as any) || 'dice'" mood="idle" size="sm" class="mb-0.5 flex-shrink-0" />
        </div>

        <!-- Message Autre Joueur (aligné à gauche) -->
        <div v-else class="flex items-start justify-start space-x-3">
          <GameMascot :name="(msg.mascot as any) || 'dice'" mood="idle" size="sm" class="mt-1 flex-shrink-0" />

          <div class="max-w-[85%] min-w-0">
            <div class="flex items-baseline space-x-2">
              <span class="font-display font-black text-xs uppercase text-game-blue truncate">{{ msg.sender }}</span>
              <span class="text-[10px] font-condensed text-ink-black/50">{{ formatTime(msg.created_at) }}</span>
            </div>

            <!-- Bulle de dialogue Pop -->
            <div class="mt-1 inline-block bg-board-white border-2 border-ink-black rounded-2xl rounded-tl-none px-3.5 py-2 shadow-pop-xs text-sm font-body font-semibold text-ink-black break-words max-w-full">
              {{ msg.content }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Si le joueur est muté : bandeau d'alerte rouge -->
    <div v-if="isMuted" class="bg-game-red text-board-white border-t-[3px] border-ink-black px-4 py-2 flex items-center justify-center space-x-2 font-display font-black text-xs uppercase">
      <VolumeX class="w-4 h-4" />
      <span>Vous êtes actuellement muet sur ordre du Master !</span>
    </div>

    <!-- Boîte de saisie -->
    <form
      v-else
      @submit.prevent="handleSend"
      class="p-3 bg-board-white border-t-[3px] border-ink-black flex items-center space-x-2"
    >
      <input
        v-model="inputContent"
        type="text"
        placeholder="Écrire un message..."
        maxlength="300"
        class="flex-1 bg-board-cream border-2 border-ink-black rounded-xl px-4 py-2 text-sm font-body font-semibold text-ink-black placeholder:text-ink-black/40 focus:outline-none focus:shadow-pop-xs transition-none"
      />

      <AppButton
        type="submit"
        variant="primary"
        size="sm"
        :disabled="!inputContent.trim()"
      >
        <Send class="w-4 h-4" />
      </AppButton>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed, nextTick } from 'vue'
import { MessageSquare, Send, VolumeX, Megaphone } from 'lucide-vue-next'
import { useRoomStore } from '~/stores/room'
import GameMascot from '~/components/ui/GameMascot.vue'
import AppButton from '~/components/ui/AppButton.vue'
import AppBadge from '~/components/ui/AppBadge.vue'

const props = defineProps<{
  onSend: (content: string) => void
}>()

const roomStore = useRoomStore()
const inputContent = ref('')
const messagesContainer = ref<HTMLElement | null>(null)

const isMuted = computed(() => {
  return roomStore.me?.is_muted ?? false
})

const handleSend = () => {
  if (!inputContent.value.trim() || isMuted.value) return
  props.onSend(inputContent.value.trim())
  inputContent.value = ''
}

const formatTime = (dateStr: string) => {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

const isAtBottom = ref(true)

const handleScroll = () => {
  const el = messagesContainer.value
  if (!el) return
  // Tolérance de 40px pour détecter si on est au fond
  isAtBottom.value = el.scrollHeight - el.scrollTop - el.clientHeight <= 40
}

const scrollToBottom = (smooth = true) => {
  nextTick(() => {
    const el = messagesContainer.value
    if (!el) return
    el.scrollTo({ top: el.scrollHeight, behavior: smooth ? 'smooth' : 'auto' })
  })
}

watch(
  () => roomStore.chatMessages,
  () => {
    if (isAtBottom.value) {
      scrollToBottom()
    }
  },
  { deep: true, flush: 'post' }
)

onMounted(() => {
  scrollToBottom(false)
})
</script>
