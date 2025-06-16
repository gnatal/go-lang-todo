package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"todo-api/internal/database"
	"todo-api/internal/models"

	"github.com/google/uuid"
)

type TodoHandler struct {
	db *database.DB
}

func NewTodoHandler(db *database.DB) *TodoHandler {
	return &TodoHandler{db: db}
}

func (h *TodoHandler) GetTodos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := `SELECT id, todo, completed, created_at, updated_at FROM todos ORDER BY created_at DESC`
	rows, err := h.db.Query(query)
	if err != nil {
		log.Printf("Error querying todos: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []models.Todo
	for rows.Next() {
		var todo models.Todo
		err := rows.Scan(&todo.ID, &todo.Todo, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning todo: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		todos = append(todos, todo)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating todos: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(todos); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *TodoHandler) GetTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := h.extractIDFromPath(r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	query := `SELECT id, todo, completed, created_at, updated_at FROM todos WHERE id = $1`
	var todo models.Todo
	err = h.db.QueryRow(query, id).Scan(&todo.ID, &todo.Todo, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}
		log.Printf("Error querying todo: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(todo); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Todo) == "" {
		http.Error(w, "Todo text is required", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO todos (todo) VALUES ($1) RETURNING id, todo, completed, created_at, updated_at`
	var todo models.Todo
	err := h.db.QueryRow(query, req.Todo).Scan(&todo.ID, &todo.Todo, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
	if err != nil {
		log.Printf("Error creating todo: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(todo); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func (h *TodoHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := h.extractIDFromPath(r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	setParts := []string{}
	args := []interface{}{}
	argCount := 1

	if req.Todo != nil {
		if strings.TrimSpace(*req.Todo) == "" {
			http.Error(w, "Todo text cannot be empty", http.StatusBadRequest)
			return
		}
		setParts = append(setParts, fmt.Sprintf("todo = $%d", argCount))
		args = append(args, *req.Todo)
		argCount++
	}

	if req.Completed != nil {
		setParts = append(setParts, fmt.Sprintf("completed = $%d", argCount))
		args = append(args, *req.Completed)
		argCount++
	}

	if len(setParts) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	query := fmt.Sprintf("UPDATE todos SET %s WHERE id = $%d RETURNING id, todo, completed, created_at, updated_at",
		strings.Join(setParts, ", "), argCount)
	args = append(args, id)

	var todo models.Todo
	err = h.db.QueryRow(query, args...).Scan(&todo.ID, &todo.Todo, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}
		log.Printf("Error updating todo: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(todo); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func (h *TodoHandler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := h.extractIDFromPath(r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM todos WHERE id = $1`
	result, err := h.db.Exec(query, id)
	if err != nil {
		log.Printf("Error deleting todo: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TodoHandler) extractIDFromPath(path string) (uuid.UUID, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return uuid.Nil, fmt.Errorf("invalid path")
	}

	return uuid.Parse(parts[len(parts)-1])
}
