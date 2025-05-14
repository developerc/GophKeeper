package main

import (
	"time"

	"github.com/golang-jwt/jwt"
)

// JWTManager структура менеджера JWT токена
type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

// NewJWTManager конструктор менеджера
func NewJWTManager(secretKey string, tokenDuration time.Duration) (*JWTManager, error) {
	return &JWTManager{secretKey, tokenDuration}, nil
}

// UserClaims структура JWT пользователя
type UserClaims struct {
	jwt.StandardClaims
	Login  string `json:"username"`
	UserID string `json:"user_id"`
}

// GenerateJWT генерирует JWT токен
func (manager *JWTManager) GenerateJWT(userID, login string) (string, error) {
	claims := UserClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(manager.tokenDuration).Unix(),
		},
		Login:  login,
		UserID: userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(manager.secretKey))
}
