package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/0xhop3/outils/internal/auth"
)

type TaskHandler struct {
	db *sql.DB
}

func NewTaskHandler(db *sql.DB) *TaskHandler {
	return &TaskHandler{db: db}
}

type Task struct {
	ID         string    `json:"id"`
	TaskListID string    `json:"task_list_id"`
	Title      string    `json:"title"`
	Completed  bool      `json:"completed"`
	CreatedAt  time.Time `json:"created_at"`
}

type CreateTaskRequest struct {
	Title string `json:"title"`
}

type UpdateTaskRequest struct {
	Title     *string `json:"title,omitempty"`
	Completed *bool   `json:"completed,omitempty"`
}

func (h *TaskHandler) verifyTaskListOwnership(ctx context.Context, taskListID, firebaseUID string) bool {
	var exists bool
	err := h.db.QueryRowContext(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM task_lists tl
			JOIN users u ON u.id = tl.user_id
			WHERE tl.id = $1 AND u.firebase_uid = $2
		)`,
		taskListID, firebaseUID,
	).Scan(&exists)
	return err == nil && exists
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	firebaseUID := auth.GetUserID(r.Context())
	taskListID := r.PathValue("id")

	if !h.verifyTaskListOwnership(r.Context(), taskListID, firebaseUID) {
		http.Error(w, `{"error":"task list not found"}`, http.StatusNotFound)
		return
	}

	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, `{"error":"title is required"}`, http.StatusBadRequest)
		return
	}

	var task Task
	err := h.db.QueryRowContext(r.Context(),
		`INSERT INTO tasks (task_list_id, title) VALUES ($1, $2)
		 RETURNING id, task_list_id, title, completed, created_at`,
		taskListID, req.Title,
	).Scan(&task.ID, &task.TaskListID, &task.Title, &task.Completed, &task.CreatedAt)

	if err != nil {
		slog.Error("failed to create task", "error", err)
		http.Error(w, `{"error":"failed to create task"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	firebaseUID := auth.GetUserID(r.Context())
	taskListID := r.PathValue("id")

	if !h.verifyTaskListOwnership(r.Context(), taskListID, firebaseUID) {
		http.Error(w, `{"error":"task list not found"}`, http.StatusNotFound)
		return
	}

	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, task_list_id, title, completed, created_at
		 FROM tasks WHERE task_list_id = $1
		 ORDER BY created_at ASC`,
		taskListID,
	)
	if err != nil {
		slog.Error("failed to get tasks", "error", err)
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.TaskListID, &t.Title, &t.Completed, &t.CreatedAt); err != nil {
			continue
		}
		tasks = append(tasks, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	firebaseUID := auth.GetUserID(r.Context())
	taskID := r.PathValue("id")

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	var task Task
	err := h.db.QueryRowContext(r.Context(),
		`UPDATE tasks t
		 SET title = COALESCE($1, t.title),
		     completed = COALESCE($2, t.completed)
		 FROM task_lists tl
		 JOIN users u ON u.id = tl.user_id
		 WHERE t.id = $3 AND t.task_list_id = tl.id AND u.firebase_uid = $4
		 RETURNING t.id, t.task_list_id, t.title, t.completed, t.created_at`,
		req.Title, req.Completed, taskID, firebaseUID,
	).Scan(&task.ID, &task.TaskListID, &task.Title, &task.Completed, &task.CreatedAt)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"task not found"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("failed to update task", "error", err)
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	firebaseUID := auth.GetUserID(r.Context())
	taskID := r.PathValue("id")

	result, err := h.db.ExecContext(r.Context(),
		`DELETE FROM tasks t
		 USING task_lists tl, users u
		 WHERE t.task_list_id = tl.id AND tl.user_id = u.id
		 AND t.id = $1 AND u.firebase_uid = $2`,
		taskID, firebaseUID,
	)
	if err != nil {
		slog.Error("failed to delete task", "error", err)
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, `{"error":"task not found"}`, http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
