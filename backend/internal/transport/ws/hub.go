package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"minigames-backend/internal/domain"
	"minigames-backend/internal/repository/redis"
)

type Hub struct {
	roomsMu     sync.RWMutex
	roomClients map[string]map[uuid.UUID]*Client // roomCode -> (userID -> *Client)
	roomRepo    *redis.RoomRepository
	gameHandler GameEventHandler

	// Période de grâce de déconnexion (45s avant suppression/transfert Master)
	timersMu         sync.Mutex
	disconnectTimers map[string]*time.Timer // key: roomCode + ":" + userID.String()

	// Rate limiting en mémoire par client pour anti-spam
	msgRateMu   sync.Mutex
	msgRates    map[uuid.UUID][]time.Time
}

type GameEventHandler interface {
	HandleGameAction(client *Client, action string, payload json.RawMessage)
	OnPlayerJoined(client *Client, room *domain.Room)
	OnPlayerLeft(roomCode string, userID uuid.UUID)
}

func NewHub(roomRepo *redis.RoomRepository) *Hub {
	return &Hub{
		roomClients:      make(map[string]map[uuid.UUID]*Client),
		roomRepo:         roomRepo,
		disconnectTimers: make(map[string]*time.Timer),
		msgRates:         make(map[uuid.UUID][]time.Time),
	}
}

func (h *Hub) SetGameHandler(handler GameEventHandler) {
	h.gameHandler = handler
}

func (h *Hub) Register(client *Client) {
	timerKey := fmt.Sprintf("%s:%s", client.roomCode, client.userID)
	h.timersMu.Lock()
	if t, exists := h.disconnectTimers[timerKey]; exists {
		t.Stop()
		delete(h.disconnectTimers, timerKey)
	}
	h.timersMu.Unlock()

	h.roomsMu.Lock()
	if _, ok := h.roomClients[client.roomCode]; !ok {
		h.roomClients[client.roomCode] = make(map[uuid.UUID]*Client)
	}
	h.roomClients[client.roomCode][client.userID] = client
	h.roomsMu.Unlock()

	ctx := context.Background()

	// Vérifier si le joueur était déjà dans la salle (reconnexion après F5 / reload)
	existingPlayer, _ := h.roomRepo.GetPlayer(ctx, client.roomCode, client.userID)
	isReconnect := existingPlayer != nil

	_ = h.roomRepo.UpdatePlayerActivity(ctx, client.roomCode, client.userID, true)
	if client.sessionToken != "" {
		_ = h.roomRepo.RefreshSession(ctx, client.sessionToken)
	}

	// Synchroniser la room avec le nouvel arrivant et les autres
	h.SyncRoom(client.roomCode)

	room, err := h.roomRepo.GetRoom(ctx, client.roomCode)
	if err == nil {
		p, _ := h.roomRepo.GetPlayer(ctx, client.roomCode, client.userID)
		nickname := "Un joueur"
		if p != nil {
			nickname = p.Nickname
		}

		if isReconnect {
			// Annoncer la reconnexion sans dupliquer de message d'arrivée
			reconnectPayload, _ := json.Marshal(map[string]interface{}{
				"user_id":  client.userID,
				"nickname": nickname,
			})
			h.BroadcastToRoom(client.roomCode, domain.WSMessage{
				Type:    "player:reconnected",
				Payload: reconnectPayload,
			})
			h.BroadcastSystemMessage(client.roomCode, fmt.Sprintf("🔄 %s s'est reconnecté", nickname))
		} else {
			h.BroadcastSystemMessage(client.roomCode, fmt.Sprintf("👋 %s a rejoint la salle", nickname))
		}

		// Envoi de l'historique du chat au nouveau client
		chatHistory, _ := h.roomRepo.GetRecentChatMessages(ctx, client.roomCode)
		chatBytes, _ := json.Marshal(chatHistory)
		client.Send(domain.WSMessage{
			Type:    "room:chat_history",
			Payload: chatBytes,
		})

		// Notification au moteur de jeu
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
	_ = h.roomRepo.UpdatePlayerActivity(ctx, client.roomCode, client.userID, false)

	// Lancer un compte à rebours de grâce de 45 secondes avant de retirer le joueur ou transférer le Master
	timerKey := fmt.Sprintf("%s:%s", client.roomCode, client.userID)
	h.timersMu.Lock()
	if t, exists := h.disconnectTimers[timerKey]; exists {
		t.Stop()
	}
	h.disconnectTimers[timerKey] = time.AfterFunc(45*time.Second, func() {
		h.timersMu.Lock()
		delete(h.disconnectTimers, timerKey)
		h.timersMu.Unlock()

		bgCtx := context.Background()
		p, err := h.roomRepo.GetPlayer(bgCtx, client.roomCode, client.userID)
		if err == nil && !p.IsConnected {
			// Le joueur ne s'est pas reconnecté dans le délai de grâce
			_ = h.roomRepo.RemovePlayer(bgCtx, client.roomCode, client.userID)
			h.BroadcastSystemMessage(client.roomCode, fmt.Sprintf("🚶 %s a quitté la salle (délai de grâce expiré)", p.Nickname))
			h.HandleMasterSuccession(client.roomCode, client.userID)

			remaining, _ := h.roomRepo.GetPlayers(bgCtx, client.roomCode)
			if len(remaining) == 0 {
				_ = h.roomRepo.CloseRoom(bgCtx, client.roomCode)
			} else {
				h.SyncRoom(client.roomCode)
			}
		}
	})
	h.timersMu.Unlock()

	// Synchroniser l'état (IsConnected: false visible sur les avatars sans éjecter)
	h.SyncRoom(client.roomCode)

	if h.gameHandler != nil {
		h.gameHandler.OnPlayerLeft(client.roomCode, client.userID)
	}
}

func (h *Hub) TouchPlayer(client *Client) {
	ctx := context.Background()
	_ = h.roomRepo.UpdatePlayerActivity(ctx, client.roomCode, client.userID, true)
	if client.sessionToken != "" {
		_ = h.roomRepo.RefreshSession(ctx, client.sessionToken)
	}
}

func (h *Hub) HandleMasterSuccession(roomCode string, departedUserID uuid.UUID) {
	ctx := context.Background()
	room, err := h.roomRepo.GetRoom(ctx, roomCode)
	if err != nil || room.MasterID != departedUserID {
		return
	}

	// Trouver le joueur connecté le plus ancien
	players, err := h.roomRepo.GetPlayers(ctx, roomCode)
	if err != nil || len(players) == 0 {
		return
	}

	var newMaster *domain.RoomPlayer
	for _, p := range players {
		if p.ID != departedUserID && p.IsConnected {
			newMaster = &p
			break
		}
	}

	// Si aucun connecté, prendre le premier non-parti
	if newMaster == nil {
		for _, p := range players {
			if p.ID != departedUserID {
				newMaster = &p
				break
			}
		}
	}

	if newMaster != nil {
		_ = h.roomRepo.UpdateRoomMaster(ctx, roomCode, newMaster.ID)
		h.BroadcastSystemMessage(roomCode, fmt.Sprintf("👑 %s est maintenant le Master de la salle !", newMaster.Nickname))
		h.SyncRoom(roomCode)
	}
}

func (h *Hub) HandleMessage(client *Client, msg domain.WSMessage) {
	switch {
	case strings.HasPrefix(msg.Type, "game:"):
		if h.gameHandler != nil {
			h.gameHandler.HandleGameAction(client, msg.Type, msg.Payload)
		}
	case msg.Type == "room:chat":
		h.handleChat(client, msg.Payload)
	case msg.Type == "room:update_settings":
		h.handleUpdateSettings(client, msg.Payload)
	case msg.Type == "room:kick":
		h.handleKick(client, msg.Payload)
	case msg.Type == "room:ban":
		h.handleBan(client, msg.Payload)
	case msg.Type == "room:mute":
		h.handleMute(client, msg.Payload)
	case msg.Type == "room:rematch", msg.Type == "room:return_lobby", msg.Type == "game:stop":
		if h.gameHandler != nil {
			h.gameHandler.HandleGameAction(client, msg.Type, msg.Payload)
		} else {
			h.handleRematch(client)
		}
	case msg.Type == "room:sync":
		h.SyncRoom(client.roomCode)
	}
}

func (h *Hub) handleChat(client *Client, payload json.RawMessage) {
	ctx := context.Background()
	p, err := h.roomRepo.GetPlayer(ctx, client.roomCode, client.userID)
	if err != nil {
		return
	}

	// Vérifier si le joueur est muet
	if p.IsMuted {
		client.SendError("Vous avez été rendu muet par le Master")
		return
	}

	// Rate-limiting anti-spam (max 5 messages en 3 secondes)
	h.msgRateMu.Lock()
	now := time.Now()
	times := h.msgRates[client.userID]
	recentTimes := make([]time.Time, 0, len(times))
	for _, t := range times {
		if now.Sub(t) < 3*time.Second {
			recentTimes = append(recentTimes, t)
		}
	}
	if len(recentTimes) >= 5 {
		h.msgRateMu.Unlock()
		client.SendError("Vous envoyez des messages trop vite !")
		return
	}
	recentTimes = append(recentTimes, now)
	h.msgRates[client.userID] = recentTimes
	h.msgRateMu.Unlock()

	var body struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return
	}

	content := strings.TrimSpace(body.Content)
	if content == "" || len(content) > 300 {
		return
	}

	chatMsg := &domain.ChatMessage{
		ID:        uuid.New().String(),
		SenderID:  p.ID,
		Sender:    p.Nickname,
		Mascot:    p.Mascot,
		Color:     p.Color,
		Content:   content,
		IsSystem:  false,
		CreatedAt: time.Now(),
	}

	_ = h.roomRepo.AddChatMessage(ctx, client.roomCode, chatMsg)

	msgBytes, _ := json.Marshal(chatMsg)
	h.BroadcastToRoom(client.roomCode, domain.WSMessage{
		Type:    "room:chat_message",
		Payload: msgBytes,
	})
}

func (h *Hub) handleUpdateSettings(client *Client, payload json.RawMessage) {
	ctx := context.Background()
	room, err := h.roomRepo.GetRoom(ctx, client.roomCode)
	if err != nil || room.MasterID != client.userID {
		client.SendError("Seul le Master peut modifier les paramètres")
		return
	}

	if room.Status != domain.RoomStatusInLobby {
		client.SendError("Impossible de modifier les paramètres pendant une manche")
		return
	}

	var newSettings domain.RoomSettings
	if err := json.Unmarshal(payload, &newSettings); err != nil {
		return
	}

	// Valider les bornes
	if newSettings.WordLength < 3 || newSettings.WordLength > 8 {
		newSettings.WordLength = 5
	}
	if newSettings.RoundDuration < 30 || newSettings.RoundDuration > 180 {
		newSettings.RoundDuration = 60
	}
	if newSettings.MaxRounds < 1 || newSettings.MaxRounds > 10 {
		newSettings.MaxRounds = 3
	}
	if newSettings.MaxAttempts < 4 || newSettings.MaxAttempts > 8 {
		newSettings.MaxAttempts = 6
	}

	_ = h.roomRepo.UpdateRoomSettings(ctx, client.roomCode, newSettings)
	h.SyncRoom(client.roomCode)
}

func (h *Hub) handleKick(client *Client, payload json.RawMessage) {
	ctx := context.Background()
	room, err := h.roomRepo.GetRoom(ctx, client.roomCode)
	if err != nil || room.MasterID != client.userID {
		client.SendError("Action réservée au Master")
		return
	}

	var body struct {
		TargetID uuid.UUID `json:"target_id"`
	}
	if err := json.Unmarshal(payload, &body); err != nil || body.TargetID == client.userID {
		return
	}

	targetPlayer, _ := h.roomRepo.GetPlayer(ctx, client.roomCode, body.TargetID)
	name := "Un joueur"
	if targetPlayer != nil {
		name = targetPlayer.Nickname
	}

	// Fermer la connexion du joueur ciblé
	h.roomsMu.RLock()
	if clients, ok := h.roomClients[client.roomCode]; ok {
		if targetClient, found := clients[body.TargetID]; found {
			targetClient.SendError("Vous avez été expulsé de la salle par le Master")
			go func(c *Client) {
				time.Sleep(200 * time.Millisecond)
				c.Close()
			}(targetClient)
		}
	}
	h.roomsMu.RUnlock()

	_ = h.roomRepo.RemovePlayer(ctx, client.roomCode, body.TargetID)
	h.BroadcastSystemMessage(client.roomCode, fmt.Sprintf("👢 %s a été expulsé par le Master", name))
	h.SyncRoom(client.roomCode)
}

func (h *Hub) handleBan(client *Client, payload json.RawMessage) {
	ctx := context.Background()
	room, err := h.roomRepo.GetRoom(ctx, client.roomCode)
	if err != nil || room.MasterID != client.userID {
		client.SendError("Action réservée au Master")
		return
	}

	var body struct {
		TargetID uuid.UUID `json:"target_id"`
	}
	if err := json.Unmarshal(payload, &body); err != nil || body.TargetID == client.userID {
		return
	}

	targetPlayer, _ := h.roomRepo.GetPlayer(ctx, client.roomCode, body.TargetID)
	name := "Un joueur"
	if targetPlayer != nil {
		name = targetPlayer.Nickname
	}

	h.roomsMu.RLock()
	if clients, ok := h.roomClients[client.roomCode]; ok {
		if targetClient, found := clients[body.TargetID]; found {
			targetClient.SendError("Vous avez été banni de la salle par le Master")
			go func(c *Client) {
				time.Sleep(200 * time.Millisecond)
				c.Close()
			}(targetClient)
		}
	}
	h.roomsMu.RUnlock()

	_ = h.roomRepo.BanPlayer(ctx, client.roomCode, body.TargetID)
	h.BroadcastSystemMessage(client.roomCode, fmt.Sprintf("⛔ %s a été banni de la salle", name))
	h.SyncRoom(client.roomCode)
}

func (h *Hub) handleMute(client *Client, payload json.RawMessage) {
	ctx := context.Background()
	room, err := h.roomRepo.GetRoom(ctx, client.roomCode)
	if err != nil || room.MasterID != client.userID {
		client.SendError("Action réservée au Master")
		return
	}

	var body struct {
		TargetID uuid.UUID `json:"target_id"`
		Mute     bool      `json:"mute"`
	}
	if err := json.Unmarshal(payload, &body); err != nil || body.TargetID == client.userID {
		return
	}

	_ = h.roomRepo.SetPlayerMuted(ctx, client.roomCode, body.TargetID, body.Mute)
	targetPlayer, _ := h.roomRepo.GetPlayer(ctx, client.roomCode, body.TargetID)
	name := "Un joueur"
	if targetPlayer != nil {
		name = targetPlayer.Nickname
	}

	action := "rendu muet"
	if !body.Mute {
		action = "autorisé à parler"
	}
	h.BroadcastSystemMessage(client.roomCode, fmt.Sprintf("🔇 %s a été %s par le Master", name, action))
	h.SyncRoom(client.roomCode)
}

func (h *Hub) handleRematch(client *Client) {
	ctx := context.Background()
	room, err := h.roomRepo.GetRoom(ctx, client.roomCode)
	if err != nil || room.MasterID != client.userID {
		client.SendError("Seul le Master peut relancer la salle")
		return
	}

	// Remettre la salle en lobby, conserver les scores et réintégrer les spectateurs
	_ = h.roomRepo.UpdateRoomStatus(ctx, client.roomCode, domain.RoomStatusInLobby)
	_ = h.roomRepo.UpdateRound(ctx, client.roomCode, 0, nil)

	// Tous les spectateurs redeviennent joueurs
	players, _ := h.roomRepo.GetPlayers(ctx, client.roomCode)
	for _, p := range players {
		if p.IsSpectator {
			_ = h.roomRepo.SetPlayerSpectator(ctx, client.roomCode, p.ID, false)
		}
	}

	h.BroadcastSystemMessage(client.roomCode, "🔄 Le Master a relancé la salle en Lobby pour une revanche !")
	h.SyncRoom(client.roomCode)
}

func (h *Hub) SyncRoom(roomCode string) {
	ctx := context.Background()
	room, err := h.roomRepo.GetRoom(ctx, roomCode)
	if err != nil {
		return
	}

	// Ne jamais exposer le mot secret en cours de manche aux clients
	roomSafe := *room
	if roomSafe.Status == domain.RoomStatusInGame {
		roomSafe.SecretWord = ""
	}

	payload, err := json.Marshal(roomSafe)
	if err != nil {
		return
	}

	h.BroadcastToRoom(roomCode, domain.WSMessage{
		Type:    "room:sync",
		Payload: payload,
	})
}

func (h *Hub) BroadcastSystemMessage(roomCode, message string) {
	ctx := context.Background()
	sysMsg := &domain.ChatMessage{
		ID:        uuid.New().String(),
		Sender:    "Système",
		Mascot:    "meeple",
		Color:     "#1D4ED8",
		Content:   message,
		IsSystem:  true,
		CreatedAt: time.Now(),
	}
	_ = h.roomRepo.AddChatMessage(ctx, roomCode, sysMsg)

	payload, _ := json.Marshal(sysMsg)
	h.BroadcastToRoom(roomCode, domain.WSMessage{
		Type:    "room:chat_message",
		Payload: payload,
	})
}

func (h *Hub) BroadcastToRoom(roomCode string, msg domain.WSMessage) {
	h.roomsMu.RLock()
	clients := make([]*Client, 0)
	if roomMap, ok := h.roomClients[roomCode]; ok {
		for _, c := range roomMap {
			clients = append(clients, c)
		}
	}
	h.roomsMu.RUnlock()

	for _, c := range clients {
		c.Send(msg)
	}
}

func (h *Hub) GetRoomClients(roomCode string) []*Client {
	h.roomsMu.RLock()
	defer h.roomsMu.RUnlock()

	clients := make([]*Client, 0)
	if roomMap, ok := h.roomClients[roomCode]; ok {
		for _, c := range roomMap {
			clients = append(clients, c)
		}
	}
	return clients
}
