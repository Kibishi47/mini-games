package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"minigames-backend/internal/domain"
	"minigames-backend/internal/repository/postgres"
	"minigames-backend/internal/repository/redis"
	"minigames-backend/internal/transport/ws"
)

type WorkerManager struct {
	roomRepo    *redis.RoomRepository
	sessionRepo *postgres.RoomSessionRepository
	hub         *ws.Hub
}

func NewWorkerManager(roomRepo *redis.RoomRepository, sessionRepo *postgres.RoomSessionRepository, hub *ws.Hub) *WorkerManager {
	return &WorkerManager{
		roomRepo:    roomRepo,
		sessionRepo: sessionRepo,
		hub:         hub,
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
			// Si le joueur est déconnecté depuis plus de 45s
			if !p.IsConnected && now.Sub(p.LastSeenAt) > gracePeriod {
				log.Printf("⚠️ [AFK] Joueur %s (%s) expulsé après 45s de déconnexion de la room %s\n", p.DisplayUsername, p.UserID, code)

				isMaster := p.Role == domain.RoleMaster
				_ = w.roomRepo.RemovePlayer(ctx, code, p.UserID)

				w.hub.BroadcastSystemMessage(code, fmt.Sprintf("%s a été retiré de la salle (délai de reconnexion dépassé)", p.DisplayUsername))

				if isMaster {
					w.hub.PromoteNextMaster(ctx, code)
				}
				w.hub.SyncRoom(code)
			}
		}
	}
}

// runRoomGarbageCollector nettoie les rooms inactives (>30min en lobby ou vides depuis >2min)
func (w *WorkerManager) runRoomGarbageCollector(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
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

		shouldClose := false
		reason := ""

		// Condition 1 : Aucun joueur dans la salle
		if len(room.Players) == 0 {
			shouldClose = true
			reason = "salle vide"
		} else {
			// Vérifier si aucun joueur n'est connecté depuis plus de 2 minutes
			allDisconnectedFor2Min := true
			for _, p := range room.Players {
				if p.IsConnected || now.Sub(p.LastSeenAt) < 2*time.Minute {
					allDisconnectedFor2Min = false
					break
				}
			}
			if allDisconnectedFor2Min {
				shouldClose = true
				reason = "tous les joueurs sont déconnectés depuis >2min"
			}
		}

		// Condition 2 : Bloqué en lobby depuis > 30 minutes
		if room.Status == domain.RoomStatusInLobby && now.Sub(room.CreatedAt) > 30*time.Minute {
			shouldClose = true
			reason = "lobby inactif depuis >30min"
		}

		if shouldClose {
			log.Printf("🧹 [GC] Fermeture et purge de la salle %s (%s)\n", code, reason)
			_ = w.roomRepo.UpdateRoomStatus(ctx, code, domain.RoomStatusClosed)
			_ = w.sessionRepo.UpdateStatus(ctx, code, domain.RoomStatusClosed)
			_ = w.roomRepo.CloseAndPurgeRoom(ctx, code)
		}
	}
}
