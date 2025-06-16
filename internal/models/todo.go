package models

import (
	"time"

	"github.com/google/uuid"
)

type Todo struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Todo      string    `json:"todo" db:"todo"`
	Completed bool      `json:"completed" db:"completed"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreateTodoRequest struct {
	Todo string `json:"todo"`
}

type UpdateTodoRequest struct {
	Todo      *string `json:"todo,omitempty"`
	Completed *bool   `json:"completed,omitempty"`
}
