package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/0xhop3/outils/internal/auth"
)

const (
	GET_USER    = `SELECT id, firebase_uid, email, display_name, created_at, updated_at FROM users WHERE firebase_uid = $1`
	CREATE_USER = `INSERT INTO users (firebase_uid, email, display_name) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
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

	var user User
	err := h.db.QueryRowContext(r.Context(), GET_USER, firebaseUID).Scan(&user.ID, &user.FirebaseUID, &user.Email, &user.DisplayName, &user.CreatedAt)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
		return
	}

	if err != sql.ErrNoRows {
		http.Error(w, `{"error": "database error"}`, http.StatusInternalServerError)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request"}`, http.StatusBadRequest)
		return
	}

	err = h.db.QueryRowContext(r.Context(), CREATE_USER, firebaseUID, req.Email, req.DisplayName).Scan(&user.ID, &user.FirebaseUID, &user.Email, &user.DisplayName, &user.CreatedAt)

	if err != nil {
		http.Error(w, `{"error": "failed to create user"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	firebaseUID := auth.GetUserID(r.Context())

	var user User
	err := h.db.QueryRowContext(r.Context(), GET_USER, firebaseUID).Scan(&user.ID, &user.FirebaseUID, &user.Email, &user.DisplayName, &user.CreatedAt)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error": "user not found"}`, http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, `{"error": "database error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
