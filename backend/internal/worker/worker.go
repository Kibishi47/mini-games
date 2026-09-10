package worker

import (
	"context"
	"log"
	"time"

	"minigames-backend/internal/domain"
	"minigames-backend/internal/repository/redis"
	"minigames-backend/internal/transport/ws"
)

type WorkerManager struct {
	roomRepo *redis.RoomRepository
	hub      *ws.Hub
}

func NewWorkerManager(roomRepo *redis.RoomRepository, hub *ws.Hub) *WorkerManager {
	return &WorkerManager{
		roomRepo: roomRepo,
		hub:      hub,
	}
}

// Start lance les deux goroutines d'arrière-plan de surveillance et nettoyage
func (w *WorkerManager) Start(ctx context.Context) {
	go w.runAFKHeartbeatWorker(ctx)
	go w.runRoomGarbageCollector(ctx)
	log.Println("✅ [Workers] AFK Monitor et Room GC démarrés")
}

// runAFKHeartbeatWorker vérifie toutes les 15s si des joueurs ont dépassé la période de grâce de 45s
func (w *WorkerManager) runAFKHeartbeatWorker(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.checkAFKPlayers(ctx)
		}
	}
}

func (w *WorkerManager) checkAFKPlayers(ctx context.Context) {
	codes, err := w.roomRepo.GetActiveRoomCodes(ctx)
	if err != nil {
		return
	}

	gracePeriod := 10 * time.Second
	now := time.Now()

	for _, code := range codes {
		players, err := w.roomRepo.GetPlayers(ctx, code)
		if err != nil {
			continue
		}

		for _, p := range players {
			// Si marqué déconnecté depuis plus de 10s
			if !p.IsConnected && now.Sub(p.LastSeenAt) > gracePeriod {
				_ = w.roomRepo.RemovePlayer(ctx, code, p.ID)
				w.hub.HandleMasterSuccession(code, p.ID)
				w.hub.SyncRoom(code)
			}
		}
	}
}

// runRoomGarbageCollector purge les rooms selon les critères stricts de cycle de vie :
// - Room vide : aucun joueur connecté depuis plus de 2 minutes
// - Lobby abandonné : room en in_lobby depuis plus de 20 minutes
// - Partie bloquée : room en in_game sans mise à jour depuis plus de 10 minutes
func (w *WorkerManager) runRoomGarbageCollector(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.cleanInactiveRooms(ctx)
		}
	}
}

func (w *WorkerManager) cleanInactiveRooms(ctx context.Context) {
	codes, err := w.roomRepo.GetActiveRoomCodes(ctx)
	if err != nil {
		return
	}

	now := time.Now()

	for _, code := range codes {
		room, err := w.roomRepo.GetRoom(ctx, code)
		if err != nil {
			// Clé ou salle déjà disparue, s'assurer qu'elle n'est plus dans rooms:active
			_ = w.roomRepo.CloseRoom(ctx, code)
			continue
		}

		players, _ := w.roomRepo.GetPlayers(ctx, code)
		connectedCount := 0
		var latestActivity time.Time = room.CreatedAt
		for _, p := range players {
			if p.IsConnected {
				connectedCount++
			}
			if p.LastSeenAt.After(latestActivity) {
				latestActivity = p.LastSeenAt
			}
		}

		// 1. Room vide : aucun joueur connecté depuis plus de 2 minutes
		if connectedCount == 0 {
			if now.Sub(latestActivity) > 2*time.Minute {
				_ = w.roomRepo.CloseRoom(ctx, code)
				log.Printf("🧹 [GC] Salle vide %s fermée (inactive > 2min)", code)
				continue
			}
		}

		// 2. Lobby abandonné : statut in_lobby sans activité depuis plus de 20 minutes
		if room.Status == domain.RoomStatusInLobby {
			if now.Sub(latestActivity) > 20*time.Minute {
				_ = w.roomRepo.CloseRoom(ctx, code)
				log.Printf("🧹 [GC] Lobby abandonné %s fermé (> 20min)", code)
				continue
			}
		}

		// 3. Partie bloquée : statut in_game sans activité depuis plus de 10 minutes
		if room.Status == domain.RoomStatusInGame {
			if now.Sub(latestActivity) > 10*time.Minute {
				_ = w.roomRepo.CloseRoom(ctx, code)
				log.Printf("🧹 [GC] Partie bloquée %s fermée (> 10min)", code)
			}
		}
	}
}
