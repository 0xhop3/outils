package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/0xhop3/outils/internal/auth"
	"github.com/0xhop3/outils/internal/config"
	"github.com/0xhop3/outils/internal/database"
	"github.com/0xhop3/outils/internal/handlers"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	config, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	db, err := database.Connect(config.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect with database", "error", err)
		os.Exit(1)
	}

	defer db.Close()

	slog.Info("Connected to database")

	firebaseAuthentication, err := auth.NewFirebaseAuthentication(config.FirebaseCredentialsFile)
	if err != nil {
		slog.Error("firebase error", "error", err)
		os.Exit(1)
	}

	slog.Info("firebase initialized")

	healthHandler := handlers.NewHealthHandler(db)
	userHandler := handlers.NewUserHandler(db)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler.Check)
	mux.HandleFunc("GET /ready", healthHandler.Ready)
	mux.Handle("POST /api/register", firebaseAuthentication.Middleware(http.HandlerFunc(userHandler.Register)))
	mux.Handle("GET /api/me", firebaseAuthentication.Middleware(http.HandlerFunc(userHandler.GetMe)))

	handler := corsMiddleware(mux)

	server := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("starting server", "port", config.Port)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting server down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped gracefully")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
