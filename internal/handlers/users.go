package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"todo-api/internal/database"
	"todo-api/internal/models"
	"todo-api/internal/utils"

	"github.com/google/uuid"
)

type UserHandler struct {
	db *database.DB
}

func NewUserHandler(db *database.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (userHander *UserHandler) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(body.Email) == "" || strings.TrimSpace(body.Password) == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}
	hashedPassword, err := utils.HashPassword(body.Password)

	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	checkUserQuery := "SELECT id, email, password, created_at, updated_at FROM users WHERE email = $1"
	var user models.User
	err = userHander.db.QueryRow(checkUserQuery, body.Email).Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err == nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	if user.ID != uuid.Nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	query := "INSERT INTO users (email, password) VALUES($1, $2) RETURNING id"

	var userID uuid.UUID

	err = userHander.db.QueryRow(query, body.Email, hashedPassword).Scan(&userID)
	if err != nil {
		fmt.Printf("Error creating user: %v", err)
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	jwt, err := utils.GenerateJWT(userID, body.Email)
	if err != nil {
		http.Error(w, "Error generating JWT", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(jwt); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
	return
}

func (userHandler *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body models.LoginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(body.Email) == "" || strings.TrimSpace(body.Password) == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	var User models.User
	err := userHandler.db.QueryRow("SELECT id, email, password, created_at, updated_at FROM users WHERE email = $1", body.Email).Scan(&User.ID, &User.Email, &User.Password, &User.CreatedAt, &User.UpdatedAt)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusUnauthorized)
		return
	}

	if User.ID == uuid.Nil {
		http.Error(w, "Something went wrong", http.StatusUnauthorized)
		return
	}

	err = utils.CheckPassword(body.Password, User.Password)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusUnauthorized)
		return
	}

	jwt, err := utils.GenerateJWT(User.ID, User.Email)
	if err != nil {
		http.Error(w, "Error generating JWT", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(jwt); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}

}
