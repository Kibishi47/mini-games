import { ref, onMounted, onUnmounted } from 'vue'
import { useProfileStore } from '~/stores/profile'
import { useRoomStore } from '~/stores/room'
import { useGameStore } from '~/stores/game'

export function useWebSocket(roomCode: string) {
  const profileStore = useProfileStore()
  const roomStore = useRoomStore()
  const gameStore = useGameStore()
  const config = useRuntimeConfig()
  const router = useRouter()

  const socket = ref<WebSocket | null>(null)
  const isConnected = ref(false)
  const isInitializing = ref(true)
  const errorMessage = ref<string | null>(null)
  let reconnectTimer: any = null

  const connect = () => {
    if (!roomCode) return

    if (socket.value) {
      socket.value.close()
    }

    // Toujours s'assurer que le profil est hydraté avec son session_token avant de connecter
    profileStore.initProfile()

    const wsBase = config.public.wsUrl || 'ws://localhost:8080/ws'
    const query = new URLSearchParams({
      room: roomCode,
      token: profileStore.sessionToken || '',
      nickname: profileStore.nickname || 'Joueur',
      mascot: profileStore.mascot || 'dice',
      color: profileStore.color || '#FFD300',
    })

    const wsUrl = `${wsBase}?${query.toString()}`
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
        console.error('Erreur parsing WebSocket:', e)
      }
    }

    ws.onerror = (err) => {
      console.error('WebSocket Error:', err)
    }

    ws.onclose = (event) => {
      isConnected.value = false
      
      // Code 4004 : Room Not Found / Fermée pour inactivité
      if (event.code === 4004) {
        isInitializing.value = false
        if (reconnectTimer) clearTimeout(reconnectTimer)
        router.push({
          path: '/',
          query: { error: 'room_closed' },
        })
        return
      }

      // Reconnexion automatique pour la Grace Period si non fermé volontairement
      if (!event.wasClean) {
        reconnectTimer = setTimeout(() => {
          connect()
        }, 1500)
      }
    }

    socket.value = ws
  }

  const handleServerMessage = (type: string, payload: any) => {
    switch (type) {
      case 'room:sync':
        roomStore.setRoom(payload)
        isInitializing.value = false

        // Synchroniser l'état de manche lors d'un F5 / reconnexion
        if (payload?.status === 'in_game') {
          if (payload?.round_state === 'round_ended') {
            gameStore.currentRound = payload.current_round || 1
            gameStore.targetWord = payload.revealed_word || ''
            const remainingSec = payload.next_round_at
              ? Math.max(0, Math.round((new Date(payload.next_round_at).getTime() - Date.now()) / 1000))
              : 8
            gameStore.roundSummary = {
              round: payload.current_round || 1,
              max_rounds: payload.settings?.max_rounds || 3,
              secret_word: payload.revealed_word || '',
              reason: 'Fin de manche',
              winner_name: '',
              round_scores: {},
              total_scores: {},
              countdown_sec: remainingSec,
            }
            gameStore.isGameOver = false
            gameStore.showRoundSummary = true
          } else if (payload?.round_state === 'game_over') {
            gameStore.currentRound = payload.current_round || 3
            gameStore.targetWord = payload.revealed_word || ''
            gameStore.isGameOver = true
            gameStore.showRoundSummary = true
          } else if (payload?.round_state === 'playing') {
            gameStore.showRoundSummary = false
          }
        } else if (payload?.status === 'in_lobby') {
          gameStore.showRoundSummary = false
        }
        break

      case 'room:state_changed':
        if (payload?.status) {
          roomStore.updateStatus(payload.status)
        }
        break

      case 'player:reconnected':
        // Joueur reconnecté
        break

      case 'room:chat_history':
        roomStore.setChatHistory(payload)
        break

      case 'room:chat_message':
        roomStore.addChatMessage(payload)
        break

      case 'game:round_start':
        gameStore.startRound(payload)
        break

      case 'game:guess_result':
        gameStore.setGuessResult(payload)
        break

      case 'game:opponent_progress':
        gameStore.updateOpponentProgress(payload)
        break

      case 'game:round_end':
        gameStore.endRound(payload)
        break

      case 'game:state_sync':
        gameStore.syncGameState(payload)
        break

      case 'player:location_changed':
        if (payload?.user_id && payload?.location) {
          roomStore.updatePlayerLocation(payload.user_id, payload.location)
        }
        break

      case 'room:scores_reset':
        roomStore.resetPlayerScores()
        break

      case 'error':
        errorMessage.value = payload?.error || 'Une erreur est survenue'
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

  // Méthodes d'action rapides
  const sendChatMessage = (content: string) => {
    send('room:chat', { content })
  }

  const updateSettings = (settings: any) => {
    send('room:update_settings', settings)
  }

  const startGame = () => {
    send('game:start', {})
  }

  const submitGuess = (guess: string) => {
    send('game:guess', { guess })
  }

  const kickPlayer = (targetId: string) => {
    send('room:kick', { target_id: targetId })
  }

  const banPlayer = (targetId: string) => {
    send('room:ban', { target_id: targetId })
  }

  const mutePlayer = (targetId: string, mute: boolean) => {
    send('room:mute', { target_id: targetId, mute })
  }

  const rematch = () => {
    send('room:rematch', {})
  }

  const stopGame = () => {
    send('game:stop', {})
  }

  const returnToLobby = () => {
    send('room:return_lobby', {})
  }

  const returnLobby = () => {
    send('player:return_lobby', {})
  }

  const resetScores = () => {
    send('room:reset_scores', {})
  }

  const leaveRoom = () => {
    send('room:leave', {})
    if (socket.value) {
      socket.value.close(1000, 'Voluntary leave')
    }
  }

  const nextRound = () => {
    send('game:next_round', {})
  }

  onMounted(() => {
    connect()
  })

  onUnmounted(() => {
    if (reconnectTimer) clearTimeout(reconnectTimer)
    if (socket.value) {
      socket.value.close(1000, 'Navigation normale')
    }
  })

  return {
    socket,
    isConnected,
    isInitializing,
    errorMessage,
    send,
    sendChatMessage,
    updateSettings,
    startGame,
    stopGame,
    returnToLobby,
    returnLobby,
    resetScores,
    leaveRoom,
    nextRound,
    submitGuess,
    kickPlayer,
    banPlayer,
    mutePlayer,
    rematch,
  }
}
