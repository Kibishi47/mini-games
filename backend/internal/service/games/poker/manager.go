package poker

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"

	"minigames-backend/internal/domain"
	redisRepo "minigames-backend/internal/repository/redis"
	"minigames-backend/internal/transport/ws"
)

const (
	actionTimeout     = 20 * time.Second
	nextHandDelay     = 6 * time.Second
	streetRevealDelay = 900 * time.Millisecond
)

type SeatStatus string

const (
	SeatStatusActive SeatStatus = "active"
	SeatStatusFolded SeatStatus = "folded"
	SeatStatusAllIn  SeatStatus = "all_in"
)

// Seat représente la position d'un joueur à la table pour la main en cours
type Seat struct {
	UserID              uuid.UUID
	Nickname            string
	Mascot              string
	Color               string
	Chips               int
	HoleCards           []Card
	Status              SeatStatus
	CommittedThisStreet int
	TotalCommitted      int
	HasActed            bool
	IsDealer            bool
	IsSB                bool
	IsBB                bool
}

// HandState représente l'état d'une main de Texas Hold'em en cours
type HandState struct {
	HandNum          int
	Deck             *Deck
	Community        []Card
	Seats            []*Seat
	DealerIdx        int
	PostflopFirstIdx int // premier siège à parler sur flop/turn/river
	ToActIdx         int
	Street           string // "preflop" | "flop" | "turn" | "river" | "showdown"
	CurrentBet       int
	MinRaise         int
	ActionDeadline   time.Time
	IsCompleted      bool
	BettingClosed    bool // true dès que plus aucune action n'est possible (run-out all-in en cours)
	StartedAt        time.Time
}

// Table représente l'état persistant d'une salle de poker (tapis entre les mains)
type Table struct {
	RoomCode   string
	Chips      map[uuid.UUID]int // tapis persistant entre les mains
	Order      []uuid.UUID       // ordre d'arrivée stable des participants
	LastDealer *uuid.UUID
	SmallBlind int
	BigBlind   int
	HandCount  int
	Hand       *HandState
}

type PokerGameManager struct {
	mu       sync.RWMutex
	hub      *ws.Hub
	roomRepo *redisRepo.RoomRepository
	tables   map[string]*Table // roomCode -> Table

	timersMu       sync.Mutex
	actionTimers   map[string]*time.Timer
	nextHandTimers map[string]*time.Timer
}

func NewPokerGameManager(hub *ws.Hub, roomRepo *redisRepo.RoomRepository) *PokerGameManager {
	return &PokerGameManager{
		hub:            hub,
		roomRepo:       roomRepo,
		tables:         make(map[string]*Table),
		actionTimers:   make(map[string]*time.Timer),
		nextHandTimers: make(map[string]*time.Timer),
	}
}

// -----------------------------------------------------------------------------
// DISPATCH DES ACTIONS WEBSOCKET
// -----------------------------------------------------------------------------

func (m *PokerGameManager) HandleGameAction(client *ws.Client, action string, payload json.RawMessage) {
	ctx := context.Background()

	switch action {
	case "game:start":
		m.handleStartGame(ctx, client)
	case "game:poker_action":
		m.handlePlayerAction(ctx, client, payload)
	case "game:get_state":
		m.sendStateToClient(client)
	case "game:stop", "room:return_lobby", "room:rematch":
		m.handleStopOrReturnLobby(ctx, client)
	}
}

func (m *PokerGameManager) OnPlayerJoined(client *ws.Client, room *domain.Room) {
	m.mu.Lock()
	table, exists := m.tables[room.Code]
	if !exists {
		m.mu.Unlock()
		return
	}

	// Un nouvel arrivant en cours de partie est spectateur jusqu'à la prochaine main
	if _, seated := table.Chips[client.UserID()]; !seated {
		table.Chips[client.UserID()] = room.Settings.StartingChips
		table.Order = append(table.Order, client.UserID())
		_ = m.roomRepo.SetPlayerSpectator(context.Background(), room.Code, client.UserID(), true)
	}
	m.mu.Unlock()

	go m.sendStateToClient(client)
}

func (m *PokerGameManager) OnPlayerLeft(roomCode string, userID uuid.UUID) {
	// Période de grâce de reconnexion gérée par le Hub ; le timer d'action gère l'absence de coup
}

// -----------------------------------------------------------------------------
// DÉMARRAGE DE PARTIE & DES MAINS
// -----------------------------------------------------------------------------

func (m *PokerGameManager) handleStartGame(ctx context.Context, client *ws.Client) {
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

	players, _ := m.roomRepo.GetPlayers(ctx, room.Code)
	eligible := make([]domain.RoomPlayer, 0, len(players))
	for _, p := range players {
		if p.Role != domain.RoleSpectator {
			eligible = append(eligible, p)
		}
	}
	if len(eligible) < 2 {
		client.SendError("Il faut au moins 2 joueurs pour démarrer une partie de poker")
		return
	}

	chips := make(map[uuid.UUID]int, len(eligible))
	order := make([]uuid.UUID, 0, len(eligible))
	for _, p := range eligible {
		chips[p.ID] = room.Settings.StartingChips
		order = append(order, p.ID)
		_ = m.roomRepo.SetPlayerScore(ctx, room.Code, p.ID, room.Settings.StartingChips)
	}

	table := &Table{
		RoomCode:   room.Code,
		Chips:      chips,
		Order:      order,
		SmallBlind: room.Settings.SmallBlind,
		BigBlind:   room.Settings.BigBlind,
	}

	m.mu.Lock()
	m.tables[room.Code] = table
	m.mu.Unlock()

	_ = m.roomRepo.UpdateRoomStatus(ctx, room.Code, domain.RoomStatusInGame)
	_ = m.roomRepo.SetAllPlayersLocation(ctx, room.Code, "in_game")

	m.startHand(ctx, room.Code, table)
}

// startHand initialise une nouvelle main : rotation du bouton, cartes, blindes
func (m *PokerGameManager) startHand(ctx context.Context, roomCode string, table *Table) {
	m.mu.Lock()

	players, _ := m.roomRepo.GetPlayers(ctx, roomCode)
	playerByID := make(map[uuid.UUID]domain.RoomPlayer, len(players))
	for _, p := range players {
		playerByID[p.ID] = p
	}

	// Sièges éligibles pour cette main : jetons > 0, joueur toujours présent et non spectateur
	seats := make([]*Seat, 0, len(table.Order))
	for _, uid := range table.Order {
		p, present := playerByID[uid]
		if !present || p.Role == domain.RoleSpectator {
			continue
		}
		if chips, ok := table.Chips[uid]; !ok || chips <= 0 {
			continue
		}
		seats = append(seats, &Seat{
			UserID:   uid,
			Nickname: p.Nickname,
			Mascot:   p.Mascot,
			Color:    p.Color,
			Chips:    table.Chips[uid],
			Status:   SeatStatusActive,
		})
	}

	if len(seats) < 2 {
		m.mu.Unlock()
		m.endGame(ctx, roomCode, "Plus assez de joueurs avec des jetons pour continuer la partie.")
		return
	}

	table.HandCount++

	dealerIdx := 0
	if table.LastDealer != nil {
		for i, s := range seats {
			if s.UserID == *table.LastDealer {
				dealerIdx = (i + 1) % len(seats)
				break
			}
		}
	}

	sbIdx, bbIdx := dealerIdx, (dealerIdx+1)%len(seats)
	if len(seats) > 2 {
		sbIdx = (dealerIdx + 1) % len(seats)
		bbIdx = (dealerIdx + 2) % len(seats)
	}

	seats[dealerIdx].IsDealer = true
	seats[sbIdx].IsSB = true
	seats[bbIdx].IsBB = true

	deck := NewShuffledDeck()
	for i := 0; i < 2; i++ {
		for _, s := range seats {
			card, _ := deck.Draw()
			s.HoleCards = append(s.HoleCards, card)
		}
	}

	hand := &HandState{
		HandNum:   table.HandCount,
		Deck:      deck,
		Community: make([]Card, 0, 5),
		Seats:     seats,
		DealerIdx: dealerIdx,
		Street:    "preflop",
		MinRaise:  table.BigBlind,
		StartedAt: time.Now(),
	}

	postBlind(seats[sbIdx], table.SmallBlind)
	postBlind(seats[bbIdx], table.BigBlind)
	hand.CurrentBet = maxInt(seats[sbIdx].CommittedThisStreet, seats[bbIdx].CommittedThisStreet)

	if len(seats) == 2 {
		hand.PostflopFirstIdx = bbIdx
		hand.ToActIdx = nextActiveIdx(seats, dealerIdx-1)
	} else {
		hand.PostflopFirstIdx = nextActiveIdx(seats, dealerIdx)
		hand.ToActIdx = nextActiveIdx(seats, bbIdx)
	}

	dealerUID := seats[dealerIdx].UserID
	table.LastDealer = &dealerUID
	table.Hand = hand
	m.mu.Unlock()

	room, err := m.roomRepo.GetRoom(ctx, roomCode)
	if err != nil {
		return
	}
	_ = m.roomRepo.UpdateRound(ctx, roomCode, table.HandCount, nil)
	_ = m.roomRepo.SetRoundState(ctx, roomCode, domain.RoundSubStatePlaying, "", nil)

	m.hub.BroadcastSystemMessage(roomCode, fmt.Sprintf("Nouvelle main #%d — Petite Blinde %d / Grosse Blinde %d", table.HandCount, table.SmallBlind, table.BigBlind))
	m.hub.SyncRoom(roomCode)
	m.broadcastState(roomCode)
	m.armActionTimer(roomCode)
	_ = room
}

func postBlind(seat *Seat, amount int) {
	pay := amount
	if pay > seat.Chips {
		pay = seat.Chips
	}
	seat.Chips -= pay
	seat.CommittedThisStreet += pay
	seat.TotalCommitted += pay
	if seat.Chips == 0 {
		seat.Status = SeatStatusAllIn
	}
}

// -----------------------------------------------------------------------------
// TRAITEMENT DES ACTIONS D'UN JOUEUR (fold / check / call / bet / raise / all_in)
// -----------------------------------------------------------------------------

type actionBody struct {
	Action string `json:"action"`
	Amount int    `json:"amount"`
}

func (m *PokerGameManager) handlePlayerAction(ctx context.Context, client *ws.Client, payload json.RawMessage) {
	var body actionBody
	if err := json.Unmarshal(payload, &body); err != nil {
		client.SendError("Format d'action invalide")
		return
	}

	m.mu.Lock()
	table, ok := m.tables[client.RoomCode()]
	if !ok || table.Hand == nil || table.Hand.IsCompleted {
		m.mu.Unlock()
		client.SendError("Aucune main en cours dans cette salle")
		return
	}
	hand := table.Hand

	if hand.BettingClosed {
		m.mu.Unlock()
		client.SendError("Les mises sont closes pour cette main, en attente de l'abattage")
		return
	}

	if hand.Seats[hand.ToActIdx].UserID != client.UserID() {
		m.mu.Unlock()
		client.SendError("Ce n'est pas votre tour de jouer")
		return
	}

	seatIdx := hand.ToActIdx
	if err := m.applyAction(table, hand, seatIdx, body.Action, body.Amount); err != nil {
		m.mu.Unlock()
		client.SendError(err.Error())
		return
	}
	m.mu.Unlock()

	m.cancelActionTimer(client.RoomCode())
	m.afterAction(ctx, client.RoomCode(), table, hand)
}

// applyAction valide et applique une action sur le siège donné (verrou déjà détenu par l'appelant)
func (m *PokerGameManager) applyAction(table *Table, hand *HandState, seatIdx int, action string, amount int) error {
	seat := hand.Seats[seatIdx]
	maxTarget := seat.CommittedThisStreet + seat.Chips

	switch action {
	case "fold":
		seat.Status = SeatStatusFolded
		seat.HasActed = true
		return nil

	case "check":
		if seat.CommittedThisStreet != hand.CurrentBet {
			return fmt.Errorf("impossible de checker, il faut suivre %d jetons", hand.CurrentBet-seat.CommittedThisStreet)
		}
		seat.HasActed = true
		return nil

	case "call":
		target := hand.CurrentBet
		if target > maxTarget {
			target = maxTarget
		}
		if target <= seat.CommittedThisStreet {
			return fmt.Errorf("aucune mise à suivre, checkez plutôt")
		}
		commitSeat(seat, target)
		return nil

	case "all_in":
		target := maxTarget
		if target <= seat.CommittedThisStreet {
			return fmt.Errorf("aucun jeton disponible pour ce tapis")
		}
		applyRaise(hand, seat, target, table.BigBlind)
		return nil

	case "bet", "raise":
		target := amount
		if target >= maxTarget {
			target = maxTarget
		} else {
			minTarget := hand.CurrentBet + hand.MinRaise
			if hand.CurrentBet == 0 {
				minTarget = table.BigBlind
			}
			if target < minTarget {
				return fmt.Errorf("la mise minimale est de %d jetons", minTarget)
			}
		}
		if target <= hand.CurrentBet {
			return fmt.Errorf("le montant doit être supérieur à la mise actuelle de %d jetons", hand.CurrentBet)
		}
		applyRaise(hand, seat, target, table.BigBlind)
		return nil

	default:
		return fmt.Errorf("action inconnue: %s", action)
	}
}

func commitSeat(seat *Seat, target int) {
	delta := target - seat.CommittedThisStreet
	if delta < 0 {
		delta = 0
	}
	if delta > seat.Chips {
		delta = seat.Chips
	}
	seat.Chips -= delta
	seat.CommittedThisStreet += delta
	seat.TotalCommitted += delta
	if seat.Chips == 0 {
		seat.Status = SeatStatusAllIn
	}
	seat.HasActed = true
}

func applyRaise(hand *HandState, seat *Seat, target int, bigBlind int) {
	raiseIncrement := target - hand.CurrentBet
	commitSeat(seat, target)
	hand.CurrentBet = target
	if raiseIncrement >= hand.MinRaise || hand.MinRaise == 0 {
		hand.MinRaise = raiseIncrement
	}
	if hand.MinRaise <= 0 {
		hand.MinRaise = bigBlind
	}
	// La relance rouvre l'action pour tous les autres sièges encore actifs
	for _, s := range hand.Seats {
		if s != seat && s.Status == SeatStatusActive {
			s.HasActed = false
		}
	}
}

// -----------------------------------------------------------------------------
// PROGRESSION DE LA MAIN APRÈS UNE ACTION
// -----------------------------------------------------------------------------

func (m *PokerGameManager) afterAction(ctx context.Context, roomCode string, table *Table, hand *HandState) {
	m.mu.Lock()

	stillIn := 0
	for _, s := range hand.Seats {
		if s.Status != SeatStatusFolded {
			stillIn++
		}
	}

	if stillIn <= 1 {
		m.mu.Unlock()
		m.concludeUncontested(ctx, roomCode, table, hand)
		return
	}

	if bettingRoundComplete(hand) {
		if !bettingCanContinue(hand) || hand.Street == "river" {
			hand.BettingClosed = true
			m.mu.Unlock()
			m.runOutAndShowdown(ctx, roomCode, table, hand)
			return
		}
		m.advanceStreet(hand)
		m.mu.Unlock()
		m.hub.SyncRoom(roomCode)
		m.broadcastState(roomCode)
		m.armActionTimer(roomCode)
		return
	}

	hand.ToActIdx = nextActiveIdx(hand.Seats, hand.ToActIdx)
	m.mu.Unlock()
	m.broadcastState(roomCode)
	m.armActionTimer(roomCode)
}

// bettingRoundComplete vérifie si tous les sièges actifs ont parlé au même niveau de mise
func bettingRoundComplete(hand *HandState) bool {
	for _, s := range hand.Seats {
		if s.Status == SeatStatusActive && (!s.HasActed || s.CommittedThisStreet != hand.CurrentBet) {
			return false
		}
	}
	return true
}

// bettingCanContinue indique s'il reste au moins 2 sièges capables de miser encore
func bettingCanContinue(hand *HandState) bool {
	active := 0
	for _, s := range hand.Seats {
		if s.Status == SeatStatusActive {
			active++
		}
	}
	return active > 1
}

func (m *PokerGameManager) advanceStreet(hand *HandState) {
	for _, s := range hand.Seats {
		s.CommittedThisStreet = 0
		if s.Status == SeatStatusActive {
			s.HasActed = false
		}
	}
	hand.CurrentBet = 0
	hand.MinRaise = 0

	switch hand.Street {
	case "preflop":
		hand.Street = "flop"
		_, _ = hand.Deck.Draw() // brûlage
		for i := 0; i < 3; i++ {
			c, _ := hand.Deck.Draw()
			hand.Community = append(hand.Community, c)
		}
	case "flop":
		hand.Street = "turn"
		_, _ = hand.Deck.Draw()
		c, _ := hand.Deck.Draw()
		hand.Community = append(hand.Community, c)
	case "turn":
		hand.Street = "river"
		_, _ = hand.Deck.Draw()
		c, _ := hand.Deck.Draw()
		hand.Community = append(hand.Community, c)
	}

	hand.ToActIdx = nextActiveIdx(hand.Seats, hand.PostflopFirstIdx-1)
}

// runOutAndShowdown termine de dévoiler le tableau (si nécessaire) puis va à l'abattage
func (m *PokerGameManager) runOutAndShowdown(ctx context.Context, roomCode string, table *Table, hand *HandState) {
	go func() {
		m.mu.Lock()
		for hand.Street != "river" {
			m.advanceStreet(hand)
			m.mu.Unlock()
			m.hub.SyncRoom(roomCode)
			m.broadcastState(roomCode)
			time.Sleep(streetRevealDelay)
			m.mu.Lock()
		}
		m.mu.Unlock()
		m.showdown(ctx, roomCode, table, hand)
	}()
}

// -----------------------------------------------------------------------------
// TIMER D'ACTION (fold/check automatique en cas d'inactivité)
// -----------------------------------------------------------------------------

func (m *PokerGameManager) armActionTimer(roomCode string) {
	m.timersMu.Lock()
	if t, exists := m.actionTimers[roomCode]; exists {
		t.Stop()
	}

	m.mu.Lock()
	table, ok := m.tables[roomCode]
	if !ok || table.Hand == nil {
		m.mu.Unlock()
		m.timersMu.Unlock()
		return
	}
	table.Hand.ActionDeadline = time.Now().Add(actionTimeout)
	m.mu.Unlock()

	m.actionTimers[roomCode] = time.AfterFunc(actionTimeout, func() {
		m.handleActionTimeout(roomCode)
	})
	m.timersMu.Unlock()
}

func (m *PokerGameManager) cancelActionTimer(roomCode string) {
	m.timersMu.Lock()
	if t, exists := m.actionTimers[roomCode]; exists {
		t.Stop()
		delete(m.actionTimers, roomCode)
	}
	m.timersMu.Unlock()
}

func (m *PokerGameManager) handleActionTimeout(roomCode string) {
	ctx := context.Background()

	m.mu.Lock()
	table, ok := m.tables[roomCode]
	if !ok || table.Hand == nil || table.Hand.IsCompleted {
		m.mu.Unlock()
		return
	}
	hand := table.Hand
	seat := hand.Seats[hand.ToActIdx]

	autoAction := "fold"
	if seat.CommittedThisStreet == hand.CurrentBet {
		autoAction = "check"
	}
	_ = m.applyAction(table, hand, hand.ToActIdx, autoAction, 0)
	m.mu.Unlock()

	label := "s'est couché (temps écoulé)"
	if autoAction == "check" {
		label = "a checké (temps écoulé)"
	}
	m.hub.BroadcastSystemMessage(roomCode, fmt.Sprintf("%s %s", seat.Nickname, label))

	m.afterAction(ctx, roomCode, table, hand)
}

// -----------------------------------------------------------------------------
// ABATTAGE, RÉPARTITION DES POTS & FIN DE MAIN
// -----------------------------------------------------------------------------

type potTier struct {
	amount   int
	eligible []*Seat
}

// computePots calcule le pot principal et les pots secondaires (tapis courts / all-in)
func computePots(seats []*Seat) []potTier {
	contenders := make([]*Seat, 0, len(seats))
	for _, s := range seats {
		if s.Status != SeatStatusFolded {
			contenders = append(contenders, s)
		}
	}
	sort.Slice(contenders, func(i, j int) bool { return contenders[i].TotalCommitted < contenders[j].TotalCommitted })

	var tiers []potTier
	prevLevel := 0
	processed := map[int]bool{}

	for _, c := range contenders {
		level := c.TotalCommitted
		if level <= prevLevel || processed[level] {
			continue
		}
		processed[level] = true

		amount := 0
		for _, s := range seats {
			contribution := s.TotalCommitted - prevLevel
			if contribution <= 0 {
				continue
			}
			if s.TotalCommitted < level {
				amount += contribution
			} else {
				amount += level - prevLevel
			}
		}

		eligible := make([]*Seat, 0, len(contenders))
		for _, c2 := range contenders {
			if c2.TotalCommitted >= level {
				eligible = append(eligible, c2)
			}
		}

		tiers = append(tiers, potTier{amount: amount, eligible: eligible})
		prevLevel = level
	}

	return tiers
}

func (m *PokerGameManager) showdown(ctx context.Context, roomCode string, table *Table, hand *HandState) {
	m.mu.Lock()
	hand.IsCompleted = true
	hand.Street = "showdown"

	results := make(map[uuid.UUID]HandResult, len(hand.Seats))
	for _, s := range hand.Seats {
		if s.Status == SeatStatusFolded {
			continue
		}
		results[s.UserID] = Evaluate7(append(append([]Card{}, s.HoleCards...), hand.Community...))
	}

	tiers := computePots(hand.Seats)
	winnings := make(map[uuid.UUID]int)
	for _, tier := range tiers {
		if len(tier.eligible) == 0 || tier.amount == 0 {
			continue
		}
		best := results[tier.eligible[0].UserID].Score
		winners := []*Seat{tier.eligible[0]}
		for _, s := range tier.eligible[1:] {
			r := results[s.UserID].Score
			if r > best {
				best = r
				winners = []*Seat{s}
			} else if r == best {
				winners = append(winners, s)
			}
		}
		share := tier.amount / len(winners)
		remainder := tier.amount % len(winners)
		for i, w := range winners {
			amt := share
			if i == 0 {
				amt += remainder // le reste indivisible va au premier gagnant (position la plus proche du bouton)
			}
			winnings[w.UserID] += amt
		}
	}

	revealed := make([]map[string]interface{}, 0, len(hand.Seats))
	for _, s := range hand.Seats {
		s.Chips += winnings[s.UserID]
		table.Chips[s.UserID] = s.Chips
		entry := map[string]interface{}{
			"user_id":  s.UserID,
			"nickname": s.Nickname,
			"folded":   s.Status == SeatStatusFolded,
			"winnings": winnings[s.UserID],
		}
		if s.Status != SeatStatusFolded {
			entry["hole_cards"] = s.HoleCards
			entry["category"] = CategoryName(results[s.UserID].Category)
		}
		revealed = append(revealed, entry)
	}
	m.mu.Unlock()

	for _, s := range hand.Seats {
		_ = m.roomRepo.SetPlayerScore(ctx, roomCode, s.UserID, s.Chips)
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"hand_num":  hand.HandNum,
		"community": hand.Community,
		"players":   revealed,
	})
	m.hub.BroadcastToRoom(roomCode, domain.WSMessage{Type: "poker:hand_end", Payload: payload})

	m.finishHand(ctx, roomCode, table, hand)
}

// concludeUncontested attribue le pot au seul joueur restant lorsque tous les autres se sont couchés
func (m *PokerGameManager) concludeUncontested(ctx context.Context, roomCode string, table *Table, hand *HandState) {
	m.mu.Lock()
	hand.IsCompleted = true

	var winner *Seat
	total := 0
	for _, s := range hand.Seats {
		total += s.TotalCommitted
		if s.Status != SeatStatusFolded {
			winner = s
		}
	}
	if winner != nil {
		winner.Chips += total
		table.Chips[winner.UserID] = winner.Chips
	}
	m.mu.Unlock()

	if winner != nil {
		_ = m.roomRepo.SetPlayerScore(ctx, roomCode, winner.UserID, winner.Chips)
		payload, _ := json.Marshal(map[string]interface{}{
			"hand_num":    hand.HandNum,
			"community":   hand.Community,
			"uncontested": true,
			"winner_id":   winner.UserID,
			"winner_name": winner.Nickname,
			"amount":      total,
		})
		m.hub.BroadcastToRoom(roomCode, domain.WSMessage{Type: "poker:hand_end", Payload: payload})
	}

	m.finishHand(ctx, roomCode, table, hand)
}

// finishHand gère l'élimination des joueurs ruinés et programme la main suivante ou la fin de partie
func (m *PokerGameManager) finishHand(ctx context.Context, roomCode string, table *Table, hand *HandState) {
	m.cancelActionTimer(roomCode)

	m.mu.Lock()
	solvent := 0
	for _, s := range hand.Seats {
		if s.Chips <= 0 {
			go m.eliminatePlayer(roomCode, s.UserID, s.Nickname)
		} else {
			solvent++
		}
	}
	m.mu.Unlock()

	m.hub.SyncRoom(roomCode)
	m.broadcastState(roomCode)

	if solvent < 2 {
		m.endGame(ctx, roomCode, "Partie terminée : un seul joueur possède encore des jetons !")
		return
	}

	m.timersMu.Lock()
	if t, exists := m.nextHandTimers[roomCode]; exists {
		t.Stop()
	}
	m.nextHandTimers[roomCode] = time.AfterFunc(nextHandDelay, func() {
		m.timersMu.Lock()
		delete(m.nextHandTimers, roomCode)
		m.timersMu.Unlock()

		m.mu.RLock()
		t, ok := m.tables[roomCode]
		m.mu.RUnlock()
		if ok {
			m.startHand(context.Background(), roomCode, t)
		}
	})
	m.timersMu.Unlock()
}

func (m *PokerGameManager) eliminatePlayer(roomCode string, userID uuid.UUID, nickname string) {
	ctx := context.Background()
	_ = m.roomRepo.SetPlayerSpectator(ctx, roomCode, userID, true)
	m.hub.BroadcastSystemMessage(roomCode, fmt.Sprintf("%s a fait tapis et perdu — élimination de la partie de poker !", nickname))
}

// endGame termine la partie faute de joueurs solvables et laisse la salle en état "game_over"
func (m *PokerGameManager) endGame(ctx context.Context, roomCode string, reason string) {
	m.mu.Lock()
	delete(m.tables, roomCode)
	m.mu.Unlock()

	_ = m.roomRepo.SetRoundState(ctx, roomCode, domain.RoundSubStateGameOver, "", nil)
	m.hub.BroadcastSystemMessage(roomCode, reason)
	m.hub.SyncRoom(roomCode)
}

// -----------------------------------------------------------------------------
// RETOUR AU LOBBY / ARRÊT DE PARTIE
// -----------------------------------------------------------------------------

func (m *PokerGameManager) handleStopOrReturnLobby(ctx context.Context, client *ws.Client) {
	room, err := m.roomRepo.GetRoom(ctx, client.RoomCode())
	if err != nil {
		client.SendError("Salle introuvable")
		return
	}

	isGameOver := room.RoundState == domain.RoundSubStateGameOver
	if room.MasterID != client.UserID() && !isGameOver {
		client.SendError("Action réservée au Master")
		return
	}

	roomCode := client.RoomCode()
	m.cancelActionTimer(roomCode)
	m.timersMu.Lock()
	if t, exists := m.nextHandTimers[roomCode]; exists {
		t.Stop()
		delete(m.nextHandTimers, roomCode)
	}
	m.timersMu.Unlock()

	m.mu.Lock()
	delete(m.tables, roomCode)
	m.mu.Unlock()

	_ = m.roomRepo.UpdateRoomStatus(ctx, roomCode, domain.RoomStatusInLobby)
	_ = m.roomRepo.SetRoundState(ctx, roomCode, domain.RoundSubStateIdle, "", nil)
	_ = m.roomRepo.UpdateRound(ctx, roomCode, 0, nil)
	_ = m.roomRepo.SetAllPlayersLocation(ctx, roomCode, "lobby")

	players, _ := m.roomRepo.GetPlayers(ctx, roomCode)
	for _, p := range players {
		if p.IsSpectator {
			_ = m.roomRepo.SetPlayerSpectator(ctx, roomCode, p.ID, false)
		}
	}

	statePayload, _ := json.Marshal(map[string]interface{}{"status": domain.RoomStatusInLobby})
	m.hub.BroadcastToRoom(roomCode, domain.WSMessage{Type: "room:state_changed", Payload: statePayload})
	m.hub.SyncRoom(roomCode)
}

// -----------------------------------------------------------------------------
// DIFFUSION D'ÉTAT (MASQUAGE DES CARTES ADVERSAIRES)
// -----------------------------------------------------------------------------

func (m *PokerGameManager) buildStatePayload(table *Table, hand *HandState, forUserID uuid.UUID) map[string]interface{} {
	revealAll := hand.Street == "showdown"
	pot := 0
	seatsView := make([]map[string]interface{}, 0, len(hand.Seats))

	for i, s := range hand.Seats {
		pot += s.TotalCommitted
		view := map[string]interface{}{
			"user_id":   s.UserID,
			"nickname":  s.Nickname,
			"mascot":    s.Mascot,
			"color":     s.Color,
			"chips":     s.Chips,
			"bet":       s.CommittedThisStreet,
			"status":    s.Status,
			"is_dealer": s.IsDealer,
			"is_sb":     s.IsSB,
			"is_bb":     s.IsBB,
			"is_turn":   i == hand.ToActIdx && !hand.IsCompleted,
		}
		if s.UserID == forUserID || revealAll {
			view["hole_cards"] = s.HoleCards
		}
		seatsView = append(seatsView, view)
	}

	toActUserID := uuid.Nil
	callAmount := 0
	if !hand.IsCompleted {
		toActSeat := hand.Seats[hand.ToActIdx]
		toActUserID = toActSeat.UserID
		callAmount = hand.CurrentBet - toActSeat.CommittedThisStreet
	}

	return map[string]interface{}{
		"hand_num":       hand.HandNum,
		"street":         hand.Street,
		"community":      hand.Community,
		"pot":            pot,
		"current_bet":    hand.CurrentBet,
		"min_raise":      hand.MinRaise,
		"call_amount":    callAmount,
		"to_act":         toActUserID,
		"action_ends_at": hand.ActionDeadline.Format(time.RFC3339),
		"small_blind":    table.SmallBlind,
		"big_blind":      table.BigBlind,
		"is_completed":   hand.IsCompleted,
		"seats":          seatsView,
	}
}

func (m *PokerGameManager) broadcastState(roomCode string) {
	m.mu.RLock()
	table, ok := m.tables[roomCode]
	if !ok || table.Hand == nil {
		m.mu.RUnlock()
		return
	}
	hand := table.Hand
	clients := m.hub.GetRoomClients(roomCode)
	payloads := make(map[uuid.UUID][]byte, len(clients))
	for _, c := range clients {
		p := m.buildStatePayload(table, hand, c.UserID())
		b, _ := json.Marshal(p)
		payloads[c.UserID()] = b
	}
	m.mu.RUnlock()

	for _, c := range clients {
		if b, exists := payloads[c.UserID()]; exists {
			c.Send(domain.WSMessage{Type: "poker:state", Payload: b})
		}
	}
}

func (m *PokerGameManager) sendStateToClient(client *ws.Client) {
	m.mu.RLock()
	table, ok := m.tables[client.RoomCode()]
	if !ok || table.Hand == nil {
		m.mu.RUnlock()
		return
	}
	payload := m.buildStatePayload(table, table.Hand, client.UserID())
	m.mu.RUnlock()

	b, _ := json.Marshal(payload)
	client.Send(domain.WSMessage{Type: "poker:state", Payload: b})
}

// -----------------------------------------------------------------------------
// HELPERS
// -----------------------------------------------------------------------------

// nextActiveIdx trouve le prochain siège en statut "active" en partant de (from+1), en cercle
func nextActiveIdx(seats []*Seat, from int) int {
	n := len(seats)
	for i := 1; i <= n; i++ {
		idx := ((from+i)%n + n) % n
		if seats[idx].Status == SeatStatusActive {
			return idx
		}
	}
	return (from + 1 + n) % n
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
