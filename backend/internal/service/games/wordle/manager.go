package wordle

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"minigames-backend/internal/domain"
	"minigames-backend/internal/repository/postgres"
	redisRepo "minigames-backend/internal/repository/redis"
	"minigames-backend/internal/transport/ws"
)

type PlayerRoundState struct {
	UserID       uuid.UUID                 `json:"user_id"`
	Attempts     [][]TileEvaluation        `json:"attempts"`      // Visible seulement par le joueur
	MaskedRows   [][]MaskedTileEvaluation  `json:"masked_rows"`   // Diffusable aux adversaires
	IsSolved     bool                      `json:"is_solved"`
	IsFinished   bool                      `json:"is_finished"`
	SolveTimeSec int                       `json:"solve_time_sec"`
	RoundScore   int                       `json:"round_score"`
}

type RoundState struct {
	TargetWord  string                       `json:"target_word"` // SECRET ABSOLU SERVEUR
	WordLength  int                          `json:"word_length"`
	MaxAttempts int                          `json:"max_attempts"`
	StartedAt   time.Time                    `json:"started_at"`
	EndsAt      time.Time                    `json:"ends_at"`
	Players     map[uuid.UUID]*PlayerRoundState `json:"players"`
	IsCompleted bool                         `json:"is_completed"`
}

type WordleGameManager struct {
	mu          sync.RWMutex
	dict        *Dictionary
	hub         *ws.Hub
	roomRepo    *redisRepo.RoomRepository
	sessionRepo *postgres.RoomSessionRepository
	userRepo    *postgres.UserRepository
	redisClient *redis.Client
	activeRounds map[string]*RoundState // roomCode -> RoundState
}

func NewWordleGameManager(
	dict *Dictionary,
	hub *ws.Hub,
	roomRepo *redisRepo.RoomRepository,
	sessionRepo *postgres.RoomSessionRepository,
	userRepo *postgres.UserRepository,
	redisClient *redis.Client,
) *WordleGameManager {
	mgr := &WordleGameManager{
		dict:         dict,
		hub:          hub,
		roomRepo:     roomRepo,
		sessionRepo:  sessionRepo,
		userRepo:     userRepo,
		redisClient:  redisClient,
		activeRounds: make(map[string]*RoundState),
	}
	hub.SetGameHandler(mgr)
	return mgr
}

func (m *WordleGameManager) HandleGameAction(client *ws.Client, action string, payload json.RawMessage) {
	ctx := context.Background()

	switch action {
	case "game:start":
		m.handleStartGame(ctx, client)
	case "game:submit_guess":
		m.handleSubmitGuess(ctx, client, payload)
	case "game:rematch":
		m.handleRematch(ctx, client)
	}
}

func (m *WordleGameManager) OnPlayerJoined(client *ws.Client, room *domain.Room) {
	// Si la room est in_game, synchroniser l'état masqué actuel avec le nouveau spectateur/joueur
	if room.Status == domain.RoomStatusInGame {
		m.syncGameStateForUser(client)
	}
}

func (m *WordleGameManager) OnPlayerLeft(roomCode string, userID uuid.UUID) {
	// Vérifier si tous les joueurs actifs ont fini la manche
	m.checkRoundCompletion(context.Background(), roomCode)
}

func (m *WordleGameManager) handleStartGame(ctx context.Context, client *ws.Client) {
	room, err := m.roomRepo.GetRoom(ctx, client.RoomCode())
	if err != nil {
		client.SendError("Salle introuvable")
		return
	}

	// Seul le Master a toute autorité pour lancer la partie
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
	if wordLen == 0 {
		wordLen = 5
	}
	lang := room.Settings.Language
	if lang == "" {
		lang = "fr"
	}

	targetWord := m.dict.GetRandomWord(lang, wordLen)
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
		// Seuls les rôles joueurs ou master participent activement
		if p.Role != domain.RoleSpectator {
			roundState.Players[p.UserID] = &PlayerRoundState{
				UserID:     p.UserID,
				Attempts:   make([][]TileEvaluation, 0),
				MaskedRows: make([][]MaskedTileEvaluation, 0),
			}
		}
	}

	m.mu.Lock()
	m.activeRounds[room.Code] = roundState
	m.mu.Unlock()

	_ = m.roomRepo.UpdateRoomStatus(ctx, room.Code, domain.RoomStatusInGame)
	_ = m.roomRepo.UpdateRound(ctx, room.Code, roundNum, &endsAt)
	_ = m.sessionRepo.UpdateStatus(ctx, room.Code, domain.RoomStatusInGame)

	// Broadcaster le début de manche avec timestamp absolu ends_at
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

	m.hub.BroadcastSystemMessage(room.Code, fmt.Sprintf("🎮 Manche %d/%d lancée ! Trouvez le mot de %d lettres en %d secondes !", roundNum, room.Settings.MaxRounds, wordLen, int(duration.Seconds())))

	m.hub.SyncRoom(room.Code)

	// Lancer un timer pour expiration automatique de la manche
	go func(code string, round int, targetEndsAt time.Time) {
		time.Sleep(time.Until(targetEndsAt))
		m.handleRoundTimeout(context.Background(), code, round)
	}(room.Code, roundNum, endsAt)
}

func (m *WordleGameManager) handleSubmitGuess(ctx context.Context, client *ws.Client, payload json.RawMessage) {
	var body struct {
		Guess string `json:"guess"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return
	}

	m.mu.Lock()
	round, ok := m.activeRounds[client.RoomCode()]
	m.mu.Unlock()

	if !ok || round.IsCompleted {
		client.SendError("Aucune manche en cours")
		return
	}

	if time.Now().After(round.EndsAt) {
		client.SendError("Le temps imparti pour cette manche est écoulé")
		return
	}

	m.mu.Lock()
	playerState, ok := round.Players[client.UserID()]
	m.mu.Unlock()

	if !ok {
		client.SendError("Vous êtes en mode spectateur pour cette manche")
		return
	}

	if playerState.IsFinished {
		client.SendError("Vous avez déjà terminé votre manche")
		return
	}

	guess := body.Guess
	if len(guess) != round.WordLength {
		client.SendError(fmt.Sprintf("Le mot doit contenir exactement %d lettres", round.WordLength))
		return
	}

	// Validation du mot dans le dictionnaire
	room, _ := m.roomRepo.GetRoom(ctx, client.RoomCode())
	lang := "fr"
	if room != nil && room.Settings.Language != "" {
		lang = room.Settings.Language
	}

	if !m.dict.IsValidWord(lang, guess) {
		client.SendError("Ce mot n'existe pas dans le dictionnaire")
		return
	}

	// Évaluation
	eval, isSolved := EvaluateGuess(round.TargetWord, guess)
	masked := MaskEvaluation(eval)

	m.mu.Lock()
	playerState.Attempts = append(playerState.Attempts, eval)
	playerState.MaskedRows = append(playerState.MaskedRows, masked)

	if isSolved {
		playerState.IsSolved = true
		playerState.IsFinished = true
		playerState.SolveTimeSec = int(time.Since(round.StartedAt).Seconds())

		// Calcul de score: base 1000 pts - pénalité par tentative - pénalité de temps
		attemptsPenalty := (len(playerState.Attempts) - 1) * 120
		timePenalty := playerState.SolveTimeSec * 5
		score := 1000 - attemptsPenalty - timePenalty
		if score < 200 {
			score = 200
		}
		playerState.RoundScore = score
	} else if len(playerState.Attempts) >= round.MaxAttempts {
		playerState.IsFinished = true
		playerState.RoundScore = 0
	}
	m.mu.Unlock()

	// 1. Envoyer le résultat complet (avec les lettres) EXCLUSIVEMENT au joueur concerné
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

	// 2. Diffuser aux AUTRES joueurs de la room l'aperçu MASQUÉ (state-masking strict, sans spoilers !)
	opponentData, _ := json.Marshal(map[string]interface{}{
		"user_id":     client.UserID(),
		"row_index":   len(playerState.Attempts) - 1,
		"masked_row":  masked,
		"is_solved":   playerState.IsSolved,
		"is_finished": playerState.IsFinished,
	})
	m.hub.BroadcastToRoom(client.RoomCode(), domain.WSMessage{
		Type:    "game:opponent_progress",
		Payload: opponentData,
	})

	if playerState.IsSolved {
		p, _ := m.roomRepo.GetPlayer(ctx, client.RoomCode(), client.UserID())
		name := "Un joueur"
		if p != nil {
			name = p.DisplayUsername
		}
		m.hub.BroadcastSystemMessage(client.RoomCode(), fmt.Sprintf("🎯 %s a trouvé le mot en %d essais (%ds) !", name, len(playerState.Attempts), playerState.SolveTimeSec))
	}

	// Vérifier si tout le monde a terminé
	m.checkRoundCompletion(ctx, client.RoomCode())
}

func (m *WordleGameManager) handleRoundTimeout(ctx context.Context, roomCode string, roundNum int) {
	m.mu.Lock()
	round, ok := m.activeRounds[roomCode]
	m.mu.Unlock()

	if !ok || round.IsCompleted {
		return
	}

	// Clôturer pour tous les joueurs qui n'ont pas encore fini
	m.mu.Lock()
	for _, p := range round.Players {
		if !p.IsFinished {
			p.IsFinished = true
			p.RoundScore = 0
		}
	}
	m.mu.Unlock()

	m.endRound(ctx, roomCode, round)
}

func (m *WordleGameManager) checkRoundCompletion(ctx context.Context, roomCode string) {
	m.mu.Lock()
	round, ok := m.activeRounds[roomCode]
	if !ok || round.IsCompleted {
		m.mu.Unlock()
		return
	}

	allFinished := true
	for _, p := range round.Players {
		if !p.IsFinished {
			allFinished = false
			break
		}
	}
	m.mu.Unlock()

	if allFinished && len(round.Players) > 0 {
		m.endRound(ctx, roomCode, round)
	}
}

func (m *WordleGameManager) endRound(ctx context.Context, roomCode string, round *RoundState) {
	m.mu.Lock()
	if round.IsCompleted {
		m.mu.Unlock()
		return
	}
	round.IsCompleted = true
	m.mu.Unlock()

	room, err := m.roomRepo.GetRoom(ctx, roomCode)
	if err != nil {
		return
	}

	sessionID, _ := m.sessionRepo.GetSessionIDByCode(ctx, roomCode)

	// Appliquer les scores et générer les grilles emojis
	type PlayerSummary struct {
		UserID          uuid.UUID `json:"user_id"`
		DisplayUsername string    `json:"display_username"`
		AvatarURL       string    `json:"avatar_url"`
		IsSolved        bool      `json:"is_solved"`
		AttemptsCount   int       `json:"attempts_count"`
		ScoreDelta      int       `json:"score_delta"`
		TotalScore      int       `json:"total_score"`
		EmojiGrid       string    `json:"emoji_grid"`
	}

	summaries := make([]PlayerSummary, 0)

	for userID, pState := range round.Players {
		// Mettre à jour le score du joueur dans Redis
		_ = m.roomRepo.UpdatePlayerScore(ctx, roomCode, userID, pState.RoundScore)
		player, _ := m.roomRepo.GetPlayer(ctx, roomCode, userID)

		totalScore := pState.RoundScore
		name := "Joueur"
		avatar := ""
		if player != nil {
			totalScore = player.Score
			name = player.DisplayUsername
			avatar = player.AvatarURL
		}

		// Persister dans Postgres
		_ = m.sessionRepo.RecordRoundScore(ctx, sessionID, userID, "wordle", room.CurrentRound, pState.RoundScore, totalScore)
		_ = m.userRepo.RecordGameResult(ctx, userID, "wordle", pState.IsSolved, totalScore)

		emojiGrid := GenerateEmojiGrid(pState.Attempts)

		summaries = append(summaries, PlayerSummary{
			UserID:          userID,
			DisplayUsername: name,
			AvatarURL:       avatar,
			IsSolved:        pState.IsSolved,
			AttemptsCount:   len(pState.Attempts),
			ScoreDelta:      pState.RoundScore,
			TotalScore:      totalScore,
			EmojiGrid:       emojiGrid,
		})
	}

	// Révélation du mot secret à la fin de manche
	endPayload, _ := json.Marshal(map[string]interface{}{
		"target_word":  round.TargetWord,
		"round":        room.CurrentRound,
		"max_rounds":   room.Settings.MaxRounds,
		"summaries":    summaries,
		"is_game_over": room.CurrentRound >= room.Settings.MaxRounds,
	})

	m.hub.BroadcastToRoom(roomCode, domain.WSMessage{
		Type:    "game:round_end",
		Payload: endPayload,
	})

	m.hub.BroadcastSystemMessage(roomCode, fmt.Sprintf("🏁 Fin de la manche %d ! Le mot secret était : %s", room.CurrentRound, round.TargetWord))

	// Passage à la manche suivante ou Game Over
	if room.CurrentRound < room.Settings.MaxRounds {
		go func() {
			time.Sleep(7 * time.Second) // Pause récapitulative
			updatedRoom, _ := m.roomRepo.GetRoom(context.Background(), roomCode)
			if updatedRoom != nil && updatedRoom.Status == domain.RoomStatusInGame {
				m.startRound(context.Background(), updatedRoom, room.CurrentRound+1)
			}
		}()
	} else {
		// Fin définitive de partie
		m.hub.BroadcastSystemMessage(roomCode, "🏆 Partie terminée ! Consultez le Scoreboard final !")
		m.hub.SyncRoom(roomCode)
	}
}

func (m *WordleGameManager) handleRematch(ctx context.Context, client *ws.Client) {
	room, err := m.roomRepo.GetRoom(ctx, client.RoomCode())
	if err != nil {
		return
	}
	if room.MasterID != client.UserID() {
		client.SendError("Seul le Master peut relancer une revanche")
		return
	}

	// Réinitialiser la room en Lobby tout en conservant l'historique et les scores cumulés
	_ = m.roomRepo.UpdateRoomStatus(ctx, client.RoomCode(), domain.RoomStatusInLobby)
	_ = m.roomRepo.UpdateRound(ctx, client.RoomCode(), 1, nil)

	m.mu.Lock()
	delete(m.activeRounds, client.RoomCode())
	m.mu.Unlock()

	// Réintégrer tous les spectateurs au statut de joueur actif pour la prochaine partie
	players, _ := m.roomRepo.GetPlayers(ctx, client.RoomCode())
	for _, p := range players {
		if p.Role == domain.RoleSpectator {
			p.Role = domain.RolePlayer
			_ = m.roomRepo.AddPlayer(ctx, client.RoomCode(), &p)
		}
	}

	m.hub.BroadcastSystemMessage(client.RoomCode(), "🔄 Le Master a lancé une revanche ! Vous êtes de retour dans le lobby.")
	m.hub.SyncRoom(client.RoomCode())
}

func (m *WordleGameManager) syncGameStateForUser(client *ws.Client) {
	m.mu.RLock()
	round, ok := m.activeRounds[client.RoomCode()]
	m.mu.RUnlock()

	if !ok || round.IsCompleted {
		return
	}

	// Préparer la vue masquée pour chaque joueur
	type OpponentView struct {
		UserID          uuid.UUID                 `json:"user_id"`
		DisplayUsername string                    `json:"display_username"`
		MaskedRows      [][]MaskedTileEvaluation  `json:"masked_rows"`
		IsSolved        bool                      `json:"is_solved"`
		IsFinished      bool                      `json:"is_finished"`
	}

	opponents := make([]OpponentView, 0)
	var myAttempts [][]TileEvaluation

	m.mu.RLock()
	for uid, pState := range round.Players {
		if uid == client.UserID() {
			myAttempts = pState.Attempts
		} else {
			p, _ := m.roomRepo.GetPlayer(context.Background(), client.RoomCode(), uid)
			name := "Joueur"
			if p != nil {
				name = p.DisplayUsername
			}
			opponents = append(opponents, OpponentView{
				UserID:          uid,
				DisplayUsername: name,
				MaskedRows:      pState.MaskedRows,
				IsSolved:        pState.IsSolved,
				IsFinished:      pState.IsFinished,
			})
		}
	}
	m.mu.RUnlock()

	syncPayload, _ := json.Marshal(map[string]interface{}{
		"word_length":  round.WordLength,
		"max_attempts": round.MaxAttempts,
		"ends_at":      round.EndsAt.Format(time.RFC3339),
		"my_attempts":  myAttempts,
		"opponents":    opponents,
	})

	client.Send(domain.WSMessage{
		Type:    "game:sync_state",
		Payload: syncPayload,
	})
}
