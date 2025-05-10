package security

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"google.golang.org/grpc/metadata"
)

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

func NewJWTManager(secretKey string, tokenDuration time.Duration) (*JWTManager, error) {
	return &JWTManager{secretKey, tokenDuration}, nil
}

type UserClaims struct {
	jwt.StandardClaims
	Login  string `json:"username"`
	UserID string `json:"user_id"`
}

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

func (manager *JWTManager) ExtractUserID(ctx context.Context) (string, error) {
	tokenString, err := manager.ExtractJWTFromContext(ctx)
	if err != nil {
		return "", err
	}

	token, err := manager.ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}
	return claims.UserID, nil
}

func (manager *JWTManager) ParseToken(accessToken string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(
		accessToken,
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected token signing method")
			}

			return []byte(manager.secretKey), nil
		},
	)
}

func (manager *JWTManager) ExtractJWTFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("metadata is not provided")
	}

	values := md["authorization"]
	if len(values) == 0 {
		return "", fmt.Errorf("authorization token is not provided")
	}
	return values[0], nil
}
