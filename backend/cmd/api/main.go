package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"minigames-backend/internal/config"
	"minigames-backend/internal/repository/redis"
	"minigames-backend/internal/service/games/wordle"
	"minigames-backend/internal/service/room"
	transportHttp "minigames-backend/internal/transport/http"
	"minigames-backend/internal/transport/ws"
	"minigames-backend/internal/worker"
)

func main() {
	cfg := config.Load()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialisation Redis 7
	redisClient, err := redis.NewClient(ctx, cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		log.Fatalf("❌ Échec connexion Redis (%s): %v", cfg.RedisAddr, err)
	}
	defer redisClient.Close()
	log.Printf("✅ Connecté à Redis sur %s", cfg.RedisAddr)

	// Initialisation Repositories
	roomRepo := redis.NewRoomRepository(redisClient)

	// Initialisation Services
	roomService := room.NewRoomService(roomRepo)

	// Initialisation Dictionnaire Wordle embarqué (//go:embed targets.txt & allowed.txt)
	dict := wordle.NewDictionary()
	log.Printf("📚 Dictionnaire Wordle initialisé : %d cibles canoniques, %d mots autorisés", dict.TargetsCount(), dict.AllowedCount())

	// Initialisation WebSocket Hub & Gestionnaire de Jeu Wordle
	wsHub := ws.NewHub(roomRepo)
	_ = wordle.NewWordleGameManager(dict, wsHub, roomRepo, redisClient)

	// Initialisation Workers d'inactivité AFK et GC de salles
	workerMgr := worker.NewWorkerManager(roomRepo, wsHub)
	workerMgr.Start(ctx)

	// Handlers HTTP
	roomHandler := transportHttp.NewRoomHandler(roomService, roomRepo, wsHub)

	// Routeur Chi
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(60 * time.Second))

	// Configuration CORS permissif pour Nuxt 3 (Localhost & Production)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://127.0.0.1:3000", cfg.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Session-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Endpoint Healthcheck
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","engine":"redis","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	})

	// Routes REST Salles
	r.Route("/api/rooms", func(r chi.Router) {
		r.Post("/", roomHandler.CreateRoom)
		r.Get("/{code}", roomHandler.GetRoom)
		r.Post("/{code}/join", roomHandler.JoinRoom)
	})

	// Endpoint WebSocket
	r.Get("/ws", roomHandler.HandleWebSocket)

	// Démarrage Serveur HTTP
	serverAddr := fmt.Sprintf(":%s", cfg.AppPort)
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("🚀 Serveur MiniGames démarré sur le port %s (Env: %s)", cfg.AppPort, cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erreur serveur HTTP: %v", err)
		}
	}()

	// Arrêt gracieux (Graceful Shutdown)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Arrêt du serveur en cours...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Forçage de l'arrêt: %v", err)
	}

	log.Println("👋 Serveur arrêté proprement")
}
