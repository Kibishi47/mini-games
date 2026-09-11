package poker

import "sort"

// Catégories de mains (valeur croissante = main plus forte)
const (
	CategoryHighCard      = 0
	CategoryPair          = 1
	CategoryTwoPair       = 2
	CategoryThreeOfAKind  = 3
	CategoryStraight      = 4
	CategoryFlush         = 5
	CategoryFullHouse     = 6
	CategoryFourOfAKind   = 7
	CategoryStraightFlush = 8
)

var categoryNames = map[int]string{
	CategoryHighCard:      "Carte Haute",
	CategoryPair:          "Paire",
	CategoryTwoPair:       "Double Paire",
	CategoryThreeOfAKind:  "Brelan",
	CategoryStraight:      "Suite",
	CategoryFlush:         "Couleur",
	CategoryFullHouse:     "Full",
	CategoryFourOfAKind:   "Carré",
	CategoryStraightFlush: "Quinte Flush",
}

// rc associe un rang à son nombre d'occurrences dans une main (utilisé pour paires/brelans/carrés)
type rc struct{ rank, count int }

// CategoryName renvoie le libellé français d'une catégorie de main
func CategoryName(category int) string {
	if name, ok := categoryNames[category]; ok {
		return name
	}
	return "Main Inconnue"
}

// HandResult contient le score comparable et le détail de la meilleure main à 5 cartes
type HandResult struct {
	Score    int64  // score total ordonnable (catégorie + kickers encodés)
	Category int    // catégorie 0-8 (pour l'affichage)
	Best5    []Card // meilleure combinaison de 5 cartes trouvée
}

// encodeScore combine une catégorie et jusqu'à 5 valeurs de départage (rangs 2-14) en un score total ordonnable
func encodeScore(category int, kickers ...int) int64 {
	score := int64(category)
	for i := 0; i < 5; i++ {
		score *= 15
		if i < len(kickers) {
			score += int64(kickers[i])
		}
	}
	return score
}

// Evaluate7 détermine la meilleure main possible parmi 5 à 7 cartes (2 cartes privées + jusqu'à 5 communes)
func Evaluate7(cards []Card) HandResult {
	best := HandResult{Score: -1}

	n := len(cards)
	if n < 5 {
		return best
	}

	// Énumération de toutes les combinaisons de 5 cartes parmi n (au plus C(7,5) = 21)
	combo := make([]int, 5)
	var generate func(start, depth int)
	generate = func(start, depth int) {
		if depth == 5 {
			hand := [5]Card{cards[combo[0]], cards[combo[1]], cards[combo[2]], cards[combo[3]], cards[combo[4]]}
			result := evaluate5(hand)
			if result.Score > best.Score {
				best = result
			}
			return
		}
		for i := start; i < n; i++ {
			combo[depth] = i
			generate(i+1, depth+1)
		}
	}
	generate(0, 0)

	return best
}

// evaluate5 évalue exactement 5 cartes et renvoie son score comparable
func evaluate5(cards [5]Card) HandResult {
	ranks := make([]int, 5)
	suits := make([]int, 5)
	for i, c := range cards {
		ranks[i] = c.Rank
		suits[i] = c.Suit
	}

	sortedCards := append([]Card{}, cards[:]...)
	sort.Slice(sortedCards, func(i, j int) bool { return sortedCards[i].Rank > sortedCards[j].Rank })

	isFlush := true
	for i := 1; i < 5; i++ {
		if suits[i] != suits[0] {
			isFlush = false
			break
		}
	}

	// Détection de suite (gère aussi la quinte basse As-2-3-4-5, dite "wheel")
	distinctRanks := make([]int, 5)
	copy(distinctRanks, ranks)
	sort.Sort(sort.Reverse(sort.IntSlice(distinctRanks)))

	isStraight := false
	straightHigh := 0
	if distinctRanks[0]-distinctRanks[4] == 4 && allUnique(distinctRanks) {
		isStraight = true
		straightHigh = distinctRanks[0]
	} else if allUnique(distinctRanks) && distinctRanks[0] == Ace && distinctRanks[1] == Five && distinctRanks[2] == Four && distinctRanks[3] == Three && distinctRanks[4] == Two {
		// Roue : A-2-3-4-5, l'As compte comme rang 1 (le plus faible)
		isStraight = true
		straightHigh = Five
	}

	if isStraight && isFlush {
		return HandResult{Score: encodeScore(CategoryStraightFlush, straightHigh), Category: CategoryStraightFlush, Best5: sortedCards}
	}

	// Comptage des occurrences par rang, trié par (nombre d'occurrences desc, rang desc)
	counts := map[int]int{}
	for _, r := range ranks {
		counts[r]++
	}
	rcs := make([]rc, 0, len(counts))
	for r, c := range counts {
		rcs = append(rcs, rc{r, c})
	}
	sort.Slice(rcs, func(i, j int) bool {
		if rcs[i].count != rcs[j].count {
			return rcs[i].count > rcs[j].count
		}
		return rcs[i].rank > rcs[j].rank
	})

	switch {
	case rcs[0].count == 4:
		return HandResult{Score: encodeScore(CategoryFourOfAKind, rcs[0].rank, rcs[1].rank), Category: CategoryFourOfAKind, Best5: sortedCards}
	case rcs[0].count == 3 && rcs[1].count == 2:
		return HandResult{Score: encodeScore(CategoryFullHouse, rcs[0].rank, rcs[1].rank), Category: CategoryFullHouse, Best5: sortedCards}
	case isFlush:
		return HandResult{Score: encodeScore(CategoryFlush, distinctRanks[0], distinctRanks[1], distinctRanks[2], distinctRanks[3], distinctRanks[4]), Category: CategoryFlush, Best5: sortedCards}
	case isStraight:
		return HandResult{Score: encodeScore(CategoryStraight, straightHigh), Category: CategoryStraight, Best5: sortedCards}
	case rcs[0].count == 3:
		kickers := kickersOf(rcs, 1)
		return HandResult{Score: encodeScore(CategoryThreeOfAKind, rcs[0].rank, kickers[0], kickers[1]), Category: CategoryThreeOfAKind, Best5: sortedCards}
	case rcs[0].count == 2 && rcs[1].count == 2:
		hi, lo := rcs[0].rank, rcs[1].rank
		if lo > hi {
			hi, lo = lo, hi
		}
		return HandResult{Score: encodeScore(CategoryTwoPair, hi, lo, rcs[2].rank), Category: CategoryTwoPair, Best5: sortedCards}
	case rcs[0].count == 2:
		kickers := kickersOf(rcs, 1)
		return HandResult{Score: encodeScore(CategoryPair, rcs[0].rank, kickers[0], kickers[1], kickers[2]), Category: CategoryPair, Best5: sortedCards}
	default:
		return HandResult{Score: encodeScore(CategoryHighCard, distinctRanks[0], distinctRanks[1], distinctRanks[2], distinctRanks[3], distinctRanks[4]), Category: CategoryHighCard, Best5: sortedCards}
	}
}

func allUnique(ranks []int) bool {
	seen := map[int]bool{}
	for _, r := range ranks {
		if seen[r] {
			return false
		}
		seen[r] = true
	}
	return true
}

func kickersOf(rcs []rc, fromIdx int) []int {
	kickers := make([]int, 0, 3)
	for i := fromIdx; i < len(rcs); i++ {
		kickers = append(kickers, rcs[i].rank)
	}
	for len(kickers) < 3 {
		kickers = append(kickers, 0)
	}
	return kickers
}
