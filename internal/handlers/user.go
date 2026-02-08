package handlers

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/0xhop3/outils/internal/auth"
)

type UserHandler struct {
	db *sql.DB
}

func NewUserHandler(db *sql.DB) *UserHandler {
	return &UserHandler{db: db}
}

type User struct {
	ID          string    `json:"id"`
	FirebaseUID string    `json:"firebase_uid"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
}

type RegisterRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	firebaseUID := auth.GetUserID(r.Context())
	slog.Info("register called", "firebaseUID", firebaseUID)

	if firebaseUID == "" {
		slog.Error("firebaseUID is empty")
		http.Error(w, `{"error":"no user id in context"}`, http.StatusUnauthorized)
		return
	}

	// Check if user exists
	var user User
	var displayName sql.NullString

	err := h.db.QueryRowContext(r.Context(),
		`SELECT id, firebase_uid, email, display_name, created_at 
		 FROM users WHERE firebase_uid = $1`,
		firebaseUID,
	).Scan(&user.ID, &user.FirebaseUID, &user.Email, &displayName, &user.CreatedAt)

	if err == nil {
		slog.Info("user exists", "id", user.ID)
		user.DisplayName = displayName.String
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
		return
	}

	if err != sql.ErrNoRows {
		slog.Error("database query error", "error", err)
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}

	slog.Info("user not found, creating new user")

	// Parse request
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("invalid request body", "error", err)
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	slog.Info("creating user", "email", req.Email, "displayName", req.DisplayName)

	// Create user
	err = h.db.QueryRowContext(r.Context(),
		`INSERT INTO users (firebase_uid, email, display_name) 
		 VALUES ($1, $2, $3)
		 RETURNING id, firebase_uid, email, display_name, created_at`,
		firebaseUID, req.Email, req.DisplayName,
	).Scan(&user.ID, &user.FirebaseUID, &user.Email, &displayName, &user.CreatedAt)

	if err != nil {
		slog.Error("insert failed", "error", err)
		http.Error(w, `{"error":"failed to create user"}`, http.StatusInternalServerError)
		return
	}

	user.DisplayName = displayName.String
	slog.Info("user created", "id", user.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	firebaseUID := auth.GetUserID(r.Context())
	slog.Info("getme called", "firebaseUID", firebaseUID)

	var user User
	var displayName sql.NullString

	err := h.db.QueryRowContext(r.Context(),
		`SELECT id, firebase_uid, email, display_name, created_at 
		 FROM users WHERE firebase_uid = $1`,
		firebaseUID,
	).Scan(&user.ID, &user.FirebaseUID, &user.Email, &displayName, &user.CreatedAt)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("database error", "error", err)
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}

	user.DisplayName = displayName.String

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
