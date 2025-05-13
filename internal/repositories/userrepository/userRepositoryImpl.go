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
	//"github.com/jackc/pgconn"
	//"github.com/omeid/pgerror"
)

const (
	insertUserQuery = "" +
		"INSERT INTO public.users (id, login, password) " +
		"VALUES ($1, $2, $3)"
	findUserByLoginQuery = "" +
		"SELECT id, login, password FROM public.users " +
		"WHERE login=$1"
)

type UserRepository interface {
	Save(ctx context.Context, userID, login, password string) error
	FindByLogin(ctx context.Context, login string) (entity.UserDTO, error)
}

type userRepositoryImpl struct {
	db *sql.DB
}

// New конструктор UserRepository
func New(db *sql.DB) UserRepository {
	return &userRepositoryImpl{
		db: db,
	}
}

func isDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// Save сохранение юзера
func (r *userRepositoryImpl) Save(ctx context.Context, userID, login, password string) error {
	//log.Info().Msgf("userrepository: save user with ID %s and login %s to db", userID, login)
	//fmt.Printf("userrepository: save user with ID %s and login %s to db\n", userID, login)
	config.ServerSettingsGlob.Logger.Info("Save", zap.String("userrepository", "save user to db"))
	_, err := r.db.ExecContext(ctx, insertUserQuery, userID, login, password)
	if err != nil {
		//log.Println(err.Error())
		if isDuplicateKeyError(err) {
			return myerrors.NewUserViolationError(login, err)
		}
		//fmt.Println(isDuplicateKeyError(err))
		//err.(*pgconn.PgError)
		/*var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return myerrors.NewUserViolationError(login, err)
			//fmt.Printf("PostgreSQL Error (Code: %s): %s\n", pgErr.Code, pgErr.Message)
		}else {
			fmt.Printf("Non-PostgreSQL error: %v\n", err)
		}*/
		//fmt.Println(reflect.TypeOf(err))
		//fmt.Println(err.(*pq.Error))
		/*if pqErr, ok := err.(*pq.Error); ok {
			fmt.Println(pqErr)
			switch pqErr.Code {
			case "23505": // Unique violation
				return myerrors.NewUserViolationError(login, err)
			default:
				log.Printf("PostgreSQL error %s: %v", pqErr.Code, pqErr.Message)
			}
		}*/
		/*if strings.Contains(err.Error(), "SQLSTATE 23505") {
			return myerrors.NewUserViolationError(login, err)
		}*/
		/*if e := pgerror.UniqueViolation(err); e != nil {
			//log.Println("ERROR UniqueViolation!!!")
			//log.Error().Msgf("userrepository: user with login %s already exists")
			return myerrors.NewUserViolationError(login, err)
		}*/
		return err
	}
	return nil
}

// FindByLogin поиск юзера в базе по логину
func (r *userRepositoryImpl) FindByLogin(ctx context.Context, login string) (entity.UserDTO, error) {
	var user entity.UserDTO
	//log.Info().Msgf("userrepository: find user with login %s in db", login)
	//fmt.Printf("userrepository: find user with login %s in db\n", login)
	config.ServerSettingsGlob.Logger.Info("FindByLogin", zap.String("userrepository", "find user from db"))
	row := r.db.QueryRowContext(ctx, findUserByLoginQuery, login)
	err := row.Scan(&user.ID, &user.Login, &user.Password)
	if err != nil {
		return entity.UserDTO{}, err
	}
	return user, nil
}
