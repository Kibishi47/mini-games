# Étape 1 : Génération des dictionnaires via Python
FROM python:3.11-slim AS dict-builder

WORKDIR /app
COPY scripts/requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

COPY scripts/build_dictionary.py ./
RUN python build_dictionary.py

# Étape 2 : Compilation du binaire Go
FROM golang:1.23-alpine AS go-builder

WORKDIR /app

# Dépendances Go
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copie du code source backend
COPY backend/ ./

# Injection des dictionnaires générés par l'étape Python
COPY --from=dict-builder /app/backend/internal/service/games/wordle/dictionary/fr/*.txt ./internal/service/games/wordle/dictionary/fr/

# Compilation statique
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /api ./cmd/api/main.go

# Étape 3 : Image d'exécution minimale
FROM alpine:3.20

WORKDIR /app
COPY --from=go-builder /api /app/api

EXPOSE 8080
USER nobody:nobody
ENTRYPOINT ["/app/api"]
