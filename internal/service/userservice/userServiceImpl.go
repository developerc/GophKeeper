// userservice пакет пользовательского сервиса
package userservice

import (
	"context"
	"encoding/base64"

	"github.com/developerc/GophKeeper/internal/config"
	"github.com/developerc/GophKeeper/internal/entity/myerrors"
	"github.com/developerc/GophKeeper/internal/repositories/userrepository"
	"go.uber.org/zap"
)

// UserService интерфейс пользовательского сервиса
type UserService interface {
	Create(ctx context.Context, login, password, userID string) error
	Login(ctx context.Context, login, password string) (string, error)
}

// UserService экземпляр пользовательского сервиса
var _ UserService = &userServiceImpl{}

// userServiceImpl структура пользовательского сервиса
type userServiceImpl struct {
	userRepository userrepository.UserRepository
}

// Create метод для регистрации пользователя
func (u userServiceImpl) Create(ctx context.Context, login, password, userID string) error {
	encodedPassword := base64.StdEncoding.EncodeToString([]byte(password))
	config.ServerSettingsGlob.Logger.Info("Create", zap.String("userservice", "save new user"))
	return u.userRepository.Save(ctx, userID, login, encodedPassword)
}

// Login метод для авторизации
func (u userServiceImpl) Login(ctx context.Context, login, password string) (string, error) {
	config.ServerSettingsGlob.Logger.Info("Login", zap.String("userservice", "find user"))
	foundUser, err := u.userRepository.FindByLogin(ctx, login)
	if err != nil {
		config.ServerSettingsGlob.Logger.Info("Login", zap.String("userservice", "user not found"))
		return "", err
	}
	decodedPassword, err := base64.StdEncoding.DecodeString(foundUser.Password)
	if err != nil {
		return "", err
	}

	if password != string(decodedPassword) {
		config.ServerSettingsGlob.Logger.Info("Login", zap.String("userservice", "password is invalid"))
		return "", &myerrors.InvalidPasswordError{Password: password}
	}

	return foundUser.ID, nil
}

// New конструктор UserService
func New(userRepository userrepository.UserRepository) UserService {
	return &userServiceImpl{
		userRepository,
	}
}
