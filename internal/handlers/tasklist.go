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

func (h *TaskListHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	firebaseUID := auth.GetUserID(r.Context())

	rows, err := h.db.QueryContext(r.Context(), `SELECT tl.id, tl.user_id, tl.name, tl.created_at
		FROM task_lists tl
		JOIN users u ON u.id = tl.user_id
		WHERE u.firebase_uid = $1
		ORDER BY tl.created_at DESC`, firebaseUID)
	if err != nil {
		slog.Error("failed to get task lists", "error", err)
		http.Error(w, `{"error": "database error"}`, http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	taskLists := []TaskList{}
	for rows.Next() {
		var tl TaskList
		if err := rows.Scan(&tl.ID, &tl.UserID, &tl.Name, &tl.CreatedAt); err != nil {
			slog.Error("failed to scan task list", "error", err)
			continue
		}

		taskLists = append(taskLists, tl)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(taskLists)
}

func (h *TaskListHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	firebaseUID := auth.GetUserID(r.Context())
	taskListID := strings.TrimPrefix(r.URL.Path, "/api/tasklists/")

	var taskList TaskList
	err := h.db.QueryRowContext(r.Context(), `SELECT tl.id, tl.user_id, tl.name, tl.created_at
		FROM task_lists tl
		JOIN users u ON u.id = tl.user_id
		WHERE tl.id = $1 AND u.firebase_uid = $2`, taskListID, firebaseUID).Scan(&taskList.ID, &taskList.UserID, &taskList.Name, &taskList.CreatedAt)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error": "task list not found"}`, http.StatusNotFound)
		return
	}

	if err != nil {
		slog.Error("failed to get task list", "error", err)
		http.Error(w, `{"error": "database error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(taskList)
}

func (h *TaskListHandler) Delete(w http.ResponseWriter, r *http.Request) {
	firebaseUID := auth.GetUserID(r.Context())
	taskListID := strings.TrimPrefix(r.URL.Path, "/api/tasklists/")

	result, err := h.db.ExecContext(r.Context(), `DELETE FROM task_lists tl
		USING users u
		WHERE tl.user_id = u.id AND tl.id = $1 AND u.firebase_uid = $2`, taskListID, firebaseUID,
	)

	if err != nil {
		slog.Error("failed to delete task list", "error", err)
		http.Error(w, `{"error": "database error"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, `{"error": "task list not found"}`, http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
