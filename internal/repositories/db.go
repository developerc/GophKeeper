package repositories

import (
	"context"
	"database/sql"

	"github.com/developerc/GophKeeper/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

const (
	initUsersTableQuery = "" +
		"CREATE TABLE IF NOT EXISTS public.users (" +
		"id varchar(45) primary key, " +
		"login varchar(45) UNIQUE not null, " +
		"password varchar(45) not null" +
		")"
	initDataTableQuery = "" +
		"CREATE TABLE IF NOT EXISTS public.raw_data (" +
		"name varchar(45) unique not null, " +
		"data_type int2 not null, " +
		"data bytea, " +
		"user_id varchar(45) references public.users (id)" +
		")"
)

var db *sql.DB

func InitDB(ctx context.Context, dbAddress string) (*sql.DB, error) {
	if db != nil {
		return db, nil
	}
	db, connectionErr := sql.Open("pgx", dbAddress)
	if connectionErr != nil {
		//log.Println(connectionErr)
		config.ServerSettingsGlob.Logger.Info("InitDB", zap.String("error", connectionErr.Error()))
		return nil, connectionErr
	}
	createTableErr := createTableIfNotExists(ctx, db)
	if createTableErr != nil {
		//log.Println(createTableErr)
		config.ServerSettingsGlob.Logger.Info("InitDB", zap.String("error", createTableErr.Error()))
		return nil, createTableErr
	}
	return db, nil
}

func createTableIfNotExists(ctx context.Context, db *sql.DB) error {
	//_, createUserTableErr := db.Exec(initUsersTableQuery)
	_, createUserTableErr := db.ExecContext(ctx, initUsersTableQuery)
	if createUserTableErr != nil {
		return createUserTableErr
	}

	//_, createRawDataTableErr := db.Exec(initDataTableQuery)
	_, createRawDataTableErr := db.ExecContext(ctx, initDataTableQuery)
	if createRawDataTableErr != nil {
		return createRawDataTableErr
	}
	return nil
}
