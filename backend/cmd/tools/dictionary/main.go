package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	defaultLexiqueURL = "http://www.lexique.org/databases/Lexique383/Lexique383.zip"
)

// normalizeWord retire les accents, met en majuscules et vérifie la validité A-Z
func normalizeWord(s string, normalizer transform.Transformer) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", false
	}

	// Décomposition et suppression des diacritiques
	res, _, err := transform.String(normalizer, s)
	if err != nil {
		return "", false
	}

	// Gestion des ligatures œ et æ
	res = strings.ReplaceAll(res, "œ", "oe")
	res = strings.ReplaceAll(res, "Œ", "OE")
	res = strings.ReplaceAll(res, "æ", "ae")
	res = strings.ReplaceAll(res, "Æ", "AE")

	res = strings.ToUpper(res)

	// Vérification stricte : uniquement A-Z
	for _, r := range res {
		if r < 'A' || r > 'Z' {
			return "", false
		}
	}

	return res, true
}

func main() {
	minLen := flag.Int("min", 5, "Longueur minimale des mots")
	maxLen := flag.Int("max", 7, "Longueur maximale des mots")
	targetLen := flag.Int("target-len", 5, "Longueur exacte pour les mots cibles (targets.txt)")
	freqThreshold := flag.Float64("freq", 4.0, "Seuil de freqlivres minimum pour être dans targets.txt")
	lexiqueURL := flag.String("url", defaultLexiqueURL, "URL du fichier Lexique383.zip")
	outDir := flag.String("out", "backend/internal/service/games/wordle/dictionary/fr", "Dossier de sortie")
	flag.Parse()

	// Si exécuté depuis la racine ou depuis backend/
	resolvedOutDir := *outDir
	if _, err := os.Stat("go.mod"); err == nil {
		// Exécuté depuis backend/
		resolvedOutDir = strings.TrimPrefix(resolvedOutDir, "backend/")
	}
	if err := os.MkdirAll(resolvedOutDir, 0755); err != nil {
		log.Fatalf("❌ Impossible de créer le dossier de sortie : %v", err)
	}

	log.Printf("📥 Téléchargement de la base lexicale depuis %s ...", *lexiqueURL)
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(*lexiqueURL)
	if err != nil {
		log.Fatalf("❌ Erreur lors du téléchargement : %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("❌ Réponse HTTP inattendue : %s", resp.Status)
	}

	zipData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("❌ Erreur lors de la lecture du flux ZIP : %v", err)
	}
	log.Printf("📦 Archive ZIP téléchargée (%d octets). Décompression en mémoire...", len(zipData))

	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		log.Fatalf("❌ Erreur d'ouverture du ZIP : %v", err)
	}

	var tsvFile *zip.File
	for _, f := range zipReader.File {
		if strings.HasSuffix(strings.ToLower(f.Name), ".txt") || strings.HasSuffix(strings.ToLower(f.Name), ".tsv") {
			tsvFile = f
			break
		}
	}
	if tsvFile == nil {
		log.Fatalf("❌ Aucun fichier TSV/TXT trouvé dans l'archive ZIP")
	}

	log.Printf("📄 Traitement du fichier : %s", tsvFile.Name)
	rc, err := tsvFile.Open()
	if err != nil {
		log.Fatalf("❌ Impossible d'ouvrir le fichier dans le ZIP : %v", err)
	}
	defer rc.Close()

	// Transformer pour enlever les accents
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

	scanner := bufio.NewScanner(rc)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	// Lire la première ligne d'en-tête
	if !scanner.Scan() {
		log.Fatalf("❌ Fichier TSV vide")
	}
	headerLine := scanner.Text()
	headers := strings.Split(headerLine, "\t")
	
	colOrtho := -1
	colFreqLivres := -1
	colCgram := -1

	for i, h := range headers {
		switch strings.TrimSpace(h) {
		case "1_ortho", "ortho":
			colOrtho = i
		case "7_freqlivres", "freqlivres":
			colFreqLivres = i
		case "4_cgram", "cgram":
			colCgram = i
		}
	}

	if colOrtho == -1 {
		colOrtho = 0
	}
	if colFreqLivres == -1 {
		colFreqLivres = 6
	}

	log.Printf("🔍 Colonnes détectées: ortho=%d, freqlivres=%d, cgram=%d", colOrtho, colFreqLivres, colCgram)

	targetsMap := make(map[string]struct{})
	allValidMap := make(map[string]struct{})

	lineCount := 0
	for scanner.Scan() {
		lineCount++
		line := scanner.Text()
		cols := strings.Split(line, "\t")
		if len(cols) <= colOrtho {
			continue
		}

		rawWord := cols[colOrtho]
		word, ok := normalizeWord(rawWord, t)
		if !ok {
			continue
		}

		wordLen := len(word)
		if wordLen < *minLen || wordLen > *maxLen {
			continue
		}

		allValidMap[word] = struct{}{}

		// Sélection pour targets.txt :
		// 1. Longueur correspondante à targetLen (par défaut 5 lettres)
		// 2. Fréquence freqlivres >= threshold
		// 3. Exclure catégories grammaticales obscures (onomatopées, abréviations si possible)
		if wordLen == *targetLen && colFreqLivres < len(cols) {
			freqStr := cols[colFreqLivres]
			freq, _ := strconv.ParseFloat(strings.TrimSpace(freqStr), 64)
			
			isObscure := false
			if colCgram != -1 && colCgram < len(cols) {
				cgram := strings.TrimSpace(cols[colCgram])
				if cgram == "ONO" || cgram == "INTERJ" {
					isObscure = true
				}
			}

			if freq >= *freqThreshold && !isObscure {
				targetsMap[word] = struct{}{}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("❌ Erreur de lecture du TSV : %v", err)
	}

	log.Printf("📊 Lignes analysées : %d", lineCount)
	log.Printf("🎯 Mots cibles préliminaires : %d", len(targetsMap))
	log.Printf("📚 Total mots valides reconnus (5-7 lettres) : %d", len(allValidMap))

	// Règle d'or : partitionnement et déduplication stricte
	// allowed = allValid \ targets
	allowedMap := make(map[string]struct{})
	for w := range allValidMap {
		if _, inTarget := targetsMap[w]; !inTarget {
			allowedMap[w] = struct{}{}
		}
	}

	// Tri alphabétique des cibles
	targetList := make([]string, 0, len(targetsMap))
	for w := range targetsMap {
		targetList = append(targetList, w)
	}
	sort.Strings(targetList)

	// Tri alphabétique des autorisés additionnels
	allowedList := make([]string, 0, len(allowedMap))
	for w := range allowedMap {
		allowedList = append(allowedList, w)
	}
	sort.Strings(allowedList)

	// Écriture de targets.txt
	targetsFile := filepath.Join(resolvedOutDir, "targets.txt")
	if err := writeLines(targetsFile, targetList); err != nil {
		log.Fatalf("❌ Erreur écriture %s : %v", targetsFile, err)
	}
	log.Printf("✅ %s écrit avec succès (%d mots)", targetsFile, len(targetList))

	// Écriture de allowed.txt
	allowedFile := filepath.Join(resolvedOutDir, "allowed.txt")
	if err := writeLines(allowedFile, allowedList); err != nil {
		log.Fatalf("❌ Erreur écriture %s : %v", allowedFile, err)
	}
	log.Printf("✅ %s écrit avec succès (%d mots)", allowedFile, len(allowedList))

	log.Printf("🎉 Dictionnaire français généré avec succès ! Disjonction stricte : 0 doublon.")
}

func writeLines(path string, lines []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, l := range lines {
		if _, err := w.WriteString(l + "\n"); err != nil {
			return err
		}
	}
	return w.Flush()
}
