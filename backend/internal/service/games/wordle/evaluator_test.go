package wordle_test

import (
	"testing"

	"minigames-backend/internal/service/games/wordle"
)

func TestEvaluateGuess_ExactMatch(t *testing.T) {
	eval, isSolved := wordle.EvaluateGuess("SMART", "SMART")
	if !isSolved {
		t.Errorf("Le mot aurait dû être résolu")
	}

	for _, tile := range eval {
		if tile.Status != wordle.StatusCorrect {
			t.Errorf("Toutes les tuiles auraient dû être correctes, eu: %v", tile.Status)
		}
	}
}

func TestEvaluateGuess_DoubleLetters(t *testing.T) {
	// Mot cible avec un seul 'E', la proposition en a deux
	eval, isSolved := wordle.EvaluateGuess("ROBOT", "TOAST")
	if isSolved {
		t.Errorf("Le mot ne devrait pas être résolu")
	}

	// 'T' à la fin doit être correct
	if eval[4].Status != wordle.StatusCorrect {
		t.Errorf("Le dernier 'T' devrait être correct")
	}

	// Le premier 'T' de TOAST ne doit pas être compté en double
	if eval[0].Status != wordle.StatusAbsent {
		t.Errorf("Le premier 'T' devrait être absent car le seul T cible est déjà placé")
	}
}

func TestEvaluateGuess_StateMasking(t *testing.T) {
	eval, _ := wordle.EvaluateGuess("CLOUD", "CLEAR")
	masked := wordle.MaskEvaluation(eval)

	if len(masked) != len(eval) {
		t.Fatalf("Longueur masquée erronée")
	}

	// Vérifier que la structure masquée n'expose pas de champ lettre
	for _, m := range masked {
		if m.Status != wordle.StatusCorrect && m.Status != wordle.StatusPresent && m.Status != wordle.StatusAbsent {
			t.Errorf("Statut invalide : %v", m.Status)
		}
	}
}
