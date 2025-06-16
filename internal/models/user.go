package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreateUserRequest struct {
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
}

type UpdateUserRequest struct {
	Password *string `json:"password" db:"password"`
}

type LoginUserRequest struct {
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
}

type LoginUserResponse struct {
	Token string `json:"token"`
}
