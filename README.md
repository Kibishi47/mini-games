# 🎮 MiniGames - Plateforme Multijoueur Temps Réel (Wordle)

Plateforme web desktop de jeux multijoueur compétitifs en temps réel avec moteur **Server-Authoritative** ultra-robuste en **Go 1.23+**, interface réactive en **Nuxt 3**, et état temps réel **100% Redis 7**.

Le projet adopte la philosophie des jeux viraux instantanés (type Skribbl.io ou Codenames) : **zéro compte, zéro mot de passe, zéro base de données SQL**. Tout repose sur un mode Invité (Guest) persistant localement sur le client, un état temps réel ultra-rapide géré dans Redis, et un binaire Go compilé ultra-léger.

---

## 🌟 Fonctionnalités Clés

* **Zéro friction / 100% Mode Invité** :
  * Pseudo + choix parmi 6 mascottes vectorielles SVG animées (`dice`, `domino`, `card`, `knight`, `d20`, `meeple`).
  * Persistance locale via `localStorage`.
* **Moteur Server-Authoritative Wordle (3 à 8 lettres)** :
  * Dictionnaires français complets embarqués à la compilation via `//go:embed` (plus de 51 000 mots valides, 4 300 cibles usuelles du quotidien).
  * Validation $O(1)$ et tirage aléatoire par longueur.
  * **State Masking strict** : les tuiles des concurrents sont diffusées en direct sans les lettres (couleurs uniquement).
  * **Chronomètre absolu (`ends_at`)** : compte à rebours client précis insensible à la latence réseau.
  * **Partage viral émojis** (🟩🟨⬛) copiable en 1 clic.
* **Résilience & Reconnexion (Grace Period)** :
  * **Période de grâce de 45 secondes** : tolérance aux rechargements de page (F5) et micro-coupures réseau grâce aux sessions éphémères Redis avec TTL de 45s.
  * **Mode Spectateur automatique** : tout joueur arrivant en cours de manche observe et intègre automatiquement la manche suivante.
  * **Passation automatique de Master** : si le créateur quitte, le joueur le plus ancien hérite instantanément des privilèges.
  * **Modération Master** : exclusion (`kick`), bannissement (`ban`), et silence (`mute`).
  * **Boucle de Revanche (`Rematch`)** : retour au lobby en conservant les scores cumulés de la session.
* **Direction Artistique Pop Moderniste** :
  * Aplats francs, encrage noir épais, ombres portées dures décalées (*hard shadows*), zéro dégradé, zéro ombre floue.
* **Workers d'arrière-plan Go** :
  * Surveillance continue des heartbeats et timeout des joueurs déconnectés.
  * Garbage Collector de salles (fermeture automatique après 30 min d'inactivité ou 2 min sans joueur).

---

## 🛠️ Stack Technique

* **Backend** : Go 1.23+, routeur Chi, WebSocket (`nhooyr/websocket`), `go-redis/v9`.
* **Frontend** : Nuxt 3 (TypeScript strict, Nitro, Tailwind CSS, Pinia, Lucide Icons).
* **Base de données / Cache** : **Redis 7 uniquement** (aucun PostgreSQL).
* **DevOps** : Docker multi-stage (Alpine non-root), Docker Compose, Makefile.

---

## 🚀 Démarrage Rapide

### Prérequis
* Docker et Docker Compose
* Make

### 1. Lancer l'environnement de production en local (ou Coolify)
```bash
make up
```
Cette commande unique :
1. Démarre Redis 7, le Backend Go et le Frontend Nuxt.
2. Rend le frontend disponible sur **http://localhost:3000** et l'API sur **http://localhost:8080**.

### 2. Lancer l'environnement de développement avec Hot-Reload
```bash
make dev
```
* **Backend** : rechargé automatiquement à chaque modification Go via **Air**.
* **Frontend** : HMR instantané via Nuxt 3 Vite.

### Autres commandes Make utiles :
```bash
make down        # Stoppe et nettoie les conteneurs
make test        # Lance la suite de tests unitaires Go
make dictionary  # Régénère les dictionnaires français depuis Lexique 383
make logs        # Affiche les logs en continu
```

---

## 🌐 Déploiement Continu sur VPS avec Coolify

Le projet est nativement configuré pour **Coolify** via `docker-compose.yml` (3 services orchestrés : `frontend`, `backend`, `redis`) :

1. Sur votre instance Coolify, créez une nouvelle ressource **Docker Compose**.
2. Liez ce dépôt GitHub (branche `dev` ou `main`).
3. Renseignez les variables d'environnement dans l'interface Coolify (copiez depuis `.env.example`) :
   * `FRONTEND_URL` (votre URL publique de frontend)
   * `NUXT_PUBLIC_API_URL` (URL publique de l'API)
   * `NUXT_PUBLIC_WS_URL` (URL publique WebSocket)
4. Cliquez sur **Deploy** : Coolify compile les conteneurs, applique les healthchecks stricts et démarre les 3 services avec volume persistant `redis_data`.

---

## 🧪 Structure du Projet

```text
.
├── Makefile                     # Commandes unifiées (make up, make dev, etc.)
├── docker-compose.yml           # Déploiement production / Coolify (3 services)
├── docker-compose.dev.yml       # Environnement dev local avec Air et HMR
├── backend/
│   ├── Dockerfile               # Build multi-stage Go Alpine avec dictionnaires
│   ├── cmd/api/main.go          # Point d'entrée Chi, WS, Redis & Workers
│   ├── cmd/tools/dictionary/    # Générateur autonome de dictionnaire
│   └── internal/
│       ├── config/              # Configuration & variables d'environnement
│       ├── domain/              # Modèles métier & structures d'événements
│       ├── repository/redis/    # Implémentation Redis (rooms, sessions, chat)
│       ├── service/room/        # Logique de salon & gouvernance
│       ├── service/games/wordle/# Moteur Wordle (3-8 lettres, //go:embed)
│       ├── transport/http/      # Handlers REST légers
│       ├── transport/ws/        # Hub WebSocket, dispatch & heartbeat
│       └── worker/              # Workers d'inactivité et garbage collector
└── frontend/
    ├── Dockerfile               # Build multi-stage Nuxt 3 Nitro
    ├── pages/
    │   ├── index.vue            # Accueil, profil invité, créer/rejoindre
    │   └── room/[code].vue      # Vue unique de salle (Lobby, Jeu, Fin)
    ├── components/
    │   ├── ui/                  # AppButton, AppCard, AppBadge, AppInput, GameMascot
    │   ├── wordle/              # WordleGrid, OpponentPreview
    │   ├── chat/                # ChatPanel
    │   └── scoreboard/          # ScoreboardModal
    ├── stores/                  # Stores Pinia (profile, room, game)
    └── composables/             # useWebSocket
```
