package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var secretKey = []byte("secret")

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func GenerateJWT(userID uuid.UUID, email string) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   userID.String(),
		"email": email,
		"aud":   "default",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	})

	token, err := claims.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return token, nil
}
