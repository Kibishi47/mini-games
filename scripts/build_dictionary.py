#!/usr/bin/env python3
"""
Génération automatisée des dictionnaires français pour le Wordle multijoueur.
Produit :
1. targets.txt : mots secrets cibles (exclusivement verbes à l'infinitif, noms et adjectifs
                 singuliers, simples et familiers, SANS AUCUN verbe conjugué).
2. allowed.txt : dictionnaire complet de validation contenant les formes fléchies et mots additionnels.
"""

import csv
import io
import os
import re
import sys
import unicodedata
import zipfile
from pathlib import Path
import requests
from wordfreq import zipf_frequency

SOURCE_ALL_URL = "https://www.listesdemots.net/touslesmots.txt"
LEXIQUE_URL = "http://www.lexique.org/databases/Lexique383/Lexique383.zip"

MIN_LEN = 3
MAX_LEN = 8

# ==============================================================================
# PARAMÈTRE DE DIFFICULTÉ (1 à 10)
# 1  = Super simple (vocabulaire ultra courant du quotidien, Zipf ~ 4.90, Score ~ 28.0)
# 5  = Standard / Équilibré (valeurs par défaut d'origine, Zipf ~ 4.30, Score ~ 13.0)
# 10 = Super dur (vocabulaire riche, soutenu et rare, Zipf ~ 3.50, Score ~ 1.0)
# Peut également être défini via la variable d'environnement WORDLE_DIFFICULTY ou en argument CLI : python build_dictionary.py [1-10]
# ==============================================================================
DIFFICULTY = 5

# Récupération dynamique si passé en CLI ou variable d'environnement
if len(sys.argv) > 1 and sys.argv[1].isdigit():
    DIFFICULTY = int(sys.argv[1])
elif os.environ.get("WORDLE_DIFFICULTY"):
    try:
        DIFFICULTY = int(os.environ["WORDLE_DIFFICULTY"])
    except ValueError:
        pass

DIFFICULTY = max(1, min(10, DIFFICULTY))

# Calcul continu des seuils en fonction du niveau de difficulté (1 -> 10) :
# - Plus la difficulté est faible (1), plus on exige une fréquence Zipf et un score d'usage élevés (mots très simples).
# - Plus la difficulté est élevée (10), plus on abaisse les seuils de fréquence pour accepter des mots plus rares et complexes.
TARGET_MIN_ZIPF = round(4.90 - ((DIFFICULTY - 1) / 9.0) * (4.90 - 3.50), 2)
MIN_SCORE = round(28.0 - ((DIFFICULTY - 1) / 9.0) * (28.0 - 1.0), 1)

def normalize_word(raw: str) -> str:
    """
    Normalise un mot :
    - Remplacement des ligatures (œ -> oe, æ -> ae)
    - Décomposition Unicode et suppression des diacritiques
    - Majuscules A-Z strictes
    """
    s = raw.strip()
    s = s.replace("œ", "oe").replace("Œ", "OE")
    s = s.replace("æ", "ae").replace("Æ", "AE")
    
    nfkd = unicodedata.normalize("NFD", s)
    no_accents = "".join(c for c in nfkd if unicodedata.category(c) != "Mn")
    
    clean = no_accents.upper()
    if re.fullmatch(r"[A-Z]+", clean):
        return clean
    return ""

def find_output_dir() -> Path:
    """
    Localise le dossier destination de manière robuste qu'on soit exécuté
    depuis la racine du projet, depuis scripts/ ou dans Docker.
    """
    candidate_paths = [
        Path("backend/internal/service/games/wordle/dictionary/fr"),
        Path("../backend/internal/service/games/wordle/dictionary/fr"),
        Path("/app/backend/internal/service/games/wordle/dictionary/fr"),
        Path("internal/service/games/wordle/dictionary/fr"),
    ]
    for p in candidate_paths:
        if p.parent.exists() or p.exists():
            p.mkdir(parents=True, exist_ok=True)
            return p
    fallback = Path("backend/internal/service/games/wordle/dictionary/fr")
    fallback.mkdir(parents=True, exist_ok=True)
    return fallback

def load_blacklist(out_dir: Path) -> set:
    bl_file = out_dir / "blacklist.txt"
    if bl_file.exists():
        return {normalize_word(line) for line in bl_file.read_text().splitlines() if line.strip()}
    return set()

def main():
    print("=" * 60)
    print("🚀 Génération des dictionnaires Wordle Français")
    print(f"   Difficulté : {DIFFICULTY}/10 (1=Super simple, 10=Super dur)")
    print(f"   Longueur   : {MIN_LEN} à {MAX_LEN} lettres | Zipf >= {TARGET_MIN_ZIPF} | Score >= {MIN_SCORE}")
    print("   Règle      : AUCUN verbe conjugué (infinitifs, noms, adjectifs)")
    print("=" * 60)

    out_dir = find_output_dir()
    blacklist = load_blacklist(out_dir)

    # 1. Téléchargement de la liste complète de validation
    print(f"📥 Téléchargement de la liste exhaustive : {SOURCE_ALL_URL} ...")
    headers = {"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"}
    resp = requests.get(SOURCE_ALL_URL, headers=headers, timeout=30)
    resp.raise_for_status()

    raw_text = resp.content.decode("utf-8", errors="ignore")
    all_valid = set()
    for token in raw_text.split():
        word = normalize_word(token)
        if word and MIN_LEN <= len(word) <= MAX_LEN:
            all_valid.add(word)

    print(f"✅ {len(all_valid):,} mots valides normalisés ({MIN_LEN}-{MAX_LEN} lettres).")

    # 2. Récupération de Lexique383 pour identifier précisément les infinitifs et les noms/adjectifs singuliers
    print(f"📚 Téléchargement de Lexique383 pour le filtrage canonique : {LEXIQUE_URL} ...")
    lex_resp = requests.get(LEXIQUE_URL, headers=headers, timeout=60)
    lex_resp.raise_for_status()

    lex_zip = zipfile.ZipFile(io.BytesIO(lex_resp.content))
    tsv_name = [n for n in lex_zip.namelist() if n.endswith(".tsv")][0]

    verb_inf_freq = {}
    verb_conj_freq = {}
    noun_adj_freq = {}

    with lex_zip.open(tsv_name) as f:
        reader = csv.DictReader(io.TextIOWrapper(f, encoding="utf-8", errors="ignore"), delimiter="\t")
        for row in reader:
            ortho = row["ortho"].strip().lower()
            cgram = row["cgram"].strip()
            islem = row["islem"].strip()
            nomb = row["nombre"].strip()
            
            try:
                ff = float(row["freqfilms2"])
                fl = float(row["freqlivres"])
            except (ValueError, TypeError):
                ff, fl = 0.0, 0.0
            score = ff * 0.7 + fl * 0.3

            if cgram == "VER":
                if islem == "1":
                    # Verbe à l'infinitif strict
                    verb_inf_freq[ortho] = verb_inf_freq.get(ortho, 0.0) + score
                else:
                    # Forme conjuguée de verbe
                    verb_conj_freq[ortho] = verb_conj_freq.get(ortho, 0.0) + score
            elif (cgram.startswith("NOM") or cgram.startswith("ADJ")) and islem == "1" and nomb != "p":
                # Nom ou Adjectif au singulier strict
                noun_adj_freq[ortho] = noun_adj_freq.get(ortho, 0.0) + score

    print(f"✅ Base Lexique analysée ({len(verb_inf_freq)} infinitifs, {len(noun_adj_freq)} noms/adj).")

    # 3. Sélection stricte des cibles familières (sans aucune forme verbale conjuguée)
    print("🔍 Sélection des mots cibles : infinitifs + noms/adjectifs familiers...")
    targets = set()

    # A. Verbes à l'infinitif
    for word, score in verb_inf_freq.items():
        if score >= MIN_SCORE and zipf_frequency(word, "fr") >= TARGET_MIN_ZIPF:
            w_upper = normalize_word(word)
            if w_upper and MIN_LEN <= len(w_upper) <= MAX_LEN:
                if w_upper in all_valid and w_upper not in blacklist:
                    targets.add(w_upper)

    # B. Noms et Adjectifs singuliers
    for word, score in noun_adj_freq.items():
        v_conj = verb_conj_freq.get(word, 0.0)
        # Règle anti-verbe conjugué : si le mot apparaît principalement ou significativement
        # comme un verbe conjugué (ex: arrive, mange, part, marche), on l'exclut formellement
        # des cibles pour éviter toute ambiguïté conjuguée !
        if v_conj > 0 and score < 2.0 * v_conj:
            continue

        if score >= MIN_SCORE and zipf_frequency(word, "fr") >= TARGET_MIN_ZIPF:
            w_upper = normalize_word(word)
            if w_upper and MIN_LEN <= len(w_upper) <= MAX_LEN:
                if w_upper in all_valid and w_upper not in blacklist:
                    # Exclusion supplémentaire des terminaisons verbales classiques
                    if not (w_upper.endswith("EZ") or w_upper.endswith("ONS") or w_upper.endswith("AIENT") or w_upper.endswith("AMES")):
                        targets.add(w_upper)

    print(f"🎯 {len(targets):,} mots cibles sélectionnés pour targets.txt.")

    # 4. Disjonction stricte pour allowed.txt
    allowed = all_valid - targets
    print(f"📖 {len(allowed):,} mots de validation additionnels pour allowed.txt.")

    # Validation d'intersection nulle
    intersection = targets.intersection(allowed)
    if intersection:
        raise ValueError(f"❌ Erreur critique : {len(intersection)} mots en doublon détectés !")
    print("✨ Validation stricte : intersection entre targets et allowed strictement nulle.")

    # 5. Écriture des fichiers
    targets_file = out_dir / "targets.txt"
    allowed_file = out_dir / "allowed.txt"

    print(f"💾 Écriture dans : {out_dir}")
    targets_file.write_text("\n".join(sorted(targets)) + "\n", encoding="utf-8")
    allowed_file.write_text("\n".join(sorted(allowed)) + "\n", encoding="utf-8")

    print(f"✅ Fichiers créés avec succès :")
    print(f"   - {targets_file} ({len(targets)} mots)")
    print(f"   - {allowed_file} ({len(allowed)} mots)")
    print("=" * 60)

if __name__ == "__main__":
    main()
