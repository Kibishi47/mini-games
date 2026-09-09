.PHONY: up dev down restart migrate seed logs test clean

# Variables
DC_DEV = docker compose -f docker-compose.dev.yml
DC_PROD = docker compose -f docker-compose.yml

# Lance tout l'environnement de production en conteneurs
up:
	@echo "🚀 Démarrage de la stack de production..."
	$(DC_PROD) up -d --build
	@echo "⏳ En attente de la base de données..."
	@sleep 3
	@echo "📦 Application des migrations..."
	$(DC_PROD) exec -T backend ./api -migrate
	@echo "🌱 Injection des données de seed (dictionnaires & users)..."
	$(DC_PROD) exec -T backend ./api -seed
	@echo "✅ Plateforme prête !"
	@echo "🌐 Frontend : http://localhost:3000"
	@echo "⚙️ Backend API : http://localhost:8080"

# Lance l'environnement de développement (Postgres/Redis Docker + Go Air + Nuxt dev)
dev:
	@echo "🛠️ Démarrage de l'environnement de développement..."
	$(DC_DEV) up --build

# Arrête tous les conteneurs
down:
	@echo "🛑 Arrêt des services..."
	$(DC_DEV) down -v --remove-orphans
	$(DC_PROD) down -v --remove-orphans

# Applique les migrations SQL
migrate:
	@echo "📦 Application des migrations PostgreSQL..."
	$(DC_DEV) exec -T backend-dev ./api -migrate || go run ./backend/cmd/api/main.go -migrate

# Injecte les données de test / dictionnaires
seed:
	@echo "🌱 Injection des données de seed..."
	$(DC_DEV) exec -T backend-dev ./api -seed || go run ./backend/cmd/api/main.go -seed

# Affiche les logs consolidés
logs:
	$(DC_DEV) logs -f

# Lance les tests unitaires et intégration
test:
	@echo "🧪 Exécution des tests Go..."
	cd backend && go test -v -race ./...

# Nettoyage des artefacts
clean:
	@echo "🧹 Nettoyage..."
	rm -rf backend/tmp frontend/.nuxt frontend/.output frontend/node_modules
