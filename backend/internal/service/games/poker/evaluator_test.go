package poker

import (
	"testing"

	"github.com/google/uuid"
)

func c(spec string) Card {
	var card Card
	rankChar := spec[0:1]
	suitChar := spec[1:2]
	for rank, sym := range rankSymbols {
		if sym == rankChar {
			card.Rank = rank
		}
	}
	for suit, sym := range suitSymbols {
		if sym == suitChar {
			card.Suit = suit
		}
	}
	return card
}

func cards(specs ...string) []Card {
	out := make([]Card, 0, len(specs))
	for _, s := range specs {
		out = append(out, c(s))
	}
	return out
}

func TestEvaluate7_Categories(t *testing.T) {
	tests := []struct {
		name     string
		hand     []Card
		category int
	}{
		{"Quinte Flush", cards("9H", "TH", "JH", "QH", "KH", "2C", "3D"), CategoryStraightFlush},
		{"Carré", cards("9H", "9D", "9C", "9S", "KH", "2C", "3D"), CategoryFourOfAKind},
		{"Full", cards("9H", "9D", "9C", "KH", "KD", "2C", "3D"), CategoryFullHouse},
		{"Couleur", cards("2H", "5H", "9H", "JH", "KH", "2C", "3D"), CategoryFlush},
		{"Suite", cards("5H", "6D", "7C", "8S", "9H", "2C", "3D"), CategoryStraight},
		{"Suite Basse (Roue)", cards("AH", "2D", "3C", "4S", "5H", "9C", "KD"), CategoryStraight},
		{"Brelan", cards("9H", "9D", "9C", "KH", "2D", "3C", "4S"), CategoryThreeOfAKind},
		{"Double Paire", cards("9H", "9D", "KC", "KS", "2D", "3C", "4S"), CategoryTwoPair},
		{"Paire", cards("9H", "9D", "KC", "2S", "5D", "3C", "7S"), CategoryPair},
		{"Carte Haute", cards("2H", "5D", "9C", "JS", "KD", "3C", "7S"), CategoryHighCard},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := Evaluate7(tc.hand)
			if result.Category != tc.category {
				t.Errorf("%s: catégorie attendue=%d (%s), obtenue=%d (%s)", tc.name, tc.category, CategoryName(tc.category), result.Category, CategoryName(result.Category))
			}
		})
	}
}

func TestEvaluate7_HigherHandWins(t *testing.T) {
	fullHouse := Evaluate7(cards("9H", "9D", "9C", "KH", "KD", "2C", "3D"))
	flush := Evaluate7(cards("2H", "5H", "9H", "JH", "KH", "2C", "3D"))

	if fullHouse.Score <= flush.Score {
		t.Errorf("un Full doit toujours battre une Couleur (full=%d, flush=%d)", fullHouse.Score, flush.Score)
	}
}

func TestEvaluate7_KickerBreaksTie(t *testing.T) {
	pairAceKingKicker := Evaluate7(cards("AH", "AD", "KC", "9S", "5D", "2C", "3S"))
	pairAceQueenKicker := Evaluate7(cards("AH", "AD", "QC", "9S", "5D", "2C", "3S"))

	if pairAceKingKicker.Score <= pairAceQueenKicker.Score {
		t.Errorf("une paire d'As avec kicker Roi doit battre une paire d'As avec kicker Dame")
	}
}

func TestDeck_UniqueAndComplete(t *testing.T) {
	deck := NewShuffledDeck()
	seen := make(map[string]bool)

	for deck.Remaining() > 0 {
		card, err := deck.Draw()
		if err != nil {
			t.Fatalf("tirage inattendu en échec: %v", err)
		}
		key := card.String()
		if seen[key] {
			t.Fatalf("carte dupliquée détectée dans le paquet: %s", key)
		}
		seen[key] = true
	}

	if len(seen) != 52 {
		t.Fatalf("le paquet doit contenir exactement 52 cartes uniques, obtenu %d", len(seen))
	}

	if _, err := deck.Draw(); err == nil {
		t.Fatalf("tirer une carte au-delà de 52 devrait échouer")
	}
}

func TestComputePots_SidePotForShortAllIn(t *testing.T) {
	// A tapis à 100, B et C suivent à 300 chacun : pot principal 300, side pot 400
	seatA := &Seat{UserID: uuid.New(), Status: SeatStatusAllIn, TotalCommitted: 100}
	seatB := &Seat{UserID: uuid.New(), Status: SeatStatusActive, TotalCommitted: 300}
	seatC := &Seat{UserID: uuid.New(), Status: SeatStatusActive, TotalCommitted: 300}

	tiers := computePots([]*Seat{seatA, seatB, seatC})

	if len(tiers) != 2 {
		t.Fatalf("attendu 2 paliers de pot (principal + side pot), obtenu %d", len(tiers))
	}
	if tiers[0].amount != 300 {
		t.Errorf("pot principal attendu=300, obtenu=%d", tiers[0].amount)
	}
	if len(tiers[0].eligible) != 3 {
		t.Errorf("le pot principal doit être disputable par les 3 joueurs, obtenu %d", len(tiers[0].eligible))
	}
	if tiers[1].amount != 400 {
		t.Errorf("side pot attendu=400, obtenu=%d", tiers[1].amount)
	}
	if len(tiers[1].eligible) != 2 {
		t.Errorf("le side pot ne doit être disputable que par B et C, obtenu %d", len(tiers[1].eligible))
	}
}
