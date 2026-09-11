.PHONY: dev down restart migrate seed logs test clean

# Variables
DC_DEV = docker compose -f docker-compose.dev.yml

# Charge les variables du .env local si présent (docker compose le lit déjà nativement pour
# ses propres ${VAR} ; on le reflète ici pour que les messages ci-dessous restent exacts)
-include .env
export

APP_PORT ?= 8080
FRONTEND_PORT ?= 3000

# Lance l'environnement de développement (Redis Docker + Go Air + Nuxt dev)
dev:
	@echo "🛠️ Démarrage de l'environnement de développement..."
	@echo "🌐 Frontend : http://localhost:$(FRONTEND_PORT)"
	@echo "⚙️ Backend API : http://localhost:$(APP_PORT)"
	$(DC_DEV) up --build

# Arrête tous les conteneurs
down:
	@echo "🛑 Arrêt des services..."
	$(DC_DEV) down -v --remove-orphans

# Affiche les logs consolidés
logs:
	$(DC_DEV) logs -f

# Lance les tests unitaires et intégration
test:
	@echo "🧪 Exécution des tests Go..."
	cd backend && go test -v -race ./...

# Génération et compilation des dictionnaires français pour Wordle
DICTIONARY_DIR = backend/internal/service/games/wordle/dictionary/fr
PYTHON := $(shell which python3.11 2>/dev/null || which python3 2>/dev/null || echo python3)

.PHONY: dictionary
dictionary:
	@if [ ! -f $(DICTIONARY_DIR)/targets.txt ] || [ ! -f $(DICTIONARY_DIR)/allowed.txt ]; then \
		echo "Dictionnaires absents. Génération via Python..."; \
		$(PYTHON) -m pip install -r scripts/requirements.txt; \
		$(PYTHON) scripts/build_dictionary.py; \
	else \
		echo "Dictionnaires déjà présents. Pour forcer la régénération : make force-dictionary"; \
	fi

.PHONY: force-dictionary
force-dictionary:
	@$(PYTHON) -m pip install -r scripts/requirements.txt
	@$(PYTHON) scripts/build_dictionary.py

.PHONY: build-api
build-api: dictionary
	@cd backend && go build -o ../bin/api ./cmd/api/main.go

# Nettoyage des artefacts
clean:
	@echo "🧹 Nettoyage..."
	rm -rf backend/tmp frontend/.nuxt frontend/.output frontend/node_modules bin/api
