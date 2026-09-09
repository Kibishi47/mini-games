import { ref, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from '~/stores/auth'
import { useRoomStore } from '~/stores/room'
import { useGameStore } from '~/stores/game'

export function useWebSocket(roomCode: string) {
  const authStore = useAuthStore()
  const roomStore = useRoomStore()
  const gameStore = useGameStore()
  const config = useRuntimeConfig()
  const router = useRouter()

  const socket = ref<WebSocket | null>(null)
  const isConnected = ref(false)
  const errorMessage = ref<string | null>(null)
  let reconnectTimer: NodeJS.Timeout | null = null

  const connect = () => {
    if (!authStore.token || !roomCode) return

    // Fermer une socket existante si nécessaire
    if (socket.value) {
      socket.value.close()
    }

    const wsUrl = `${config.public.wsUrl}?token=${encodeURIComponent(authStore.token)}&room=${encodeURIComponent(roomCode)}`
    const ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      isConnected.value = true
      errorMessage.value = null
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        handleServerMessage(msg.type, msg.payload)
      } catch (e) {
        console.error('Erreur parsing message WS', e)
      }
    }

    ws.onerror = (err) => {
      console.error('WebSocket Error:', err)
    }

    ws.onclose = (event) => {
      isConnected.value = false
      // Reconnexion automatique avec le token pour la Grace Period de 45 secondes
      if (!event.wasClean) {
        reconnectTimer = setTimeout(() => {
          connect()
        }, 2500)
      }
    }

    socket.value = ws
  }

  const handleServerMessage = (type: string, payload: any) => {
    switch (type) {
      case 'room:sync':
        roomStore.setRoom(payload.room)
        if (payload.chat) {
          roomStore.setChat(payload.chat)
        }
        // Redirection automatique si le statut a changé vers in_game
        if (payload.room.status === 'in_game' && router.currentRoute.value.path.includes('/lobby/')) {
          router.push(`/game/${roomCode}`)
        } else if (payload.room.status === 'in_lobby' && router.currentRoute.value.path.includes('/game/')) {
          router.push(`/lobby/${roomCode}`)
        }
        break

      case 'chat:message':
        roomStore.addChatMessage(payload)
        break

      case 'game:round_start':
        gameStore.startNewRound(payload)
        if (!router.currentRoute.value.path.includes('/game/')) {
          router.push(`/game/${roomCode}`)
        }
        break

      case 'game:guess_result':
        gameStore.myAttempts = payload.attempts
        gameStore.isSolved = payload.is_solved
        gameStore.isFinished = payload.is_finished
        gameStore.myScore = payload.round_score
        break

      case 'game:opponent_progress':
        gameStore.updateOpponentProgress(payload)
        break

      case 'game:round_end':
        gameStore.setRoundEnd(payload)
        break

      case 'game:sync_state':
        gameStore.wordLength = payload.word_length
        gameStore.maxAttempts = payload.max_attempts
        gameStore.endsAt = new Date(payload.ends_at)
        gameStore.myAttempts = payload.my_attempts || []
        gameStore.opponents = payload.opponents || []
        break

      case 'error':
        errorMessage.value = payload.error
        setTimeout(() => {
          errorMessage.value = null
        }, 4000)
        break
    }
  }

  const send = (type: string, payload: any = {}) => {
    if (socket.value && socket.value.readyState === WebSocket.OPEN) {
      socket.value.send(JSON.stringify({ type, payload }))
    }
  }

  const sendChat = (content: string) => {
    send('chat:send', { content })
  }

  const updateSettings = (settings: any) => {
    send('room:update_settings', settings)
  }

  const performAction = (actionType: 'kick' | 'ban' | 'mute' | 'unmute', targetUserId: string) => {
    send('room:action', { action_type: actionType, target_user_id: targetUserId })
  }

  const startGame = () => {
    send('game:start', {})
  }

  const submitGuess = (guess: string) => {
    send('game:submit_guess', { guess })
  }

  const requestRematch = () => {
    send('game:rematch', {})
  }

  const leaveRoom = () => {
    send('room:leave', {})
    if (socket.value) {
      socket.value.close(1000, 'User left')
    }
    router.push('/')
  }

  onMounted(() => {
    connect()
  })

  onUnmounted(() => {
    if (reconnectTimer) clearTimeout(reconnectTimer)
    if (socket.value) {
      socket.value.close()
    }
  })

  return {
    isConnected,
    errorMessage,
    send,
    sendChat,
    updateSettings,
    performAction,
    startGame,
    submitGuess,
    requestRematch,
    leaveRoom,
  }
}
