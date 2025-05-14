// security пакет шифрования
package security

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/developerc/GophKeeper/internal/config"
	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// JWTManager менеджер JWT
type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

// NewJWTManager конструктор менеджера JWT
func NewJWTManager(secretKey string, tokenDuration time.Duration) (*JWTManager, error) {
	return &JWTManager{secretKey, tokenDuration}, nil
}

// UserClaims структура пользователя
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

// ExtractUserID разбирает JWT токен
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

// ParseToken парсит токен
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

// ExtractJWTFromContext вытаскивает JWT токен из контекста
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

// ServerJwtInterceptor серверный перехватчик. Если это не запрос на регистрацию или аутентификацию проверяет срок действия токена.
// Если токен перехвачен злоумышленником, им пытаются воспользоваться, то по окончании token_duration это будет не возможно.
func ServerJwtInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if strings.Contains(info.FullMethod, "CreateUser") || strings.Contains(info.FullMethod, "LoginUser") {
		config.ServerSettingsGlob.Logger.Info("ServerJwtInterceptor", zap.String("method", info.FullMethod))
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("metadata is not provided")
	}

	accessToken := md["authorization"]
	if len(accessToken) == 0 {
		return "", fmt.Errorf("authorization token is not provided")
	}

	token, err := jwt.ParseWithClaims(
		accessToken[0],
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected token signing method")
			}

			return []byte(config.ServerSettingsGlob.Key), nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error token parse with claims")
	}
	if !token.Valid {
		return nil, fmt.Errorf("error token not valid")
	}

	claims, ok := token.Claims.(*UserClaims)

	if !ok {
		return nil, fmt.Errorf("error token parse with claims")
	} else {
		expirationTime := time.Unix(claims.StandardClaims.ExpiresAt, 0)
		if time.Now().After(expirationTime) {
			return nil, fmt.Errorf("error jwt token is old")
		}
	}

	return handler(ctx, req)
}
