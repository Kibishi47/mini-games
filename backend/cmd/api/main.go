package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"minigames-backend/internal/config"
	"minigames-backend/internal/repository/postgres"
	"minigames-backend/internal/repository/redis"
	"minigames-backend/internal/service/auth"
	"minigames-backend/internal/service/games/wordle"
	"minigames-backend/internal/service/room"
	transportHttp "minigames-backend/internal/transport/http"
	"minigames-backend/internal/transport/ws"
	"minigames-backend/internal/worker"
)

func main() {
	migrateFlag := flag.Bool("migrate", false, "Exécuter les migrations SQL et quitter")
	seedFlag := flag.Bool("seed", false, "Injecter les données de test et quitter")
	flag.Parse()

	cfg := config.Load()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialisation PostgreSQL
	pgPool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Échec connexion PostgreSQL: %v", err)
	}
	defer pgPool.Close()

	// Chemins relatifs / absolus pour migrations et dictionnaires
	migrationsDir := "migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		migrationsDir = filepath.Join("backend", "migrations")
	}


	// Mode Migration
	if *migrateFlag {
		log.Println("📦 Exécution des migrations...")
		if err := postgres.RunMigrations(ctx, pgPool, migrationsDir); err != nil {
			log.Fatalf("Échec migrations: %v", err)
		}
		log.Println("✅ Migrations terminées avec succès")
		return
	}

	// Application systématique des migrations au démarrage
	_ = postgres.RunMigrations(ctx, pgPool, migrationsDir)

	// Mode Seed
	if *seedFlag {
		log.Println("🌱 Seed des données de démonstration...")
		userRepo := postgres.NewUserRepository(pgPool)
		authService := auth.NewAuthService(cfg, userRepo)
		_, _ = authService.RegisterLocal(ctx, "admin", "Administrateur", "admin@minigames.local", "admin1234")
		_, _ = authService.RegisterLocal(ctx, "champion", "WordleMaster", "champ@minigames.local", "champion123")
		log.Println("✅ Seed terminé avec succès")
		return
	}

	// Initialisation Redis
	redisClient, err := redis.NewClient(ctx, cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		log.Fatalf("Échec connexion Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialisation Repositories
	userRepo := postgres.NewUserRepository(pgPool)
	roomSessionRepo := postgres.NewRoomSessionRepository(pgPool)
	roomRedisRepo := redis.NewRoomRepository(redisClient)

	// Initialisation Services
	authService := auth.NewAuthService(cfg, userRepo)
	roomService := room.NewRoomService(roomRedisRepo, roomSessionRepo, userRepo)

	// Initialisation Dictionnaire Wordle (moteur embarqué)
	dict := wordle.NewDictionary()

	// Initialisation WebSocket Hub & Gestionnaire de Jeu
	wsHub := ws.NewHub(roomRedisRepo, roomSessionRepo)
	_ = wordle.NewWordleGameManager(dict, wsHub, roomRedisRepo, roomSessionRepo, userRepo, redisClient)

	// Initialisation Workers
	workerMgr := worker.NewWorkerManager(roomRedisRepo, roomSessionRepo, wsHub)
	workerMgr.Start(ctx)

	// Handlers HTTP
	authHandler := transportHttp.NewAuthHandler(authService, userRepo)
	roomHandler := transportHttp.NewRoomHandler(roomService, wsHub, authService)

	// Routeur Chi
	r := chi.NewRouter()

	// Middlewares globaux
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(transportHttp.RecoverMiddleware)

	// CORS pour autoriser Nuxt HMR et requêtes cross-origin
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://127.0.0.1:3000", cfg.FrontendURL, "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","time":"` + time.Now().Format(time.RFC3339) + `"}`))
	})

	// WebSocket route
	r.Get("/ws", roomHandler.HandleWebSocket)

	// API Routes
	r.Route("/api", func(r chi.Router) {
		// Auth publique
		r.Post("/auth/guest", authHandler.GuestLogin)
		r.Post("/auth/register", authHandler.RegisterLocal)
		r.Post("/auth/login", authHandler.LoginLocal)
		r.Get("/auth/discord/login", authHandler.DiscordAuthURL)
		r.Get("/auth/discord/callback", authHandler.DiscordCallback)

		// Routes protégées par JWT
		r.Group(func(r chi.Router) {
			r.Use(transportHttp.AuthMiddleware(authService))

			// User
			r.Get("/users/me", authHandler.GetMe)
			r.Put("/users/profile", authHandler.UpdateProfile)
			r.Post("/users/upgrade-guest", authHandler.UpgradeGuest)

			// Rooms
			r.Post("/rooms", roomHandler.CreateRoom)
			r.Get("/rooms/{code}", roomHandler.GetRoom)
			r.Post("/rooms/{code}/join", roomHandler.JoinRoom)
		})
	})

	// Serveur HTTP
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.AppPort),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("🚀 Serveur backend démarré sur :%s (Env: %s)", cfg.AppPort, cfg.AppEnv)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erreur serveur HTTP: %v", err)
		}
	}()

	// Arrêt gracieux (Graceful Shutdown)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Fermeture du serveur en cours...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Erreur lors du shutdown: %v", err)
	}
	log.Println("👋 Serveur arrêté proprement")
}
