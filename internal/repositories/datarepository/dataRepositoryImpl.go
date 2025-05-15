// datarepository пакет репозитория данных
package datarepository

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
	insertDataQuery = "" +
		"INSERT INTO public.raw_data (name, data_type, data, user_id) " +
		"VALUES ($1, $2, $3, $4)"
	getDataQuery = "" +
		"SELECT data FROM public.raw_data " +
		"WHERE user_id=$1 AND name=$2 AND data_type=$3"
	getAllDataNamesByUserIDQuery = "" +
		"SELECT name " +
		"FROM public.raw_data " +
		"WHERE user_id=$1"
	delDataByNameUserId = "" +
		"DELETE FROM public.raw_data WHERE name=$1 AND user_id=$2"
)

// RawDataRepository интерфейс репозитория данных
type RawDataRepository interface {
	Save(ctx context.Context, userID, name string, data []byte, dataType entity.DataType) error
	GetByNameAndTypeAndUserID(ctx context.Context, userID, name string, dataType entity.DataType) ([]byte, error)
	GetAllSavedDataNames(ctx context.Context, userID string) ([]string, error)
	DelDataByNameUserId(ctx context.Context, name, userID string) error
}

// rawDataRepositoryImpl структура репозитория данных
type rawDataRepositoryImpl struct {
	db *sql.DB
}

// New конструктор UserRepository
func New(db *sql.DB) RawDataRepository {
	return &rawDataRepositoryImpl{
		db: db,
	}
}

func isDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// Save сохранение зашифрованных данных
func (r *rawDataRepositoryImpl) Save(ctx context.Context, userID, name string, data []byte, dataType entity.DataType) error {
	config.ServerSettingsGlob.Logger.Info("Save", zap.String("datarepository", "save data to db"))
	_, err := r.db.ExecContext(ctx, insertDataQuery, name, dataType, data, userID)

	if err != nil {
		if isDuplicateKeyError(err) {
			return myerrors.NewDataViolationError(name, err)
		}
		return err
	}
	return nil
}

// GetByNameAndTypeAndUserID получение зашифрованных данных
func (r *rawDataRepositoryImpl) GetByNameAndTypeAndUserID(ctx context.Context, userID, name string, dataType entity.DataType) ([]byte, error) {
	var data []byte

	config.ServerSettingsGlob.Logger.Info("GetByNameAndTypeAndUserID", zap.String("datarepository", "get data from db"))
	row := r.db.QueryRowContext(ctx, getDataQuery, userID, name, dataType)
	err := row.Scan(&data)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, myerrors.NewNotFoundError(name, err)
		}
		return nil, err
	}
	return data, nil
}

// GetAllSavedDataNames метод для получения всех названий сохранений
func (r *rawDataRepositoryImpl) GetAllSavedDataNames(ctx context.Context, userID string) ([]string, error) {
	nameList := make([]string, 0)

	config.ServerSettingsGlob.Logger.Info("GetAllSavedDataNames", zap.String("datarepository", "get all names from db"))
	rows, err := r.db.QueryContext(ctx, getAllDataNamesByUserIDQuery, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var n string
		err = rows.Scan(&n)
		if err != nil {
			return nil, err
		}
		nameList = append(nameList, n)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return nameList, nil
}

func (r *rawDataRepositoryImpl) DelDataByNameUserId(ctx context.Context, name, userID string) error {
	config.ServerSettingsGlob.Logger.Info("DelDataByNameUserId", zap.String("RawDataRepository", "delete data from db"))
	_, err := r.db.ExecContext(ctx, delDataByNameUserId, name, userID)
	if err != nil {
		return err
	}
	return nil
}
