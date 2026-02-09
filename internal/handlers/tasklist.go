package handlers

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/0xhop3/outils/internal/auth"
)

type TaskListHandler struct {
	db *sql.DB
}

func NewTaskListHandler(db *sql.DB) *TaskListHandler {
	return &TaskListHandler{db: db}
}

type TaskList struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateTaskListRequest struct {
	Name string `json:"name"`
}

func (h *TaskListHandler) Create(w http.ResponseWriter, r *http.Request) {
	firebaseUID := auth.GetUserID(r.Context())

	var userID string
	err := h.db.QueryRowContext(r.Context(), `SELECT id FROM where firebase_uid = $1`, firebaseUID).Scan(&userID)
	if err != nil {
		slog.Error("user not found", "error", err)
		http.Error(w, `{"error": "user not found"}`, http.StatusNotFound)
		return
	}

	var request CreateTaskListRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, `{"error": "invalid request"}`, http.StatusBadRequest)
		return
	}

	if request.Name == "" {
		http.Error(w, `{"error": "name is required"}`, http.StatusBadRequest)
		return
	}

	var taskList TaskList
	err = h.db.QueryRowContext(r.Context(),
		`INSERT INTO task_list (user_id, name) VALUES ($1, $2)
		RETURNING id, user_id, name, created_at`,
		userID, request.Name).Scan(&taskList.ID, &taskList.UserID, &taskList.Name, &taskList.CreatedAt)

	if err != nil {
		slog.Error("failed to create task list", "error", err)
		http.Error(w, `{"error": "failed to create task list"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(taskList)
}
