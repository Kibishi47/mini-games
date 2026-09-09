package wordle

import (
	"bufio"
	"strings"
	"testing"
)

func TestDictionary_EmbeddingAndDisjoint(t *testing.T) {
	// 1. Charger et parser targets.txt
	targetsMap := make(map[string]int)
	scannerTargets := bufio.NewScanner(strings.NewReader(frTargetsRaw))
	targetCount := 0
	for scannerTargets.Scan() {
		w := strings.TrimSpace(scannerTargets.Text())
		if w == "" {
			continue
		}
		targetCount++

		// Vérifier que le mot est en majuscules non accentuées [A-Z]+
		for _, r := range w {
			if r < 'A' || r > 'Z' {
				t.Fatalf("Le mot cible '%s' contient un caractère non autorisé (doit être A-Z)", w)
			}
		}

		if len(w) < 3 || len(w) > 8 {
			t.Errorf("Le mot cible '%s' n'a pas une longueur entre 3 et 8 lettres (longueur=%d)", w, len(w))
		}

		targetsMap[w]++
		if targetsMap[w] > 1 {
			t.Fatalf("Doublon interne détecté dans targets.txt : '%s'", w)
		}
	}

	if targetCount == 0 {
		t.Fatalf("targets.txt est vide !")
	}

	// 2. Charger et parser allowed.txt
	allowedMap := make(map[string]int)
	scannerAllowed := bufio.NewScanner(strings.NewReader(frAllowedRaw))
	allowedCount := 0
	for scannerAllowed.Scan() {
		w := strings.TrimSpace(scannerAllowed.Text())
		if w == "" {
			continue
		}
		allowedCount++

		// Vérifier que le mot est en majuscules non accentuées [A-Z]+
		for _, r := range w {
			if r < 'A' || r > 'Z' {
				t.Fatalf("Le mot autorisé '%s' contient un caractère non autorisé (doit être A-Z)", w)
			}
		}

		if len(w) < 3 || len(w) > 8 {
			t.Errorf("Le mot autorisé '%s' n'a pas une longueur entre 3 et 8 lettres (longueur=%d)", w, len(w))
		}

		// Règle d'or absolue : ZERO doublon avec targets.txt
		if _, existsInTargets := targetsMap[w]; existsInTargets {
			t.Fatalf("VIOLATION DE DISJONCTION : le mot '%s' est présent à la fois dans targets.txt et allowed.txt !", w)
		}

		allowedMap[w]++
		if allowedMap[w] > 1 {
			t.Fatalf("Doublon interne détecté dans allowed.txt : '%s'", w)
		}
	}

	if allowedCount == 0 {
		t.Fatalf("allowed.txt est vide !")
	}

	t.Logf("✅ Disjonction stricte validée : %d mots cibles, %d mots autorisés additionnels, 0 doublon.", targetCount, allowedCount)
}

func TestDictionary_EngineMethods(t *testing.T) {
	dict := NewDictionary()

	if dict.TargetsCount() == 0 {
		t.Fatalf("TargetsCount() doit être > 0")
	}

	if dict.AllowedCount() != dict.TargetsCount()+len(strings.Split(strings.TrimSpace(frAllowedRaw), "\n")) {
		t.Errorf("AllowedCount() attendu = %d, obtenu = %d", dict.TargetsCount()+len(strings.Split(strings.TrimSpace(frAllowedRaw), "\n")), dict.AllowedCount())
	}

	// Test PickRandom()
	picked := dict.PickRandom()
	if len(picked) < 3 || len(picked) > 8 {
		t.Errorf("PickRandom() a retourné '%s' dont la longueur n'est pas dans [3, 8]", picked)
	}

	// Test PickRandomByLength pour chaque longueur de 3 à 8
	for l := 3; l <= 8; l++ {
		pickedLen := dict.PickRandomByLength(l)
		if len(pickedLen) != l {
			t.Errorf("PickRandomByLength(%d) a retourné '%s' de longueur %d", l, pickedLen, len(pickedLen))
		}
		if !dict.IsValid(pickedLen) {
			t.Errorf("Le mot '%s' tiré pour la longueur %d n'est pas reconnu par IsValid()", pickedLen, l)
		}
	}

	// Le mot pioché doit être valide et appartenir au dictionnaire
	if !dict.IsValid(picked) {
		t.Errorf("Le mot tiré au sort '%s' n'est pas reconnu par IsValid()", picked)
	}

	// Test IsValid() pour un mot cible
	firstTarget := dict.targets[0]
	if !dict.IsValid(firstTarget) {
		t.Errorf("Le mot cible '%s' devrait être valide", firstTarget)
	}

	// Test IsValid() pour un mot autorisé non cible
	scannerAllowed := bufio.NewScanner(strings.NewReader(frAllowedRaw))
	if scannerAllowed.Scan() {
		sampleAllowed := strings.TrimSpace(scannerAllowed.Text())
		if !dict.IsValid(sampleAllowed) {
			t.Errorf("Le mot autorisé '%s' devrait être valide dans IsValid()", sampleAllowed)
		}
	}

	// Test IsValid() avec casse et espaces (insensible à la casse / trim)
	lowerTarget := strings.ToLower(firstTarget)
	if !dict.IsValid(lowerTarget) {
		t.Errorf("IsValid('%s') devrait être insensible à la casse", lowerTarget)
	}
	if !dict.IsValid("  " + firstTarget + "  ") {
		t.Errorf("IsValid() devrait accepter les espaces de bordure")
	}

	// Test mot bidon invalide
	if dict.IsValid("ZZZZZZZZZ") {
		t.Errorf("Le mot non existant 'ZZZZZZZZZ' ne devrait pas être valide")
	}
	if dict.IsValid("12345") {
		t.Errorf("Le mot '12345' ne devrait pas être valide")
	}
}
