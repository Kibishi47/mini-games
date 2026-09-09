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
	minLen := flag.Int("min", 3, "Longueur minimale des mots")
	maxLen := flag.Int("max", 8, "Longueur maximale des mots")
	freqThreshold := flag.Float64("freq", 5.0, "Seuil de fréquence combinée pour targets")
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
	colLemme := -1
	colCgram := -1
	colNombre := -1
	colFreqLivres := -1
	colFreqFilms := -1

	for i, h := range headers {
		cleanHeader := strings.TrimSpace(h)
		switch cleanHeader {
		case "1_ortho", "ortho":
			colOrtho = i
		case "3_lemme", "lemme":
			colLemme = i
		case "4_cgram", "cgram":
			colCgram = i
		case "6_nombre", "nombre":
			colNombre = i
		case "7_freqlivres", "freqlivres":
			colFreqLivres = i
		case "8_freqfilms2", "freqfilms2", "freqfilms":
			colFreqFilms = i
		}
	}

	if colOrtho == -1 {
		colOrtho = 0
	}
	if colLemme == -1 {
		colLemme = 2
	}
	if colCgram == -1 {
		colCgram = 3
	}
	if colNombre == -1 {
		colNombre = 5
	}
	if colFreqLivres == -1 {
		colFreqLivres = 6
	}
	if colFreqFilms == -1 {
		colFreqFilms = 7
	}

	log.Printf("🔍 Colonnes détectées: ortho=%d, lemme=%d, cgram=%d, nombre=%d, freqlivres=%d, freqfilms=%d",
		colOrtho, colLemme, colCgram, colNombre, colFreqLivres, colFreqFilms)

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

		// Tout mot valide de 3 à 8 lettres entre dans le dictionnaire global
		allValidMap[word] = struct{}{}

		// -------------------------------------------------------------
		// Critères d'éligibilité pour targets.txt (Mots cibles canoniques)
		// -------------------------------------------------------------
		// 1. Longueur : minLen..maxLen (3..8)
		// 2. Forme canonique pure :
		//    - ortho normalisé == lemme normalisé
		//    - nombre == "s" ou vide (pas de pluriel 'p')
		// 3. Catégorie grammaticale sémantiquement forte : NOM, VER, ADJ
		//    - Exclure formellement : ONO, INTERJ, PRO, ART, PRE, CON, etc.
		// 4. Fréquence d'usage usuel :
		//    - (freqfilms >= 8.0 || freqlivres >= 8.0) && (freqlivres + freqfilms) / 2 >= freqThreshold
		// -------------------------------------------------------------
		var rawLemme string
		if colLemme < len(cols) {
			rawLemme = cols[colLemme]
		}
		normLemme, okLemme := normalizeWord(rawLemme, t)
		if !okLemme || normLemme != word {
			continue
		}

		// Vérification du nombre (pas de pluriel)
		if colNombre < len(cols) {
			nombre := strings.TrimSpace(cols[colNombre])
			if nombre != "" && nombre != "s" {
				continue
			}
		}

		// Catégorie grammaticale
		if colCgram >= len(cols) {
			continue
		}
		cgram := strings.TrimSpace(cols[colCgram])
		if cgram != "NOM" && cgram != "VER" && cgram != "ADJ" {
			continue
		}

		// Fréquences freqlivres et freqfilms
		var freqLivres, freqFilms float64
		if colFreqLivres < len(cols) {
			freqLivres, _ = strconv.ParseFloat(strings.TrimSpace(cols[colFreqLivres]), 64)
		}
		if colFreqFilms < len(cols) {
			freqFilms, _ = strconv.ParseFloat(strings.TrimSpace(cols[colFreqFilms]), 64)
		}

		meanFreq := (freqLivres + freqFilms) / 2.0
		isFrequent := (freqFilms >= 8.0 || freqLivres >= 8.0) && meanFreq >= *freqThreshold
		if isFrequent {
			targetsMap[word] = struct{}{}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("❌ Erreur de lecture du TSV : %v", err)
	}

	log.Printf("📊 Lignes analysées : %d", lineCount)
	log.Printf("🎯 Total mots cibles canoniques retenus (3-8 lettres) : %d", len(targetsMap))
	log.Printf("📚 Total mots valides reconnus (3-8 lettres) : %d", len(allValidMap))

	// Règle absolue : partitionnement et déduplication stricte
	// allowed = allValid \ targets
	allowedMap := make(map[string]struct{})
	for w := range allValidMap {
		if _, inTarget := targetsMap[w]; !inTarget {
			allowedMap[w] = struct{}{}
		}
	}

	// Tri alphabétique des cibles
	targetList := make([]string, 0, len(targetsMap))
	targetsByLen := make(map[int]int)
	for w := range targetsMap {
		targetList = append(targetList, w)
		targetsByLen[len(w)]++
	}
	sort.Strings(targetList)

	// Tri alphabétique des mots autorisés additionnels
	allowedList := make([]string, 0, len(allowedMap))
	allowedByLen := make(map[int]int)
	for w := range allowedMap {
		allowedList = append(allowedList, w)
		allowedByLen[len(w)]++
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

	// Logs détaillés par longueur
	log.Printf("📈 Répartition des cibles (targets.txt) par longueur :")
	for l := *minLen; l <= *maxLen; l++ {
		log.Printf("   • %d lettres : %d mots", l, targetsByLen[l])
	}

	log.Printf("📈 Répartition des mots autorisés (allowed.txt) par longueur :")
	for l := *minLen; l <= *maxLen; l++ {
		log.Printf("   • %d lettres : %d mots", l, allowedByLen[l])
	}

	// Vérification de disjonction stricte
	intersectionCount := 0
	for _, w := range targetList {
		if _, exists := allowedMap[w]; exists {
			intersectionCount++
		}
	}

	if intersectionCount > 0 {
		log.Fatalf("❌ ÉCHEC : %d doublons détectés entre targets.txt et allowed.txt !", intersectionCount)
	}

	log.Printf("🎉 Confirmation explicite : 0 intersection entre targets.txt et allowed.txt !")
	log.Printf("🚀 Génération du dictionnaire français 3-8 lettres achevée avec succès !")
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
