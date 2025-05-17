// userrepository пакет репозитория пользователей
package userrepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/developerc/GophKeeper/internal/config"
	"github.com/developerc/GophKeeper/internal/entity"
	"github.com/developerc/GophKeeper/internal/entity/myerrors"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

const (
	insertUserQuery = "" +
		"INSERT INTO public.users (id, login, password) " +
		"VALUES ($1, $2, $3)"
	findUserByLoginQuery = "" +
		"SELECT id, login, password FROM public.users " +
		"WHERE login=$1"
)

// UserRepository интерфейс пользовательского репозитория
type UserRepository interface {
	Save(ctx context.Context, userID, login, password string) error
	FindByLogin(ctx context.Context, login string) (entity.UserDTO, error)
}

// userRepositoryImpl структура пользовательского репозитория
type userRepositoryImpl struct {
	db *sql.DB
}

// New конструктор UserRepository
func New(db *sql.DB) UserRepository {
	return &userRepositoryImpl{
		db: db,
	}
}

// isDuplicateKeyError проверяет наличие ошибки при добавлении существующего пользователя
func isDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// Save сохранение юзера
func (r *userRepositoryImpl) Save(ctx context.Context, userID, login, password string) error {
	config.ServerSettingsGlob.Logger.Info("Save", zap.String("userrepository", "save user to db"))
	_, err := r.db.ExecContext(ctx, insertUserQuery, userID, login, password)
	if err != nil {
		if isDuplicateKeyError(err) {
			return myerrors.NewUserViolationError(login, err)
		}
		return err
	}
	return nil
}

// FindByLogin поиск юзера в базе по логину
func (r *userRepositoryImpl) FindByLogin(ctx context.Context, login string) (entity.UserDTO, error) {
	var user entity.UserDTO

	config.ServerSettingsGlob.Logger.Info("FindByLogin", zap.String("userrepository", "find user from db"))
	row := r.db.QueryRowContext(ctx, findUserByLoginQuery, login)
	err := row.Scan(&user.ID, &user.Login, &user.Password)
	if err != nil {
		return entity.UserDTO{}, err
	}
	return user, nil
}
