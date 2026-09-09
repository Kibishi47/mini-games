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

// Candidate représente un mot éligible à la liste des cibles avec son score de popularité combiné
type Candidate struct {
	Word  string
	Score float64
}

var topNByLength = map[int]int{
	3: 200,
	4: 600,
	5: 1500,
	6: 2000,
	7: 2000,
	8: 1500,
}

func main() {
	minLen := flag.Int("min", 3, "Longueur minimale des mots")
	maxLen := flag.Int("max", 8, "Longueur maximale des mots")
	freqFilmsMin := flag.Float64("freqfilms-min", 1.5, "Seuil minimum plancher freqfilms (oral/cinéma)")
	freqLivresMin := flag.Float64("freqlivres-min", 1.0, "Seuil minimum plancher freqlivres (écrit/littéraire)")
	minScore := flag.Float64("min-score", 2.5, "Score combiné minimum d'éligibilité pour targets")
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

	// Transformer pour enlever les accents et normaliser en ASCII majuscule
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

	// 1. Chargement de blacklist.txt si existant
	blacklistedMap := make(map[string]struct{})
	blacklistPath := filepath.Join(resolvedOutDir, "blacklist.txt")
	if fBlacklist, err := os.Open(blacklistPath); err == nil {
		scannerBlacklist := bufio.NewScanner(fBlacklist)
		for scannerBlacklist.Scan() {
			line := strings.TrimSpace(scannerBlacklist.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			normWord, ok := normalizeWord(line, t)
			if ok {
				blacklistedMap[normWord] = struct{}{}
			}
		}
		fBlacklist.Close()
		log.Printf("🚫 Blacklist chargée depuis %s : %d mots exclus des cibles", blacklistPath, len(blacklistedMap))
	} else {
		log.Printf("ℹ️ Aucun fichier blacklist.txt trouvé à %s (ignoré)", blacklistPath)
	}

	// 2. Téléchargement et ouverture de l'archive Lexique 383
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
		colFreqLivres = 9
	}
	if colFreqFilms == -1 {
		colFreqFilms = 8
	}

	log.Printf("🔍 Colonnes détectées: ortho=%d, lemme=%d, cgram=%d, nombre=%d, freqlivres=%d, freqfilms=%d",
		colOrtho, colLemme, colCgram, colNombre, colFreqLivres, colFreqFilms)

	candidatesByLength := make(map[int]map[string]float64)
	for l := *minLen; l <= *maxLen; l++ {
		candidatesByLength[l] = make(map[string]float64)
	}
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

		// Tout mot valide de 3 à 8 lettres entre dans la réserve globale
		allValidMap[word] = struct{}{}

		// -------------------------------------------------------------
		// Critères stricts d'éligibilité pour targets.txt
		// -------------------------------------------------------------
		// 0. Si présent dans la blacklist, rejet des cibles immédiat
		if _, isBlacklisted := blacklistedMap[word]; isBlacklisted {
			continue
		}

		// 1. Forme canonique pure :
		//    orthographe normalisée == lemme normalisé
		var rawLemme string
		if colLemme < len(cols) {
			rawLemme = cols[colLemme]
		}
		normLemme, okLemme := normalizeWord(rawLemme, t)
		if !okLemme || normLemme != word {
			continue
		}

		// 2. Exclusion des formes plurielles : nombre == "s" ou vide
		if colNombre < len(cols) {
			nombre := strings.TrimSpace(cols[colNombre])
			if nombre != "" && nombre != "s" {
				continue
			}
		}

		// 3. Catégorie grammaticale sémantiquement forte : NOM, VER, ADJ
		if colCgram >= len(cols) {
			continue
		}
		cgram := strings.TrimSpace(cols[colCgram])
		if cgram != "NOM" && cgram != "VER" && cgram != "ADJ" {
			continue
		}

		// 4. Plancher anti-bruit : freqfilms >= 1.5 ET freqlivres >= 1.0
		var freqLivres, freqFilms float64
		if colFreqLivres < len(cols) {
			freqLivres, _ = strconv.ParseFloat(strings.TrimSpace(cols[colFreqLivres]), 64)
		}
		if colFreqFilms < len(cols) {
			freqFilms, _ = strconv.ParseFloat(strings.TrimSpace(cols[colFreqFilms]), 64)
		}

		if freqFilms < *freqFilmsMin || freqLivres < *freqLivresMin {
			continue
		}

		// 5. Calcul du score de popularité combiné (sur-pondération oral contemporain)
		score := (freqFilms * 0.65) + (freqLivres * 0.35)

		// 6. Condition d'éligibilité : score >= minScore (par défaut 2.5)
		if score < *minScore {
			continue
		}

		// Conserver le meilleur score si le mot apparaît plusieurs fois (ex: homographes NOM/VER)
		if existingScore, exists := candidatesByLength[wordLen][word]; !exists || score > existingScore {
			candidatesByLength[wordLen][word] = score
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("❌ Erreur de lecture du TSV : %v", err)
	}

	log.Printf("📊 Lignes analysées : %d", lineCount)
	log.Printf("📚 Total mots valides reconnus (3-8 lettres) : %d", len(allValidMap))

	// Plafonnement Top N par longueur de mot
	targetsMap := make(map[string]struct{})
	targetsByLenCount := make(map[int]int)

	for l := *minLen; l <= *maxLen; l++ {
		candidateMap := candidatesByLength[l]
		candidates := make([]Candidate, 0, len(candidateMap))
		for w, score := range candidateMap {
			candidates = append(candidates, Candidate{Word: w, Score: score})
		}

		// Tri décroissant par score de popularité
		sort.Slice(candidates, func(i, j int) bool {
			if candidates[i].Score == candidates[j].Score {
				return candidates[i].Word < candidates[j].Word
			}
			return candidates[i].Score > candidates[j].Score
		})

		quota, ok := topNByLength[l]
		if !ok {
			quota = 750
		}

		selectedCount := len(candidates)
		if selectedCount > quota {
			selectedCount = quota
		}

		for i := 0; i < selectedCount; i++ {
			targetsMap[candidates[i].Word] = struct{}{}
		}
		targetsByLenCount[l] = selectedCount
	}

	// Disjonction stricte : allowed = allValid \ targets
	allowedMap := make(map[string]struct{})
	for w := range allValidMap {
		if _, inTarget := targetsMap[w]; !inTarget {
			allowedMap[w] = struct{}{}
		}
	}

	// Tri alphabétique de targets.txt
	targetList := make([]string, 0, len(targetsMap))
	for w := range targetsMap {
		targetList = append(targetList, w)
	}
	sort.Strings(targetList)

	// Tri alphabétique de allowed.txt
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

	// Logs détaillés finaux
	log.Println("=========================================================")
	log.Println("📋 RÉCAPITULATIF DE LA GÉNÉRATION DU DICTIONNAIRE")
	log.Println("=========================================================")
	log.Printf("🚫 Mots dans blacklist.txt pris en compte : %d", len(blacklistedMap))
	log.Println("🎯 Nombre de mots cibles générés par longueur (targets.txt) :")
	for l := *minLen; l <= *maxLen; l++ {
		log.Printf("   • %d lettres : %d mots (sur Top %d)", l, targetsByLenCount[l], topNByLength[l])
	}
	log.Printf("🏆 Nombre total de mots dans targets.txt : %d", len(targetList))
	log.Printf("📚 Nombre total de mots dans allowed.txt : %d", len(allowedList))

	// Assertion de validation stricte : 0 intersection
	intersectionCount := 0
	for _, w := range targetList {
		if _, exists := allowedMap[w]; exists {
			intersectionCount++
		}
	}

	if intersectionCount > 0 {
		log.Fatalf("❌ ÉCHEC CRITIQUE : %d doublons détectés entre targets.txt et allowed.txt !", intersectionCount)
	}

	log.Printf("🛡️ Assertion vérifiée avec succès : strictement 0 intersection entre targets.txt et allowed.txt !")
	log.Println("=========================================================")
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
