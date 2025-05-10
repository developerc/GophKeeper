package repositories

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	initUsersTableQuery = "" +
		"CREATE TABLE IF NOT EXISTS public.users (" +
		"id varchar(45) primary key, " +
		"login varchar(45) unique not null, " +
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
		log.Println(connectionErr)
		return nil, connectionErr
	}
	createTableErr := createTableIfNotExists(ctx, db)
	if createTableErr != nil {
		log.Println(createTableErr)
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
