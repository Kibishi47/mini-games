package games

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"minigames-backend/internal/domain"
	redisRepo "minigames-backend/internal/repository/redis"
	"minigames-backend/internal/transport/ws"
)

// Router aiguille les événements du Hub WebSocket vers le moteur de jeu correspondant
// au type de partie configuré pour la salle ("wordle" ou "poker"). Le Hub ne peut porter
// qu'un seul GameEventHandler global : ce Router en est l'unique point d'entrée.
type Router struct {
	roomRepo *redisRepo.RoomRepository
	handlers map[string]ws.GameEventHandler
}

func NewRouter(roomRepo *redisRepo.RoomRepository) *Router {
	return &Router{
		roomRepo: roomRepo,
		handlers: make(map[string]ws.GameEventHandler),
	}
}

// Register associe un moteur de jeu à un type de partie (ex: "wordle", "poker")
func (r *Router) Register(gameType string, handler ws.GameEventHandler) {
	r.handlers[gameType] = handler
}

func (r *Router) resolveHandler(gameType string) ws.GameEventHandler {
	if h, ok := r.handlers[gameType]; ok {
		return h
	}
	return r.handlers[domain.GameTypeWordle]
}

func (r *Router) HandleGameAction(client *ws.Client, action string, payload json.RawMessage) {
	ctx := context.Background()
	room, err := r.roomRepo.GetRoom(ctx, client.RoomCode())
	if err != nil {
		return
	}
	if h := r.resolveHandler(room.Settings.GameType); h != nil {
		h.HandleGameAction(client, action, payload)
	}
}

func (r *Router) OnPlayerJoined(client *ws.Client, room *domain.Room) {
	if h := r.resolveHandler(room.Settings.GameType); h != nil {
		h.OnPlayerJoined(client, room)
	}
}

func (r *Router) OnPlayerLeft(roomCode string, userID uuid.UUID) {
	ctx := context.Background()
	room, err := r.roomRepo.GetRoom(ctx, roomCode)
	if err != nil {
		// Salle déjà fermée (dernier joueur parti) : aucun moteur de jeu n'a besoin d'être notifié
		return
	}
	if h := r.resolveHandler(room.Settings.GameType); h != nil {
		h.OnPlayerLeft(roomCode, userID)
	}
}
