package poker

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
)

// Rang des cartes : 2 à 14 (14 = As, valeur haute par défaut)
const (
	Two   = 2
	Three = 3
	Four  = 4
	Five  = 5
	Six   = 6
	Seven = 7
	Eight = 8
	Nine  = 9
	Ten   = 10
	Jack  = 11
	Queen = 12
	King  = 13
	Ace   = 14
)

// Couleurs
const (
	Clubs    = 0 // Trèfle
	Diamonds = 1 // Carreau
	Hearts   = 2 // Cœur
	Spades   = 3 // Pique
)

// Card représente une carte à jouer (rang 2-14, couleur 0-3)
type Card struct {
	Rank int
	Suit int
}

var rankSymbols = map[int]string{
	Two: "2", Three: "3", Four: "4", Five: "5", Six: "6", Seven: "7",
	Eight: "8", Nine: "9", Ten: "T", Jack: "J", Queen: "Q", King: "K", Ace: "A",
}

var suitSymbols = map[int]string{
	Clubs: "C", Diamonds: "D", Hearts: "H", Spades: "S",
}

// String encode la carte en notation compacte (ex: "AS" = As de Pique, "TD" = 10 de Carreau)
func (c Card) String() string {
	return rankSymbols[c.Rank] + suitSymbols[c.Suit]
}

// MarshalJSON diffuse la carte au format string compact pour le frontend
func (c Card) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

// UnmarshalJSON pour désérialisation (utilisé dans les tests)
func (c *Card) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if len(s) != 2 {
		return fmt.Errorf("format de carte invalide: %s", s)
	}
	for rank, sym := range rankSymbols {
		if sym == string(s[0]) {
			c.Rank = rank
			break
		}
	}
	for suit, sym := range suitSymbols {
		if sym == string(s[1]) {
			c.Suit = suit
			break
		}
	}
	return nil
}

// Deck représente un paquet de 52 cartes mélangé, tiré depuis le haut
type Deck struct {
	cards []Card
	pos   int
}

// NewShuffledDeck crée un paquet de 52 cartes complet mélangé cryptographiquement
func NewShuffledDeck() *Deck {
	cards := make([]Card, 0, 52)
	for suit := 0; suit < 4; suit++ {
		for rank := Two; rank <= Ace; rank++ {
			cards = append(cards, Card{Rank: rank, Suit: suit})
		}
	}

	// Mélange Fisher-Yates avec source aléatoire cryptographique (crypto/rand)
	for i := len(cards) - 1; i > 0; i-- {
		jBig, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		j := int(jBig.Int64())
		cards[i], cards[j] = cards[j], cards[i]
	}

	return &Deck{cards: cards, pos: 0}
}

// Draw tire la carte suivante du dessus du paquet
func (d *Deck) Draw() (Card, error) {
	if d.pos >= len(d.cards) {
		return Card{}, fmt.Errorf("paquet épuisé")
	}
	c := d.cards[d.pos]
	d.pos++
	return c, nil
}

// Remaining renvoie le nombre de cartes restantes à tirer
func (d *Deck) Remaining() int {
	return len(d.cards) - d.pos
}
