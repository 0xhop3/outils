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
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config error", "error", err)
		os.Exit(1)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("database error", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("Connected to database")

	firebaseAuth, err := auth.NewFirebaseAuthentication(cfg.FirebaseCredentialsFile)
	if err != nil {
		slog.Error("firebase error", "error", err)
		os.Exit(1)
	}
	slog.Info("firebase initialized")

	// Handlers
	healthHandler := handlers.NewHealthHandler(db)
	userHandler := handlers.NewUserHandler(db)
	taskListHandler := handlers.NewTaskListHandler(db)
	taskHandler := handlers.NewTaskHandler(db)

	mux := http.NewServeMux()

	// Public
	mux.HandleFunc("GET /health", healthHandler.Check)
	mux.HandleFunc("GET /ready", healthHandler.Ready)

	// User
	mux.Handle("POST /api/register", firebaseAuth.Middleware(http.HandlerFunc(userHandler.Register)))
	mux.Handle("GET /api/me", firebaseAuth.Middleware(http.HandlerFunc(userHandler.GetMe)))

	// Task Lists
	mux.Handle("POST /api/tasklists", firebaseAuth.Middleware(http.HandlerFunc(taskListHandler.Create)))
	mux.Handle("GET /api/tasklists", firebaseAuth.Middleware(http.HandlerFunc(taskListHandler.GetAll)))
	mux.Handle("GET /api/tasklist/{id}", firebaseAuth.Middleware(http.HandlerFunc(taskListHandler.GetOne)))
	mux.Handle("DELETE /api/tasklist/{id}", firebaseAuth.Middleware(http.HandlerFunc(taskListHandler.Delete)))

	// Tasks
	mux.Handle("POST /api/tasklist/{id}/tasks", firebaseAuth.Middleware(http.HandlerFunc(taskHandler.Create)))
	mux.Handle("GET /api/tasklist/{id}/tasks", firebaseAuth.Middleware(http.HandlerFunc(taskHandler.GetAll)))
	mux.Handle("PUT /api/tasks/{id}", firebaseAuth.Middleware(http.HandlerFunc(taskHandler.Update)))
	mux.Handle("DELETE /api/tasks/{id}", firebaseAuth.Middleware(http.HandlerFunc(taskHandler.Delete)))

	handler := corsMiddleware(mux)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		slog.Info("starting server", "port", cfg.Port)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	server.Shutdown(ctx)
	slog.Info("server stopped")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for ALL requests
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
}
