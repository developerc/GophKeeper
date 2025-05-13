package userservice

import (
	"context"
	"encoding/base64"

	"github.com/developerc/GophKeeper/internal/config"
	"github.com/developerc/GophKeeper/internal/entity/myerrors"
	"github.com/developerc/GophKeeper/internal/repositories/userrepository"
	"go.uber.org/zap"
)

type UserService interface {
	Create(ctx context.Context, login, password, userID string) error
	Login(ctx context.Context, login, password string) (string, error)
}

var _ UserService = &userServiceImpl{}

type userServiceImpl struct {
	userRepository userrepository.UserRepository
}

// Create метод для регистрации пользователя
func (u userServiceImpl) Create(ctx context.Context, login, password, userID string) error {
	encodedPassword := base64.StdEncoding.EncodeToString([]byte(password))
	//config.Config.GetServerSettings().Logger.Info("Init gRPC service", zap.String("error", err.Error()))
	//log.Info().Msgf("save user with ID %s and login %s to db", userID, login)
	//fmt.Printf("save user with ID %s and login %s to db", userID, login)
	config.ServerSettingsGlob.Logger.Info("Create", zap.String("userservice", "save new user"))
	return u.userRepository.Save(ctx, userID, login, encodedPassword)
}

// Login метод для авторизации
func (u userServiceImpl) Login(ctx context.Context, login, password string) (string, error) {
	//log.Info().Msgf("userservice: find user with login %s in db", login)
	//fmt.Printf("userservice: find user with login %s in db", login)
	config.ServerSettingsGlob.Logger.Info("Login", zap.String("userservice", "find user"))
	foundUser, err := u.userRepository.FindByLogin(ctx, login)
	if err != nil {
		//log.Error().Msgf("user with login %s not found", login)
		//fmt.Printf("user with login %s not found", login)
		config.ServerSettingsGlob.Logger.Info("Login", zap.String("userservice", "user not found"))
		return "", err
	}
	decodedPassword, err := base64.StdEncoding.DecodeString(foundUser.Password)
	if err != nil {
		return "", err
	}

	if password != string(decodedPassword) {
		//log.Error().Msgf("userservice: password %s is invalid", password)
		//fmt.Printf("userservice: password %s is invalid", password)
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
