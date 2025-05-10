package datarepository

import (
	"context"
	"database/sql"
	"log"

	"github.com/omeid/pgerror"
	//"github.com/rs/zerolog/log"

	"github.com/developerc/GophKeeper/internal/entity"
	"github.com/developerc/GophKeeper/internal/entity/myerrors"
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
)

type RawDataRepository interface {
	Save(ctx context.Context, userID, name string, data []byte, dataType entity.DataType) error
	GetByNameAndTypeAndUserID(ctx context.Context, userID, name string, dataType entity.DataType) ([]byte, error)

	GetAllSavedDataNames(ctx context.Context, userID string) ([]string, error)
}

type rawDataRepositoryImpl struct {
	db *sql.DB
}

// New конструктор UserRepository
func New(db *sql.DB) RawDataRepository {
	return &rawDataRepositoryImpl{
		db: db,
	}
}

// Save сохранение зашифрованных данных
func (r *rawDataRepositoryImpl) Save(ctx context.Context, userID, name string, data []byte, dataType entity.DataType) error {
	//log.Info().Msgf("datarepository: save data with name %s for user with ID %s to db", name, userID)
	log.Printf("datarepository: save data with name %s for user with ID %s to db", name, userID)
	_, err := r.db.ExecContext(ctx, insertDataQuery, name, dataType, data, userID)

	if err != nil {
		if e := pgerror.UniqueViolation(err); e != nil {
			return myerrors.NewDataViolationError(name, err)
		}
		return err
	}
	return nil
}

// GetByNameAndTypeAndUserID получение зашифрованных данных
func (r *rawDataRepositoryImpl) GetByNameAndTypeAndUserID(ctx context.Context, userID, name string, dataType entity.DataType) ([]byte, error) {
	var data []byte
	//log.Info().Msgf("datarepository: get data type of %s with name %s for user with ID %s from db", dataType.String(), name, userID)
	log.Printf("datarepository: get data type of %s with name %s for user with ID %s from db", dataType.String(), name, userID)
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

	//log.Info().Msgf("datarepository: get data names for user with ID %s from db", userID)
	log.Printf("datarepository: get data names for user with ID %s from db", userID)
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
