<template>
  <div class="glass-panel flex flex-col h-full rounded-2xl overflow-hidden border border-white/10 shadow-2xl">
    <!-- Header -->
    <div class="px-5 py-4 border-b border-white/10 flex items-center justify-between bg-brand-surface/40">
      <div class="flex items-center space-x-2">
        <MessageSquare class="w-5 h-5 text-indigo-400" />
        <span class="font-semibold text-white tracking-wide text-sm">Salon de discussion</span>
      </div>
      <span class="text-xs px-2.5 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 font-medium">
        {{ roomStore.chatMessages.length }} msg
      </span>
    </div>

    <!-- Message List -->
    <div ref="messagesContainer" class="flex-1 p-4 overflow-y-auto space-y-3">
      <div
        v-for="msg in roomStore.chatMessages"
        :key="msg.id"
        class="transition-all duration-200"
      >
        <!-- Message Système -->
        <div v-if="msg.is_system" class="flex items-center justify-center my-2">
          <span class="text-xs px-3 py-1 rounded-full bg-white/5 border border-white/5 text-slate-400 italic">
            {{ msg.content }}
          </span>
        </div>

        <!-- Message Utilisateur -->
        <div v-else class="flex items-start space-x-3 group">
          <img
            :src="msg.avatar_url || 'https://api.dicebear.com/7.x/bottts/svg?seed=' + msg.sender"
            alt="Avatar"
            class="w-8 h-8 rounded-lg bg-indigo-950/60 border border-white/10 p-0.5 mt-0.5 flex-shrink-0"
          />
          <div class="flex-1 min-w-0">
            <div class="flex items-baseline space-x-2">
              <span class="font-medium text-xs text-indigo-300 truncate">{{ msg.sender }}</span>
              <span class="text-[10px] text-slate-500">{{ formatTime(msg.created_at) }}</span>
            </div>
            <p class="text-sm text-slate-200 break-words mt-0.5 leading-relaxed">{{ msg.content }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Input Box -->
    <form @submit.prevent="handleSend" class="p-3 bg-brand-card/60 border-t border-white/10 flex items-center space-x-2">
      <input
        v-model="inputContent"
        type="text"
        placeholder="Envoyer un message au lobby..."
        maxlength="200"
        class="flex-1 bg-brand-surface/70 border border-white/10 rounded-xl px-4 py-2.5 text-sm text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all"
      />
      <button
        type="submit"
        :disabled="!inputContent.trim()"
        class="p-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 disabled:hover:bg-indigo-600 text-white shadow-lg shadow-indigo-600/30 transition-all duration-150 flex items-center justify-center"
      >
        <Send class="w-4 h-4" />
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { MessageSquare, Send } from 'lucide-vue-next'
import { useRoomStore } from '~/stores/room'

const props = defineProps<{
  onSend: (content: string) => void
}>()

const roomStore = useRoomStore()
const inputContent = ref('')
const messagesContainer = ref<HTMLElement | null>(null)

const handleSend = () => {
  if (!inputContent.value.trim()) return
  props.onSend(inputContent.value.trim())
  inputContent.value = ''
}

const formatTime = (dateStr: string) => {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

watch(
  () => roomStore.chatMessages.length,
  async () => {
    await nextTick()
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  }
)
</script>
