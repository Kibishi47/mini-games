package wordle

import (
	"strings"
)

type TileStatus string

const (
	StatusCorrect TileStatus = "correct" // Vert (bonne lettre, bonne place)
	StatusPresent TileStatus = "present" // Jaune (bonne lettre, mauvaise place)
	StatusAbsent  TileStatus = "absent"  // Gris (lettre absente)
)

type TileEvaluation struct {
	Letter string     `json:"letter"`
	Status TileStatus `json:"status"`
}

// MaskedTileEvaluation ne transmet QUE le statut aux adversaires, JAMAIS la lettre réelle !
type MaskedTileEvaluation struct {
	Status TileStatus `json:"status"`
}

// EvaluateGuess évalue une proposition par rapport au mot cible selon les règles officielles Wordle
// Gère parfaitement les lettres en double (occurrences limitées)
func EvaluateGuess(target, guess string) ([]TileEvaluation, bool) {
	target = strings.ToUpper(strings.TrimSpace(target))
	guess = strings.ToUpper(strings.TrimSpace(guess))

	length := len(target)
	result := make([]TileEvaluation, length)

	targetChars := []rune(target)
	guessChars := []rune(guess)

	// Fréquence des lettres restantes dans le mot cible
	targetLetterCounts := make(map[rune]int)
	for _, ch := range targetChars {
		targetLetterCounts[ch]++
	}

	isSolved := true

	// 1ère passe : identifier les tuiles Vertes (Correctes)
	for i := 0; i < length; i++ {
		g := guessChars[i]
		result[i].Letter = string(g)

		if g == targetChars[i] {
			result[i].Status = StatusCorrect
			targetLetterCounts[g]--
		} else {
			isSolved = false
		}
	}

	// 2ème passe : identifier les tuiles Jaunes (Présentes) et Grises (Absentes)
	for i := 0; i < length; i++ {
		if result[i].Status == StatusCorrect {
			continue
		}

		g := guessChars[i]
		if targetLetterCounts[g] > 0 {
			result[i].Status = StatusPresent
			targetLetterCounts[g]--
		} else {
			result[i].Status = StatusAbsent
		}
	}

	return result, isSolved
}

// MaskEvaluation masque les lettres pour le partage public / temps réel aux adversaires
func MaskEvaluation(eval []TileEvaluation) []MaskedTileEvaluation {
	masked := make([]MaskedTileEvaluation, len(eval))
	for i, e := range eval {
		masked[i] = MaskedTileEvaluation{
			Status: e.Status,
		}
	}
	return masked
}

// GenerateEmojiGrid produit la représentation émojis virale (🟩🟨⬛)
func GenerateEmojiGrid(attempts [][]TileEvaluation) string {
	var sb strings.Builder
	for _, row := range attempts {
		for _, tile := range row {
			switch tile.Status {
			case StatusCorrect:
				sb.WriteString("🟩")
			case StatusPresent:
				sb.WriteString("🟨")
			default:
				sb.WriteString("⬛")
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
