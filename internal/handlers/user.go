package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/0xhop3/outils/internal/auth"
)

const (
	GET_USER = `SELECT id, firebase_uid, email, display_name, created_at, updated_at
		FROM users
		WHERE firebase_uid = $1`
	CREATE_USER = `INSERT INTO users (firebase_uid, email, display_name) VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`
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
	CreatedAt   time.Time `json:"created_at"`
	DisplayName string    `json:"display_name"`
}

type RegisterRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	firebaseUID := auth.GetUserID(r.Context())
}
