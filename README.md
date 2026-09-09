# 🎮 MiniGames - Plateforme Multijoueur Temps Réel (Wordle)

Plateforme web desktop de jeux multijoueur compétitifs en temps réel avec moteur **Server-Authoritative** ultra-robuste en **Go 1.23+**, interface réactive en **Nuxt 3**, et persistance hybride **PostgreSQL + Redis**.

Développé selon les standards Staff Engineer : machine à états stricte, pas de ready-check (autorité Master), state-masking des tuiles adverses, grace period de 45s en cas de coupure réseau, et orchestration unifiée `make`.

---

## 🌟 Fonctionnalités Clés

* **Moteur Server-Authoritative** :
  * Évaluation officielle Wordle avec gestion exacte des doublons.
  * **State Masking** : la solution reste secrète côté serveur ; les autres joueurs reçoivent les tuiles colorées en direct sans les lettres.
  * **Chronomètre absolu (`ends_at`)** : calcul côté client pour neutraliser la latence réseau.
  * **Scoreboard de session & Revanche (`Rematch`)** : cumul des points manche par manche et relance sans perte d'historique.
  * **Partage viral émojis** (🟩🟨⬛) copiable en 1 clic.
* **Résilience & Temps Réel** :
  * **Grace Period de 45 secondes** : tolérance aux micro-coupures et aux rechargements de page (F5) sans éjection ni perte de score.
  * **Mode Spectateur automatique** : tout joueur arrivant en cours de partie entre en spectateur passif et intègre la manche suivante.
  * **Passation automatique de Master** : si le créateur quitte, le joueur le plus ancien hérite instantanément des privilèges.
  * **Modération Master** : kick, ban temporaire (30 min) et mute d'un joueur, avec interdiction stricte de se cibler soi-même.
* **Authentification Flexible** :
  * **Mode Invité (Guest) 1-clic** avec pseudonyme automatique ou personnalisé.
  * **Compte Local** sécurisé avec sel et hachage **Argon2id**.
  * **Discord OAuth2** prêt à l'emploi.
  * Conversion 1-clic d'un compte invité en compte permanent.
* **Workers d'arrière-plan Go** :
  * Détecteur d'inactivité AFK (timeout et délai de grâce).
  * Garbage Collector de salles (fermeture et purge après 30 min d'inactivité ou 2 min sans joueur).

---

## 🛠️ Stack Technique

* **Backend** : Go 1.23+, routeur Chi, WebSocket (`nhooyr/websocket`), `pgx/v5`, `go-redis/v9`, `argon2`, `golang-jwt`.
* **Frontend** : Nuxt 3 (TypeScript strict, Nitro, Tailwind CSS, Pinia, Lucide Icons).
* **Bases de données** : PostgreSQL 16 & Redis 7.
* **DevOps** : Docker multi-stage (Alpine/Distroless non-root), Docker Compose, Air (rechargement à chaud Go), Makefile.

---

## 🚀 Démarrage Rapide

### Prérequis
* Docker et Docker Compose
* Make

### 1. Lancer l'environnement de production en local
```bash
make up
```
Cette commande unique :
1. Démarre PostgreSQL, Redis, le Backend Go et le Frontend Nuxt.
2. Applique automatiquement les migrations SQL PostgreSQL.
3. Injecte les dictionnaires et utilisateurs de démonstration (`admin` et `champion`).
4. Rend le frontend disponible sur **http://localhost:3000** et l'API sur **http://localhost:8080**.

### 2. Lancer l'environnement de développement avec Hot-Reload
```bash
make dev
```
* **Backend** : rechargé automatiquement à chaque modification Go via **Air**.
* **Frontend** : HMR instantané via Nuxt 3 Vite.

### Autres commandes Make utiles :
```bash
make down     # Stoppe et nettoie les conteneurs
make migrate  # Exécute les migrations de schéma SQL
make seed     # Injecte les données de test
make test     # Lance la suite de tests unitaires Go
make logs     # Affiche les logs en continu
```

---

## 🌐 Déploiement Continu sur VPS avec Coolify

Le projet est nativement conçu pour un déploiement zero-friction sur **Coolify** via `docker-compose.yml` :

1. Sur votre instance Coolify, créez un nouveau projet et sélectionnez **Docker Compose**.
2. Liez ce dépôt GitHub.
3. Renseignez les variables d'environnement dans l'interface Coolify (copiez depuis `.env.example`) :
   * `POSTGRES_PASSWORD`
   * `JWT_SECRET` (clé sécurisée de 32 caractères minimum)
   * `FRONTEND_URL` (votre nom de domaine public)
   * `DISCORD_CLIENT_ID` et `DISCORD_CLIENT_SECRET` (optionnels)
4. Cliquez sur **Deploy** : Coolify compile les deux Dockerfiles multi-stage, vérifie les healthchecks stricts et démarre la stack de manière isolée avec volumes persistants `pg_data` et `redis_data`.

---

## 🧪 Structure du Projet

```
.
├── Makefile                     # Commandes unifiées (make up, make dev, etc.)
├── docker-compose.yml           # Déploiement production / Coolify
├── docker-compose.dev.yml       # Environnement dev local avec Air et HMR
├── backend/
│   ├── Dockerfile               # Build multi-stage Go Alpine léger
│   ├── cmd/api/main.go          # Point d'entrée Chi, WS, DBs & Workers
│   ├── internal/
│   │   ├── config/              # Configuration env
│   │   ├── domain/              # Entités User, Room, GameState, WSMessage
│   │   ├── repository/postgres/ # Pools pgx/v5 & migrations
│   │   ├── repository/redis/    # Clés Redis, présences, chat & ratelimit
│   │   ├── service/auth/        # Argon2id, JWT, Invité & Discord
│   │   ├── service/room/        # FSM Lobby/Game/Closed, Master passation
│   │   ├── service/games/wordle/# Dictionnaires FR/EN, State-Masking, Evaluator
│   │   ├── transport/http/      # Handlers Chi & Middlewares
│   │   ├── transport/ws/        # Hub WebSocket, Client & Dispatcher
│   │   └── worker/              # Worker AFK & Inactivity GC
│   └── migrations/              # Schémas PostgreSQL versionnés
└── frontend/
    ├── Dockerfile               # Build multi-stage Node / Nitro production
    ├── pages/                   # Index, Lobby, Room
    ├── stores/                  # Pinia auth, room, game
    ├── components/              # ChatPanel, WordleGrid, OpponentPreview, ScoreboardModal
    └── composables/             # useWebSocket avec reconnexion automatique
```
