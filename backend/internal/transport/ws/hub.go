package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"minigames-backend/internal/domain"
	"minigames-backend/internal/repository/postgres"
	"minigames-backend/internal/repository/redis"
)

type Hub struct {
	roomsMu      sync.RWMutex
	roomClients  map[string]map[uuid.UUID]*Client // roomCode -> (userID -> *Client)
	roomRepo     *redis.RoomRepository
	sessionRepo  *postgres.RoomSessionRepository
	gameHandler  GameEventHandler
}

type GameEventHandler interface {
	HandleGameAction(client *Client, action string, payload json.RawMessage)
	OnPlayerJoined(client *Client, room *domain.Room)
	OnPlayerLeft(roomCode string, userID uuid.UUID)
}

func NewHub(roomRepo *redis.RoomRepository, sessionRepo *postgres.RoomSessionRepository) *Hub {
	return &Hub{
		roomClients: make(map[string]map[uuid.UUID]*Client),
		roomRepo:    roomRepo,
		sessionRepo: sessionRepo,
	}
}

func (h *Hub) SetGameHandler(handler GameEventHandler) {
	h.gameHandler = handler
}

func (h *Hub) Register(client *Client) {
	h.roomsMu.Lock()
	if _, ok := h.roomClients[client.roomCode]; !ok {
		h.roomClients[client.roomCode] = make(map[uuid.UUID]*Client)
	}
	h.roomClients[client.roomCode][client.userID] = client
	h.roomsMu.Unlock()

	ctx := context.Background()
	_ = h.roomRepo.UpdatePlayerActivity(ctx, client.roomCode, client.userID, true)

	// Synchroniser la room avec le nouvel arrivant
	h.SyncRoom(client.roomCode)

	// Notification de présence
	room, err := h.roomRepo.GetRoom(ctx, client.roomCode)
	if err == nil {
		p, _ := h.roomRepo.GetPlayer(ctx, client.roomCode, client.userID)
		if p != nil {
			h.BroadcastSystemMessage(client.roomCode, fmt.Sprintf("%s a rejoint la salle", p.DisplayUsername))
		}

		if h.gameHandler != nil {
			h.gameHandler.OnPlayerJoined(client, room)
		}
	}
}

func (h *Hub) Unregister(client *Client) {
	h.roomsMu.Lock()
	if clients, ok := h.roomClients[client.roomCode]; ok {
		delete(clients, client.userID)
		if len(clients) == 0 {
			delete(h.roomClients, client.roomCode)
		}
	}
	h.roomsMu.Unlock()

	ctx := context.Background()
	// Marquer comme déconnecté mais conserver le slot (Grace Period de 45 secondes)
	_ = h.roomRepo.UpdatePlayerActivity(ctx, client.roomCode, client.userID, false)

	// Notifier les autres joueurs
	p, err := h.roomRepo.GetPlayer(ctx, client.roomCode, client.userID)
	if err == nil && p != nil {
		h.BroadcastSystemMessage(client.roomCode, fmt.Sprintf("%s s'est déconnecté (en attente...)", p.DisplayUsername))
	}

	h.SyncRoom(client.roomCode)

	if h.gameHandler != nil {
		h.gameHandler.OnPlayerLeft(client.roomCode, client.userID)
	}
}

func (h *Hub) BroadcastToRoom(roomCode string, msg domain.WSMessage) {
	h.roomsMu.RLock()
	clients, ok := h.roomClients[roomCode]
	if !ok {
		h.roomsMu.RUnlock()
		return
	}
	// Copier la liste pour ne pas bloquer le mutex pendant l'envoi
	recipients := make([]*Client, 0, len(clients))
	for _, c := range clients {
		recipients = append(recipients, c)
	}
	h.roomsMu.RUnlock()

	for _, c := range recipients {
		c.Send(msg)
	}
}

func (h *Hub) SendToUser(roomCode string, userID uuid.UUID, msg domain.WSMessage) {
	h.roomsMu.RLock()
	client, ok := h.roomClients[roomCode][userID]
	h.roomsMu.RUnlock()
	if ok && client != nil {
		client.Send(msg)
	}
}

func (h *Hub) SyncRoom(roomCode string) {
	ctx := context.Background()
	room, err := h.roomRepo.GetRoom(ctx, roomCode)
	if err != nil {
		return
	}

	chat, _ := h.roomRepo.GetRecentChat(ctx, roomCode)

	payload, _ := json.Marshal(map[string]interface{}{
		"room": room,
		"chat": chat,
	})

	h.BroadcastToRoom(roomCode, domain.WSMessage{
		Type:    "room:sync",
		Payload: payload,
	})
}

func (h *Hub) BroadcastSystemMessage(roomCode, text string) {
	ctx := context.Background()
	msg := &domain.ChatMessage{
		ID:        uuid.New().String(),
		SenderID:  uuid.Nil,
		Sender:    "Système",
		Content:   text,
		IsSystem:  true,
		CreatedAt: time.Now(),
	}
	_ = h.roomRepo.AddChatMessage(ctx, roomCode, msg)

	payload, _ := json.Marshal(msg)
	h.BroadcastToRoom(roomCode, domain.WSMessage{
		Type:    "chat:message",
		Payload: payload,
	})
}

// Dispatcher central des messages WebSocket reçus des clients
func (h *Hub) Dispatch(client *Client, msg domain.WSMessage) {
	ctx := context.Background()

	switch msg.Type {
	case "chat:send":
		h.handleChatSend(ctx, client, msg.Payload)
	case "room:update_settings":
		h.handleUpdateSettings(ctx, client, msg.Payload)
	case "room:action":
		h.handleModerationAction(ctx, client, msg.Payload)
	case "room:leave":
		h.handlePlayerLeave(ctx, client)
	default:
		// Déléguer aux handlers de jeu (ex: game:start, game:submit_guess)
		if h.gameHandler != nil {
			h.gameHandler.HandleGameAction(client, msg.Type, msg.Payload)
		}
	}
}

func (h *Hub) handleChatSend(ctx context.Context, client *Client, payload json.RawMessage) {
	var body struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(payload, &body); err != nil || len(body.Content) == 0 {
		return
	}

	// Rate limiting chat : max 5 messages par 3 secondes
	allowed, _ := h.roomRepo.CheckRateLimit(ctx, fmt.Sprintf("chat:%s:%s", client.roomCode, client.userID), 5, 3*time.Second)
	if !allowed {
		client.SendError("Vous envoyez des messages trop rapidement")
		return
	}

	player, err := h.roomRepo.GetPlayer(ctx, client.roomCode, client.userID)
	if err != nil || player == nil {
		return
	}

	// Vérification du statut Mute
	if player.IsMuted {
		client.SendError("Vous êtes actuellement muet dans cette salle")
		return
	}

	chatMsg := &domain.ChatMessage{
		ID:        uuid.New().String(),
		SenderID:  client.userID,
		Sender:    player.DisplayUsername,
		AvatarURL: player.AvatarURL,
		Content:   body.Content,
		IsSystem:  false,
		CreatedAt: time.Now(),
	}

	_ = h.roomRepo.AddChatMessage(ctx, client.roomCode, chatMsg)

	data, _ := json.Marshal(chatMsg)
	h.BroadcastToRoom(client.roomCode, domain.WSMessage{
		Type:    "chat:message",
		Payload: data,
	})
}

func (h *Hub) handleUpdateSettings(ctx context.Context, client *Client, payload json.RawMessage) {
	room, err := h.roomRepo.GetRoom(ctx, client.roomCode)
	if err != nil {
		return
	}
	if room.MasterID != client.userID {
		client.SendError("Seul le Master peut modifier les paramètres de la salle")
		return
	}
	if room.Status != domain.RoomStatusInLobby {
		client.SendError("Impossible de modifier les règles pendant une partie")
		return
	}

	var newSettings domain.RoomSettings
	if err := json.Unmarshal(payload, &newSettings); err != nil {
		client.SendError("Paramètres invalides")
		return
	}

	// Bornes de sécurité
	if newSettings.WordLength < 4 || newSettings.WordLength > 8 {
		newSettings.WordLength = 5
	}
	if newSettings.RoundDuration < 30 || newSettings.RoundDuration > 300 {
		newSettings.RoundDuration = 60
	}
	if newSettings.MaxRounds < 1 || newSettings.MaxRounds > 10 {
		newSettings.MaxRounds = 3
	}

	_ = h.roomRepo.UpdateRoomSettings(ctx, client.roomCode, newSettings)
	h.BroadcastSystemMessage(client.roomCode, "Les paramètres de la partie ont été mis à jour par le Master")
	h.SyncRoom(client.roomCode)
}

func (h *Hub) handleModerationAction(ctx context.Context, client *Client, payload json.RawMessage) {
	room, err := h.roomRepo.GetRoom(ctx, client.roomCode)
	if err != nil || room.MasterID != client.userID {
		client.SendError("Privilège insuffisant : vous n'êtes pas le Master")
		return
	}

	var action struct {
		TargetID uuid.UUID `json:"target_user_id"`
		Type     string    `json:"action_type"` // "kick", "ban", "mute", "unmute"
	}
	if err := json.Unmarshal(payload, &action); err != nil {
		return
	}

	// INTERDICTION ABSOLUE de cibler soi-même
	if action.TargetID == client.userID {
		client.SendError("Action impossible sur votre propre profil")
		return
	}

	targetPlayer, err := h.roomRepo.GetPlayer(ctx, client.roomCode, action.TargetID)
	if err != nil || targetPlayer == nil {
		client.SendError("Joueur cible introuvable")
		return
	}

	switch action.Type {
	case "mute":
		_ = h.roomRepo.UpdatePlayerMute(ctx, client.roomCode, action.TargetID, true)
		h.BroadcastSystemMessage(client.roomCode, fmt.Sprintf("%s a été réduit au silence par le Master", targetPlayer.DisplayUsername))
		h.SyncRoom(client.roomCode)

	case "unmute":
		_ = h.roomRepo.UpdatePlayerMute(ctx, client.roomCode, action.TargetID, false)
		h.BroadcastSystemMessage(client.roomCode, fmt.Sprintf("%s peut de nouveau parler", targetPlayer.DisplayUsername))
		h.SyncRoom(client.roomCode)

	case "kick":
		h.roomsMu.RLock()
		targetClient := h.roomClients[client.roomCode][action.TargetID]
		h.roomsMu.RUnlock()

		_ = h.roomRepo.RemovePlayer(ctx, client.roomCode, action.TargetID)
		h.BroadcastSystemMessage(client.roomCode, fmt.Sprintf("%s a été expulsé par le Master", targetPlayer.DisplayUsername))
		h.SyncRoom(client.roomCode)

		if targetClient != nil {
			targetClient.SendError("Vous avez été expulsé de la salle")
			targetClient.Close()
		}

	case "ban":
		_ = h.roomRepo.BanUser(ctx, client.roomCode, action.TargetID, 30*time.Minute)
		h.roomsMu.RLock()
		targetClient := h.roomClients[client.roomCode][action.TargetID]
		h.roomsMu.RUnlock()

		_ = h.roomRepo.RemovePlayer(ctx, client.roomCode, action.TargetID)
		h.BroadcastSystemMessage(client.roomCode, fmt.Sprintf("%s a été banni de la salle", targetPlayer.DisplayUsername))
		h.SyncRoom(client.roomCode)

		if targetClient != nil {
			targetClient.SendError("Vous avez été banni de cette salle pour 30 minutes")
			targetClient.Close()
		}
	}
}

func (h *Hub) handlePlayerLeave(ctx context.Context, client *Client) {
	player, _ := h.roomRepo.GetPlayer(ctx, client.roomCode, client.userID)
	isMaster := false
	if player != nil && player.Role == domain.RoleMaster {
		isMaster = true
	}

	_ = h.roomRepo.RemovePlayer(ctx, client.roomCode, client.userID)

	if player != nil {
		h.BroadcastSystemMessage(client.roomCode, fmt.Sprintf("%s a quitté la salle", player.DisplayUsername))
	}

	// Si le Master est parti, passation automatique au joueur le plus ancien
	if isMaster {
		h.PromoteNextMaster(ctx, client.roomCode)
	}

	h.SyncRoom(client.roomCode)
	client.Close()
}

// PromoteNextMaster sélectionne le plus ancien joueur actif pour devenir Master
func (h *Hub) PromoteNextMaster(ctx context.Context, roomCode string) {
	players, err := h.roomRepo.GetPlayers(ctx, roomCode)
	if err != nil || len(players) == 0 {
		// La salle est vide, elle sera fermée par le worker ou basculée
		_ = h.roomRepo.UpdateRoomStatus(ctx, roomCode, domain.RoomStatusClosed)
		_ = h.sessionRepo.UpdateStatus(ctx, roomCode, domain.RoomStatusClosed)
		return
	}

	// Le premier dans la liste triée est le plus ancien
	newMaster := players[0]
	_ = h.roomRepo.SetMaster(ctx, roomCode, newMaster.UserID)
	h.BroadcastSystemMessage(roomCode, fmt.Sprintf("👑 %s est maintenant le nouveau Master de la salle", newMaster.DisplayUsername))
}
