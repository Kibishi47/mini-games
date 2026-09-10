package wordle

import (
	"bufio"
	"crypto/rand"
	_ "embed"
	"math/big"
	"os"
	"strings"
	"sync"
)

//go:embed dictionary/fr/targets.txt
var frTargetsRaw string

//go:embed dictionary/fr/allowed.txt
var frAllowedRaw string

type Dictionary struct {
	mu sync.RWMutex

	// Moteur moderne Wordle (spécifique FR embed)
	targets         []string
	targetsByLength map[int][]string // length -> list of secret targets ONLY
	allowedSet      map[string]struct{}

	// Rétrocompatibilité multi-langues et multi-tailles
	words map[string]map[int][]string // lang -> length -> list of words
	valid map[string]map[string]bool  // lang -> word -> bool
}

// NewDictionary initialise le dictionnaire avec les listes françaises embarquées via //go:embed
func NewDictionary() *Dictionary {
	d := &Dictionary{
		targets:         make([]string, 0),
		targetsByLength: make(map[int][]string),
		allowedSet:      make(map[string]struct{}),
		words:           make(map[string]map[int][]string),
		valid:           make(map[string]map[string]bool),
	}

	d.words["fr"] = make(map[int][]string)
	d.valid["fr"] = make(map[string]bool)

	// Charger targets.txt (Cibles tirables UNIQUEMENT)
	scannerTargets := bufio.NewScanner(strings.NewReader(frTargetsRaw))
	for scannerTargets.Scan() {
		w := strings.ToUpper(strings.TrimSpace(scannerTargets.Text()))
		if w == "" {
			continue
		}
		d.targets = append(d.targets, w)
		l := len(w)
		d.targetsByLength[l] = append(d.targetsByLength[l], w)
		d.allowedSet[w] = struct{}{}

		// Rétrocompatibilité
		d.words["fr"][l] = append(d.words["fr"][l], w)
		d.valid["fr"][w] = true
	}

	// Charger allowed.txt (Mots autorisés additionnels pour la VALIDATION uniquement, JAMAIS tirés au sort)
	scannerAllowed := bufio.NewScanner(strings.NewReader(frAllowedRaw))
	for scannerAllowed.Scan() {
		w := strings.ToUpper(strings.TrimSpace(scannerAllowed.Text()))
		if w == "" {
			continue
		}
		d.allowedSet[w] = struct{}{}

		// Rétrocompatibilité pour la validation de mot
		l := len(w)
		d.words["fr"][l] = append(d.words["fr"][l], w)
		d.valid["fr"][w] = true
	}

	return d
}

// PickRandom sélectionne un mot aléatoire dans la liste targets en O(1)
func (d *Dictionary) PickRandom() string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(d.targets) == 0 {
		return "POMME"
	}

	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(d.targets))))
	if err != nil {
		return d.targets[0]
	}
	return d.targets[n.Int64()]
}

// PickRandomByLength sélectionne un mot secret cible aléatoire d'une longueur spécifique UNIQUEMENT parmi targets.txt en O(1)
func (d *Dictionary) PickRandomByLength(length int) string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if targetList, ok := d.targetsByLength[length]; ok && len(targetList) > 0 {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(targetList))))
		if err == nil {
			return targetList[n.Int64()]
		}
		return targetList[0]
	}

	return d.PickRandom()
}

// IsValid vérifie en O(1) si un mot soumis figure dans le dictionnaire étendu (targets + allowed)
func (d *Dictionary) IsValid(guess string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	w := strings.ToUpper(strings.TrimSpace(guess))
	_, exists := d.allowedSet[w]
	return exists
}

// TargetsCount retourne le nombre de mots secrets cibles
func (d *Dictionary) TargetsCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.targets)
}

// AllowedCount retourne le nombre total de mots valides acceptés
func (d *Dictionary) AllowedCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.allowedSet)
}

// LoadFromFile permet d'enrichir le dictionnaire depuis un fichier externe
func (d *Dictionary) LoadFromFile(lang, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.words[lang]; !ok {
		d.words[lang] = make(map[int][]string)
		d.valid[lang] = make(map[string]bool)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		w := strings.ToUpper(strings.TrimSpace(scanner.Text()))
		if len(w) >= 3 && len(w) <= 10 {
			l := len(w)
			d.words[lang][l] = append(d.words[lang][l], w)
			d.valid[lang][w] = true
			if lang == "fr" {
				d.allowedSet[w] = struct{}{}
			}
		}
	}

	return scanner.Err()
}

// GetRandomWord sélectionne un mot selon la langue et la taille
func (d *Dictionary) GetRandomWord(lang string, length int) string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if lang == "fr" || lang == "" {
		if length <= 0 {
			length = 5
		}
		if targetList, ok := d.targetsByLength[length]; ok && len(targetList) > 0 {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(len(targetList))))
			if err == nil {
				return targetList[n.Int64()]
			}
			return targetList[0]
		}
		if len(d.targets) > 0 {
			return d.targets[0]
		}
	}

	wordList, ok := d.words[lang][length]
	if !ok || len(wordList) == 0 {
		// Fallback anglais ou targets
		wordList = d.words["en"][length]
		if len(wordList) == 0 {
			if len(d.targets) > 0 {
				return d.targets[0]
			}
			return "SMART"
		}
	}

	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(wordList))))
	return wordList[n.Int64()]
}

// IsValidWord vérifie la validité d'un mot pour une langue donnée
func (d *Dictionary) IsValidWord(lang, word string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	w := strings.ToUpper(strings.TrimSpace(word))
	if lang == "fr" || lang == "" {
		if _, exists := d.allowedSet[w]; exists {
			return true
		}
	}

	if langMap, ok := d.valid[lang]; ok {
		return langMap[w]
	}
	return false
}
