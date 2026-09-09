package wordle

import (
	"bufio"
	"crypto/rand"
	"math/big"
	"os"
	"strings"
	"sync"
)

type Dictionary struct {
	mu    sync.RWMutex
	words map[string]map[int][]string // lang -> length -> list of words
	valid map[string]map[string]bool  // lang -> word -> bool
}

func NewDictionary() *Dictionary {
	return &Dictionary{
		words: make(map[string]map[int][]string),
		valid: make(map[string]map[string]bool),
	}
}

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
		}
	}

	return scanner.Err()
}

func (d *Dictionary) GetRandomWord(lang string, length int) string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	wordList, ok := d.words[lang][length]
	if !ok || len(wordList) == 0 {
		// Fallback anglais si introuvable
		wordList = d.words["en"][length]
		if len(wordList) == 0 {
			return "SMART" // secours par défaut
		}
	}

	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(wordList))))
	return wordList[n.Int64()]
}

func (d *Dictionary) IsValidWord(lang, word string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	w := strings.ToUpper(strings.TrimSpace(word))
	if langMap, ok := d.valid[lang]; ok {
		return langMap[w]
	}
	return false
}
