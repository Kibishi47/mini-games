.PHONY: up dev down restart migrate seed logs test clean

# Variables
DC_DEV = docker compose -f docker-compose.dev.yml
DC_PROD = docker compose -f docker-compose.yml

# Lance tout l'environnement de production en conteneurs
up:
	@echo "🚀 Démarrage de la stack de production (Coolify / Docker)..."
	$(DC_PROD) up -d --build
	@echo "✅ Plateforme prête !"
	@echo "🌐 Frontend : http://localhost:3000"
	@echo "⚙️ Backend API : http://localhost:8080"

# Lance l'environnement de développement (Redis Docker + Go Air + Nuxt dev)
dev:
	@echo "🛠️ Démarrage de l'environnement de développement..."
	$(DC_DEV) up --build

# Arrête tous les conteneurs
down:
	@echo "🛑 Arrêt des services..."
	$(DC_DEV) down -v --remove-orphans
	$(DC_PROD) down -v --remove-orphans

# Affiche les logs consolidés
logs:
	$(DC_DEV) logs -f

# Lance les tests unitaires et intégration
test:
	@echo "🧪 Exécution des tests Go..."
	cd backend && go test -v -race ./...

# Génération et compilation des dictionnaires français pour Wordle
dictionary:
	@echo "Téléchargement et compilation des dictionnaires français..."
	@cd backend && go run ./cmd/tools/dictionary/main.go

# Nettoyage des artefacts
clean:
	@echo "🧹 Nettoyage..."
	rm -rf backend/tmp frontend/.nuxt frontend/.output frontend/node_modules
