package wordle

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"minigames-backend/internal/domain"
	redisRepo "minigames-backend/internal/repository/redis"
	"minigames-backend/internal/transport/ws"
)

type PlayerRoundState struct {
	UserID       uuid.UUID                `json:"user_id"`
	Nickname     string                   `json:"nickname"`
	Mascot       string                   `json:"mascot"`
	Color        string                   `json:"color"`
	Attempts     [][]TileEvaluation       `json:"attempts"`    // Visible seulement par le joueur
	MaskedRows   [][]MaskedTileEvaluation `json:"masked_rows"` // Diffusable aux adversaires (sans lettres)
	IsSolved     bool                     `json:"is_solved"`
	IsFinished   bool                     `json:"is_finished"`
	SolveTimeSec int                      `json:"solve_time_sec"`
	RoundScore   int                      `json:"round_score"`
}

type RoundState struct {
	TargetWord  string                          `json:"target_word"` // SECRET ABSOLU SERVEUR
	WordLength  int                             `json:"word_length"`
	MaxAttempts int                             `json:"max_attempts"`
	StartedAt   time.Time                       `json:"started_at"`
	EndsAt      time.Time                       `json:"ends_at"`
	Players     map[uuid.UUID]*PlayerRoundState `json:"players"`
	IsCompleted bool                            `json:"is_completed"`
}

type WordleGameManager struct {
	mu           sync.RWMutex
	dict         *Dictionary
	hub          *ws.Hub
	roomRepo     *redisRepo.RoomRepository
	redisClient  *redis.Client
	activeRounds map[string]*RoundState // roomCode -> RoundState

	// Timer de transition entre manches
	roundTimersMu sync.Mutex
	nextRoundTimers map[string]*time.Timer // roomCode -> Timer
}

func NewWordleGameManager(
	dict *Dictionary,
	hub *ws.Hub,
	roomRepo *redisRepo.RoomRepository,
	redisClient *redis.Client,
) *WordleGameManager {
	mgr := &WordleGameManager{
		dict:            dict,
		hub:             hub,
		roomRepo:        roomRepo,
		redisClient:     redisClient,
		activeRounds:    make(map[string]*RoundState),
		nextRoundTimers: make(map[string]*time.Timer),
	}
	hub.SetGameHandler(mgr)
	return mgr
}

func (m *WordleGameManager) HandleGameAction(client *ws.Client, action string, payload json.RawMessage) {
	ctx := context.Background()

	switch action {
	case "game:start":
		m.handleStartGame(ctx, client)
	case "game:next_round", "game:start_round":
		m.handleNextRound(ctx, client)
	case "game:guess":
		m.handleGuess(ctx, client, payload)
	case "game:get_state":
		m.handleGetState(client)
	case "game:stop", "room:return_lobby", "room:rematch":
		m.handleStopOrReturnLobby(ctx, client, action)
	}
}

func (m *WordleGameManager) handleNextRound(ctx context.Context, client *ws.Client) {
	room, err := m.roomRepo.GetRoom(ctx, client.RoomCode())
	if err != nil {
		client.SendError("Salle introuvable")
		return
	}

	if room.MasterID != client.UserID() {
		client.SendError("Seul le Master peut lancer la manche suivante")
		return
	}

	// Couper le timer de compte à rebours automatique
	m.roundTimersMu.Lock()
	if timer, exists := m.nextRoundTimers[client.RoomCode()]; exists {
		timer.Stop()
		delete(m.nextRoundTimers, client.RoomCode())
	}
	m.roundTimersMu.Unlock()

	// Démarrer la manche suivante si la partie est encore active
	if room.CurrentRound < room.Settings.MaxRounds {
		m.startRound(ctx, room, room.CurrentRound+1)
	} else {
		// Revenir au lobby si toutes les manches sont jouées
		_ = m.roomRepo.UpdateRoomStatus(ctx, client.RoomCode(), domain.RoomStatusInLobby)
		m.hub.BroadcastSystemMessage(client.RoomCode(), "Partie terminée. Retrouvez le classement général.")
		m.hub.SyncRoom(client.RoomCode())
	}
}

func (m *WordleGameManager) handleStopOrReturnLobby(ctx context.Context, client *ws.Client, action string) {
	room, err := m.roomRepo.GetRoom(ctx, client.RoomCode())
	if err != nil {
		client.SendError("Salle introuvable")
		return
	}

	if room.MasterID != client.UserID() {
		client.SendError("Action réservée au Master")
		return
	}

	// Annuler tout timer de prochaine manche
	m.roundTimersMu.Lock()
	if timer, exists := m.nextRoundTimers[client.RoomCode()]; exists {
		timer.Stop()
		delete(m.nextRoundTimers, client.RoomCode())
	}
	m.roundTimersMu.Unlock()

	// 1. Stopper la manche en cours
	m.mu.Lock()
	if round, exists := m.activeRounds[client.RoomCode()]; exists {
		round.IsCompleted = true
		delete(m.activeRounds, client.RoomCode())
	}
	m.mu.Unlock()

	// 2. Mettre à jour l'état de la salle dans Redis
	_ = m.roomRepo.UpdateRoomStatus(ctx, client.RoomCode(), domain.RoomStatusInLobby)
	_ = m.roomRepo.SetSecretWord(ctx, client.RoomCode(), "")
	_ = m.roomRepo.UpdateRound(ctx, client.RoomCode(), 0, nil)
	_ = m.roomRepo.SetRoundState(ctx, client.RoomCode(), "", "", nil)

	// 3. Réintégrer les spectateurs en joueurs
	players, _ := m.roomRepo.GetPlayers(ctx, client.RoomCode())
	for _, p := range players {
		if p.IsSpectator {
			_ = m.roomRepo.SetPlayerSpectator(ctx, client.RoomCode(), p.ID, false)
		}
	}

	// 4. Broadcaster room:state_changed et room:sync
	statePayload, _ := json.Marshal(map[string]interface{}{
		"status": domain.RoomStatusInLobby,
	})
	m.hub.BroadcastToRoom(client.RoomCode(), domain.WSMessage{
		Type:    "room:state_changed",
		Payload: statePayload,
	})

	if action == "game:stop" {
		m.hub.BroadcastSystemMessage(client.RoomCode(), "Le Master a arrêté la partie et ramené tout le monde au lobby.")
	} else if action == "room:rematch" {
		m.hub.BroadcastSystemMessage(client.RoomCode(), "Le Master a relancé la salle en Lobby pour une revanche.")
	} else {
		m.hub.BroadcastSystemMessage(client.RoomCode(), "Le Master a ramené la salle au Lobby.")
	}

	m.hub.SyncRoom(client.RoomCode())
}

func (m *WordleGameManager) handleStartGame(ctx context.Context, client *ws.Client) {
	room, err := m.roomRepo.GetRoom(ctx, client.RoomCode())
	if err != nil {
		client.SendError("Salle introuvable")
		return
	}

	if room.MasterID != client.UserID() {
		client.SendError("Seul le Master peut lancer la partie")
		return
	}

	if room.Status == domain.RoomStatusInGame {
		client.SendError("La partie est déjà en cours")
		return
	}

	// Démarrer la première manche
	m.startRound(ctx, room, 1)
}

func (m *WordleGameManager) startRound(ctx context.Context, room *domain.Room, roundNum int) {
	wordLen := room.Settings.WordLength
	if wordLen < 3 || wordLen > 8 {
		wordLen = 5
	}

	// Tirage aléatoire d'un mot secret cible de la longueur exacte
	targetWord := m.dict.PickRandomByLength(wordLen)
	duration := time.Duration(room.Settings.RoundDuration) * time.Second
	if duration == 0 {
		duration = 60 * time.Second
	}

	now := time.Now()
	endsAt := now.Add(duration)

	players, _ := m.roomRepo.GetPlayers(ctx, room.Code)

	roundState := &RoundState{
		TargetWord:  targetWord,
		WordLength:  wordLen,
		MaxAttempts: room.Settings.MaxAttempts,
		StartedAt:   now,
		EndsAt:      endsAt,
		Players:     make(map[uuid.UUID]*PlayerRoundState),
		IsCompleted: false,
	}

	for _, p := range players {
		if p.Role != domain.RoleSpectator {
			roundState.Players[p.ID] = &PlayerRoundState{
				UserID:       p.ID,
				Nickname:     p.Nickname,
				Mascot:       p.Mascot,
				Color:        p.Color,
				Attempts:     make([][]TileEvaluation, 0),
				MaskedRows:   make([][]MaskedTileEvaluation, 0),
				IsSolved:     false,
				IsFinished:   false,
				SolveTimeSec: 0,
				RoundScore:   0,
			}
		}
	}

	m.mu.Lock()
	m.activeRounds[room.Code] = roundState
	m.mu.Unlock()

	_ = m.roomRepo.UpdateRoomStatus(ctx, room.Code, domain.RoomStatusInGame)
	_ = m.roomRepo.SetSecretWord(ctx, room.Code, targetWord)
	_ = m.roomRepo.UpdateRound(ctx, room.Code, roundNum, &endsAt)
	_ = m.roomRepo.SetRoundState(ctx, room.Code, domain.RoundSubStatePlaying, "", nil)

	// Broadcaster le début de manche avec horodatage absolu ends_at
	startPayload, _ := json.Marshal(map[string]interface{}{
		"round":          roundNum,
		"max_rounds":     room.Settings.MaxRounds,
		"word_length":    wordLen,
		"max_attempts":   room.Settings.MaxAttempts,
		"ends_at":        endsAt.Format(time.RFC3339),
		"round_duration": int(duration.Seconds()),
	})

	m.hub.BroadcastToRoom(room.Code, domain.WSMessage{
		Type:    "game:round_start",
		Payload: startPayload,
	})

	m.hub.SyncRoom(room.Code)
	m.hub.BroadcastSystemMessage(room.Code, fmt.Sprintf("Manche %d lancée ! Mot secret de %d lettres.", roundNum, wordLen))

	// Timer de fin de manche automatique
	go func(roomCode string, rNum int, targetEndsAt time.Time) {
		time.Sleep(time.Until(targetEndsAt))
		m.mu.Lock()
		activeRound, exists := m.activeRounds[roomCode]
		m.mu.Unlock()

		if exists && !activeRound.IsCompleted {
			m.endRound(context.Background(), roomCode, "Temps écoulé !")
		}
	}(room.Code, roundNum, endsAt)
}

func (m *WordleGameManager) handleGuess(ctx context.Context, client *ws.Client, payload json.RawMessage) {
	var body struct {
		Guess string `json:"guess"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		client.SendError("Format de proposition invalide")
		return
	}

	m.mu.RLock()
	round, ok := m.activeRounds[client.RoomCode()]
	m.mu.RUnlock()

	if !ok || round.IsCompleted {
		client.SendError("Aucune manche active dans cette salle")
		return
	}

	playerState, exists := round.Players[client.UserID()]
	if !exists {
		client.SendError("Vous ne participez pas à cette manche")
		return
	}

	if playerState.IsFinished {
		client.SendError("Vous avez déjà terminé vos essais pour cette manche")
		return
	}

	guess := strings.ToUpper(strings.TrimSpace(body.Guess))
	if len(guess) != round.WordLength {
		client.SendError(fmt.Sprintf("Le mot doit comporter exactement %d lettres", round.WordLength))
		return
	}

	// Validation du mot dans le dictionnaire étendu (targets + allowed)
	if !m.dict.IsValid(guess) {
		client.SendError(fmt.Sprintf("'%s' n'est pas un mot valide du dictionnaire français", guess))
		return
	}

	// Évaluation de la proposition
	eval, isSolved := EvaluateGuess(round.TargetWord, guess)
	masked := MaskEvaluation(eval)

	m.mu.Lock()
	playerState.Attempts = append(playerState.Attempts, eval)
	playerState.MaskedRows = append(playerState.MaskedRows, masked)

	if isSolved {
		playerState.IsSolved = true
		playerState.IsFinished = true
		playerState.SolveTimeSec = int(time.Since(round.StartedAt).Seconds())

		// Calcul des points de la manche (essais restants + bonus temps)
		attemptsUsed := len(playerState.Attempts)
		remainingAttempts := round.MaxAttempts - attemptsUsed + 1
		timeBonus := int(round.EndsAt.Sub(time.Now()).Seconds()) / 5
		if timeBonus < 0 {
			timeBonus = 0
		}
		playerState.RoundScore = (remainingAttempts * 100) + timeBonus

	} else if len(playerState.Attempts) >= round.MaxAttempts {
		playerState.IsFinished = true
		playerState.IsSolved = false
		playerState.RoundScore = 0
	}
	m.mu.Unlock()

	// 1. Envoyer le résultat complet (avec lettres) au joueur qui a deviné
	playerData, _ := json.Marshal(map[string]interface{}{
		"attempts":    playerState.Attempts,
		"is_solved":   playerState.IsSolved,
		"is_finished": playerState.IsFinished,
		"round_score": playerState.RoundScore,
	})
	client.Send(domain.WSMessage{
		Type:    "game:guess_result",
		Payload: playerData,
	})

	// 2. Diffuser aux adversaires le State Masking (SANS LES LETTRES)
	opponentData, _ := json.Marshal(map[string]interface{}{
		"user_id":      playerState.UserID,
		"nickname":     playerState.Nickname,
		"mascot":       playerState.Mascot,
		"color":        playerState.Color,
		"masked_rows":  playerState.MaskedRows,
		"is_solved":    playerState.IsSolved,
		"is_finished":  playerState.IsFinished,
		"attempts_cnt": len(playerState.MaskedRows),
	})
	m.hub.BroadcastToRoom(client.RoomCode(), domain.WSMessage{
		Type:    "game:opponent_progress",
		Payload: opponentData,
	})

	// Notification sobre sans émoji
	if isSolved {
		m.hub.BroadcastSystemMessage(client.RoomCode(), fmt.Sprintf("%s a trouvé le mot en %d essai(s) !", playerState.Nickname, len(playerState.Attempts)))
	}

	// Vérifier si tous les joueurs actifs ont terminé la manche
	allFinished := true
	m.mu.RLock()
	for _, p := range round.Players {
		if !p.IsFinished {
			allFinished = false
			break
		}
	}
	m.mu.RUnlock()

	if allFinished {
		// Délai de 1,5 seconde pour laisser le temps de voir l'animation de victoire/défaite
		go func(rCode string) {
			time.Sleep(1500 * time.Millisecond)
			m.endRound(context.Background(), rCode, "Tous les joueurs ont terminé !")
		}(client.RoomCode())
	}
}

func (m *WordleGameManager) endRound(ctx context.Context, roomCode, reason string) {
	m.mu.Lock()
	round, ok := m.activeRounds[roomCode]
	if !ok || round.IsCompleted {
		m.mu.Unlock()
		return
	}
	round.IsCompleted = true
	m.mu.Unlock()

	room, err := m.roomRepo.GetRoom(ctx, roomCode)
	if err != nil {
		return
	}

	// Mise à jour des scores dans Redis
	roundScoresMap := make(map[string]int)
	roundSolveTimes := make(map[string]int)
	var winnerID *uuid.UUID
	var winnerName string
	highestRoundScore := -1

	for uid, p := range round.Players {
		roundScoresMap[uid.String()] = p.RoundScore
		roundSolveTimes[uid.String()] = p.SolveTimeSec
		if p.RoundScore > 0 {
			_, _ = m.roomRepo.AddScore(ctx, roomCode, uid, p.RoundScore)
		}
		if p.IsSolved && p.RoundScore > highestRoundScore {
			highestRoundScore = p.RoundScore
			wID := uid
			winnerID = &wID
			winnerName = p.Nickname
		}
	}

	// Enregistrer dans l'historique
	_ = m.roomRepo.AddRoundHistory(ctx, roomCode, &domain.RoundHistory{
		RoundNum:   room.CurrentRound,
		SecretWord: round.TargetWord,
		WinnerID:   winnerID,
		WinnerName: winnerName,
		Scores:     roundScoresMap,
		EndedAt:    time.Now(),
	})

	allScores, _ := m.roomRepo.GetScores(ctx, roomCode)

	// Broadcaster la fin de manche avec RÉVÉLATION DU MOT et compte à rebours de 8s
	countdownSec := 8
	endPayload, _ := json.Marshal(map[string]interface{}{
		"round":         room.CurrentRound,
		"max_rounds":    room.Settings.MaxRounds,
		"secret_word":   round.TargetWord,
		"reason":        reason,
		"winner_name":   winnerName,
		"round_scores":  roundScoresMap,
		"solve_times":   roundSolveTimes,
		"total_scores":  allScores,
		"countdown_sec": countdownSec,
	})

	m.hub.BroadcastToRoom(roomCode, domain.WSMessage{
		Type:    "game:round_end",
		Payload: endPayload,
	})

	m.hub.BroadcastSystemMessage(roomCode, fmt.Sprintf("Fin de manche. Le mot était : %s", round.TargetWord))

	// Vérifier s'il reste des manches à jouer
	if room.CurrentRound < room.Settings.MaxRounds {
		nextRoundAt := time.Now().Add(time.Duration(countdownSec) * time.Second)
		_ = m.roomRepo.SetRoundState(ctx, roomCode, domain.RoundSubStateRoundEnded, round.TargetWord, &nextRoundAt)
		m.hub.SyncRoom(roomCode)

		// Compte à rebours automatique avant la manche suivante (annulable par le Master)
		m.roundTimersMu.Lock()
		if oldTimer, exists := m.nextRoundTimers[roomCode]; exists {
			oldTimer.Stop()
		}
		nextRoundNum := room.CurrentRound + 1
		m.nextRoundTimers[roomCode] = time.AfterFunc(time.Duration(countdownSec)*time.Second, func() {
			m.roundTimersMu.Lock()
			delete(m.nextRoundTimers, roomCode)
			m.roundTimersMu.Unlock()

			r, getErr := m.roomRepo.GetRoom(context.Background(), roomCode)
			if getErr == nil && r.Status == domain.RoomStatusInGame {
				m.startRound(context.Background(), r, nextRoundNum)
			}
		})
		m.roundTimersMu.Unlock()
	} else {
		// Partie terminée !
		_ = m.roomRepo.SetRoundState(ctx, roomCode, domain.RoundSubStateGameOver, round.TargetWord, nil)
		_ = m.roomRepo.UpdateRoomStatus(ctx, roomCode, domain.RoomStatusInLobby)
		m.hub.BroadcastSystemMessage(roomCode, "Partie terminée ! Retrouvez le classement général.")
		m.hub.SyncRoom(roomCode)
	}
}

func (m *WordleGameManager) handleGetState(client *ws.Client) {
	m.mu.RLock()
	round, ok := m.activeRounds[client.RoomCode()]
	m.mu.RUnlock()

	if !ok {
		return
	}

	m.mu.RLock()
	playerState := round.Players[client.UserID()]
	opponents := make([]map[string]interface{}, 0)
	for uid, p := range round.Players {
		if uid != client.UserID() {
			opponents = append(opponents, map[string]interface{}{
				"user_id":      p.UserID,
				"nickname":     p.Nickname,
				"mascot":       p.Mascot,
				"color":        p.Color,
				"masked_rows":  p.MaskedRows,
				"is_solved":    p.IsSolved,
				"is_finished":  p.IsFinished,
				"attempts_cnt": len(p.MaskedRows),
			})
		}
	}
	m.mu.RUnlock()

	var attempts [][]TileEvaluation
	isSolved := false
	isFinished := false
	if playerState != nil {
		attempts = playerState.Attempts
		isSolved = playerState.IsSolved
		isFinished = playerState.IsFinished
	}

	statePayload, _ := json.Marshal(map[string]interface{}{
		"word_length":  round.WordLength,
		"max_attempts": round.MaxAttempts,
		"ends_at":      round.EndsAt.Format(time.RFC3339),
		"attempts":     attempts,
		"is_solved":    isSolved,
		"is_finished":  isFinished,
		"opponents":    opponents,
	})

	client.Send(domain.WSMessage{
		Type:    "game:state_sync",
		Payload: statePayload,
	})
}

func (m *WordleGameManager) OnPlayerJoined(client *ws.Client, room *domain.Room) {
	m.mu.Lock()
	defer m.mu.Unlock()

	round, ok := m.activeRounds[room.Code]
	if !ok || round.IsCompleted {
		return
	}

	// Si le joueur n'était pas dans la manche active, le marquer comme spectateur
	if _, exists := round.Players[client.UserID()]; !exists {
		_ = m.roomRepo.SetPlayerSpectator(context.Background(), room.Code, client.UserID(), true)
	}

	// Lui envoyer l'état actuel de la manche
	go m.handleGetState(client)
}

func (m *WordleGameManager) OnPlayerLeft(roomCode string, userID uuid.UUID) {
	// Période de grâce de 45s gérée par Redis et les workers
}
