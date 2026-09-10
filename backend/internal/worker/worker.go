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

	gracePeriod := 45 * time.Second
	now := time.Now()

	for _, code := range codes {
		players, err := w.roomRepo.GetPlayers(ctx, code)
		if err != nil {
			continue
		}

		for _, p := range players {
			// Si marqué déconnecté depuis plus de 45s
			if !p.IsConnected && now.Sub(p.LastSeenAt) > gracePeriod {
				_ = w.roomRepo.RemovePlayer(ctx, code, p.ID)
				w.hub.HandleMasterSuccession(code, p.ID)
				w.hub.SyncRoom(code)
			}
		}
	}
}

// runRoomGarbageCollector purge les rooms vides depuis plus de 2min ou les lobbies abandonnés depuis plus de 30min
func (w *WorkerManager) runRoomGarbageCollector(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
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
			continue
		}

		players, _ := w.roomRepo.GetPlayers(ctx, code)
		connectedCount := 0
		for _, p := range players {
			if p.IsConnected {
				connectedCount++
			}
		}

		// 1. Room totalement vide depuis plus de 2 minutes
		if len(players) == 0 || connectedCount == 0 {
			if now.Sub(room.CreatedAt) > 2*time.Minute {
				_ = w.roomRepo.CloseRoom(ctx, code)
				log.Printf("🧹 [GC] Salle vide %s fermée", code)
				continue
			}
		}

		// 2. Lobby inactif abandonné depuis plus de 30 minutes
		if room.Status == domain.RoomStatusInLobby && now.Sub(room.CreatedAt) > 30*time.Minute {
			_ = w.roomRepo.CloseRoom(ctx, code)
			log.Printf("🧹 [GC] Lobby abandonné %s fermé", code)
		}
	}
}
