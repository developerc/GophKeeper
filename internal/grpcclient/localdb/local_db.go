package localdb

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	"os"

	"github.com/developerc/GophKeeper/internal/security"

	_ "modernc.org/sqlite"
)

type CredentialsDTO struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type CardDataDTO struct {
	Number     string `json:"number"`
	Month      string `json:"month"`
	Year       string `json:"year"`
	CardHolder string `json:"card_holder"`
	Cvv        string `json:"cvv"`
}

// LocalSQLiteManager интерфейс менеджера локальной БД SQLite
type LocalSQLiteManager interface {
	SaveRawData(ctx context.Context, name, data, comment, userID string) error
	GetRawData(ctx context.Context, name string) (string, string, error)
	SaveLoginWithPassword(ctx context.Context, name, lgn, psw, comment, userID string) error
	GetLoginWithPassword(ctx context.Context, name string) (string, string, string, error)
	SaveBinaryData(ctx context.Context, name string, data []byte, comment, userID string) error
	GetBinaryData(ctx context.Context, name string) (string, string, error)
	SaveCardData(ctx context.Context, name, number, month, year, cardHolder, cvv, comment, userID string) error
	GetCardData(ctx context.Context, name string) (string, string, string, string, string, string, error)
	GetAllSavedDataNames(ctx context.Context) ([]string, error)
	DelRawData(ctx context.Context, name string) error
	DelLoginWithPassword(ctx context.Context, name string) error
	DelBinaryData(ctx context.Context, name string) error
	DelCardData(ctx context.Context, name string) error
	UpdRawData(ctx context.Context, name, data, comment, userID string) error
	UpdLoginWithPassword(ctx context.Context, name, lgn, psw, comment, userID string) error
	UpdBinaryData(ctx context.Context, name string, binData []byte, comment, userID string) error
	UpdCardData(ctx context.Context, name, number, month, year, cardHolder, cvv, comment, userID string) error
}

// LocalSQLite структура менеджера локальной БД SQLite
type LocalSQLite struct {
	DB            *sql.DB
	DBPath        string
	Key           string
	cipherManager *security.CipherManager
}

// NewLocalSQLiteManager конструктор менеджера локальной БД SQLite
func NewLocalSQLiteManager(dbPath string, key string) (*LocalSQLite, error) {
	localSQLite := LocalSQLite{}
	localSQLite.Key = key
	localSQLite.DBPath = dbPath
	_, err := os.Stat(dbPath)
	dbExists := !os.IsNotExist(err)
	if !dbExists {
		os.Create(dbPath)
	}

	localSQLite.DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	defer localSQLite.DB.Close()
	if !dbExists {
		_, err = localSQLite.DB.Exec(`
		CREATE TABLE IF NOT EXISTS raw_data (
			name TEXT,
			data_type INTEGER,
			data BLOB,
			user_id TEXT,
			comment TEXT
		);
	`)
	}
	if err != nil {
		return nil, err
	}

	localSQLite.cipherManager, err = security.NewCipherManager(localSQLite.Key)
	if err != nil {
		return nil, err
	}
	return &localSQLite, nil
}

// SaveRawData сохраняет в локальную базу сырые данные
func (ls *LocalSQLite) SaveRawData(ctx context.Context, name, data, comment, userID string) error {
	var err error
	ls.DB, err = sql.Open("sqlite", ls.DBPath)
	if err != nil {
		return err
	}
	defer ls.DB.Close()
	encriptData, err := ls.cipherManager.Encrypt([]byte(data))
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = ls.DB.ExecContext(ctx, "INSERT INTO raw_data (name, data_type, data, user_id, comment) VALUES (?,?,?,?,?)", name, 1, encriptData, userID, comment)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// GetRawData получает из локальной базы сырые данные
func (ls *LocalSQLite) GetRawData(ctx context.Context, name string) (string, string, error) {
	var err error
	ls.DB, err = sql.Open("sqlite", ls.DBPath)
	if err != nil {
		return "", "", err
	}
	defer ls.DB.Close()
	row := ls.DB.QueryRowContext(context.Background(), "SELECT data, comment from raw_data WHERE name = ?", name)
	var data []byte
	var comment string
	err = row.Scan(&data, &comment)
	if err != nil {
		log.Println(err)
		return "", "", err
	}

	decriptData, err := ls.cipherManager.Decrypt(data)
	if err != nil {
		log.Println(err)
		return "", "", err
	}

	return string(decriptData), comment, nil
}

// SaveLoginWithPassword сохраняет логин и пароль для авторизованного пользователя
func (ls *LocalSQLite) SaveLoginWithPassword(ctx context.Context, name, lgn, psw, comment, userID string) error {
	var err error
	cred := CredentialsDTO{
		Login:    lgn,
		Password: psw,
	}
	marshalledCred, err := json.Marshal(cred)
	if err != nil {
		return err
	}

	ls.DB, err = sql.Open("sqlite", ls.DBPath)
	if err != nil {
		return err
	}
	defer ls.DB.Close()

	encriptData, err := ls.cipherManager.Encrypt(marshalledCred)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = ls.DB.ExecContext(ctx, "INSERT INTO raw_data (name, data_type, data, user_id, comment) VALUES (?,?,?,?,?)", name, 2, encriptData, userID, comment)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// GetLoginWithPassword получает логин и пароль по названию для авторизованного пользователя
func (ls *LocalSQLite) GetLoginWithPassword(ctx context.Context, name string) (string, string, string, error) {
	var err error
	ls.DB, err = sql.Open("sqlite", ls.DBPath)
	if err != nil {
		return "", "", "", err
	}
	defer ls.DB.Close()
	row := ls.DB.QueryRowContext(context.Background(), "SELECT data, comment from raw_data WHERE name = ?", name)
	var data []byte
	var comment string
	err = row.Scan(&data, &comment)
	if err != nil {
		log.Println(err)
		return "", "", "", err
	}

	decriptData, err := ls.cipherManager.Decrypt(data)
	if err != nil {
		log.Println(err)
		return "", "", "", err
	}

	cred := CredentialsDTO{}
	err = json.Unmarshal(decriptData, &cred)
	if err != nil {
		log.Println(err)
		return "", "", "", err
	}

	return cred.Login, cred.Password, comment, nil
}

// SaveBinaryData сохранение произвольных бинарных данных для авторизованного пользователя
func (ls *LocalSQLite) SaveBinaryData(ctx context.Context, name string, data []byte, comment, userID string) error {
	var err error
	ls.DB, err = sql.Open("sqlite", ls.DBPath)
	if err != nil {
		return err
	}
	defer ls.DB.Close()
	encriptData, err := ls.cipherManager.Encrypt(data)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = ls.DB.ExecContext(ctx, "INSERT INTO raw_data (name, data_type, data, user_id, comment) VALUES (?,?,?,?,?)", name, 3, encriptData, userID, comment)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// GetBinaryData получение произвольных бинарных данных по названию для авторизованного пользователя
func (ls *LocalSQLite) GetBinaryData(ctx context.Context, name string) (string, string, error) {
	var err error
	ls.DB, err = sql.Open("sqlite", ls.DBPath)
	if err != nil {
		return "", "", err
	}
	defer ls.DB.Close()
	row := ls.DB.QueryRowContext(context.Background(), "SELECT data, comment from raw_data WHERE name = ?", name)
	var data []byte
	var comment string
	err = row.Scan(&data, &comment)
	if err != nil {
		log.Println(err)
		return "", "", err
	}
	log.Println(string(data), comment)
	decriptData, err := ls.cipherManager.Decrypt(data)
	if err != nil {
		log.Println(err)
		return "", "", err
	}

	return string(decriptData), comment, nil
}

// SaveCardData сохранение данных банковской карты для авторизованного пользователя
func (ls *LocalSQLite) SaveCardData(ctx context.Context, name, number, month, year, cardHolder, cvv, comment, userID string) error {
	card := CardDataDTO{
		Number:     number,
		Month:      month,
		Year:       year,
		CardHolder: cardHolder,
		Cvv:        cvv,
	}

	marshalledCard, err := json.Marshal(card)
	if err != nil {
		return err
	}

	ls.DB, err = sql.Open("sqlite", ls.DBPath)
	if err != nil {
		return err
	}
	defer ls.DB.Close()

	encriptData, err := ls.cipherManager.Encrypt(marshalledCard)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = ls.DB.ExecContext(ctx, "INSERT INTO raw_data (name, data_type, data, user_id, comment) VALUES (?,?,?,?,?)", name, 4, encriptData, userID, comment)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// GetCardData получение данных банковской карты по названию для авторизованного пользователя
func (ls *LocalSQLite) GetCardData(ctx context.Context, name string) (string, string, string, string, string, string, error) {
	var err error
	ls.DB, err = sql.Open("sqlite", ls.DBPath)
	if err != nil {
		return "", "", "", "", "", "", err
	}
	defer ls.DB.Close()
	row := ls.DB.QueryRowContext(context.Background(), "SELECT data, comment from raw_data WHERE name = ?", name)
	var data []byte
	var comment string
	err = row.Scan(&data, &comment)
	if err != nil {
		log.Println(err)
		return "", "", "", "", "", "", err
	}

	decriptData, err := ls.cipherManager.Decrypt(data)
	if err != nil {
		log.Println(err)
		return "", "", "", "", "", "", err
	}

	card := CardDataDTO{}
	err = json.Unmarshal(decriptData, &card)
	if err != nil {
		log.Println(err)
		return "", "", "", "", "", "", err
	}

	return card.Number, card.Month, card.Year, card.CardHolder, card.Cvv, comment, nil
}

// GetAllSavedDataNames получение всех названий сохранений
func (ls *LocalSQLite) GetAllSavedDataNames(ctx context.Context) ([]string, error) {
	var err error
	ls.DB, err = sql.Open("sqlite", ls.DBPath)
	if err != nil {
		return nil, err
	}
	defer ls.DB.Close()
	rows, err := ls.DB.QueryContext(ctx, "SELECT name FROM raw_data")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	names := make([]string, 0)
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return names, nil
}

// DelRawData удаляет сырые данные
func (ls *LocalSQLite) DelRawData(ctx context.Context, name string) error {
	var err error
	ls.DB, err = sql.Open("sqlite", ls.DBPath)
	if err != nil {
		return err
	}
	defer ls.DB.Close()

	_, err = ls.DB.ExecContext(ctx, "DELETE FROM raw_data WHERE name=?", name)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// DelLoginWithPassword удаляет данные логин, пароль
func (ls *LocalSQLite) DelLoginWithPassword(ctx context.Context, name string) error {
	var err error
	ls.DB, err = sql.Open("sqlite", ls.DBPath)
	if err != nil {
		return err
	}
	defer ls.DB.Close()

	_, err = ls.DB.ExecContext(ctx, "DELETE FROM raw_data WHERE name=?", name)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// DelBinaryData удаляет бинарные данные
func (ls *LocalSQLite) DelBinaryData(ctx context.Context, name string) error {
	var err error
	ls.DB, err = sql.Open("sqlite", ls.DBPath)
	if err != nil {
		return err
	}
	defer ls.DB.Close()

	_, err = ls.DB.ExecContext(ctx, "DELETE FROM raw_data WHERE name=?", name)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// DelCardData удаляет данные карты
func (ls *LocalSQLite) DelCardData(ctx context.Context, name string) error {
	var err error
	ls.DB, err = sql.Open("sqlite", ls.DBPath)
	if err != nil {
		return err
	}
	defer ls.DB.Close()

	_, err = ls.DB.ExecContext(ctx, "DELETE FROM raw_data WHERE name=?", name)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// UpdRawData обновляет произвольную текстовую информацию для авторизованного пользователя
func (ls *LocalSQLite) UpdRawData(ctx context.Context, name, data, comment, userID string) error {
	var err error
	ls.DB, err = sql.Open("sqlite", ls.DBPath)
	if err != nil {
		return err
	}
	defer ls.DB.Close()
	encriptData, err := ls.cipherManager.Encrypt([]byte(data))
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = ls.DB.ExecContext(ctx, "UPDATE raw_data SET data=?, user_id=?, comment=? WHERE name=?", encriptData, userID, comment, name)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// UpdLoginWithPassword обновляет логин и пароль для авторизованного пользователя
func (ls *LocalSQLite) UpdLoginWithPassword(ctx context.Context, name, lgn, psw, comment, userID string) error {
	cred := CredentialsDTO{
		Login:    lgn,
		Password: psw,
	}
	marshalledCred, err := json.Marshal(cred)
	if err != nil {
		return err
	}

	ls.DB, err = sql.Open("sqlite", ls.DBPath)
	if err != nil {
		return err
	}
	defer ls.DB.Close()

	encriptData, err := ls.cipherManager.Encrypt(marshalledCred)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = ls.DB.ExecContext(ctx, "UPDATE raw_data SET data=?, user_id=?, comment=? WHERE name=?", encriptData, userID, comment, name)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// SaveBinaryData обновляет произвольных бинарных данных для авторизованного пользователя
func (ls *LocalSQLite) UpdBinaryData(ctx context.Context, name string, binData []byte, comment, userID string) error {
	var err error
	ls.DB, err = sql.Open("sqlite", ls.DBPath)
	if err != nil {
		return err
	}
	defer ls.DB.Close()
	encriptData, err := ls.cipherManager.Encrypt(binData)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = ls.DB.ExecContext(ctx, "UPDATE raw_data SET data=?, user_id=?, comment=? WHERE name=?", encriptData, userID, comment, name)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// UpdCardData обновляет данных банковской карты для авторизованного пользователя
func (ls *LocalSQLite) UpdCardData(ctx context.Context, name, number, month, year, cardHolder, cvv, comment, userID string) error {
	card := CardDataDTO{
		Number:     number,
		Month:      month,
		Year:       year,
		CardHolder: cardHolder,
		Cvv:        cvv,
	}

	marshalledCard, err := json.Marshal(card)
	if err != nil {
		return err
	}

	ls.DB, err = sql.Open("sqlite", ls.DBPath)
	if err != nil {
		return err
	}
	defer ls.DB.Close()

	encriptData, err := ls.cipherManager.Encrypt(marshalledCard)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = ls.DB.ExecContext(ctx, "UPDATE raw_data SET data=?, user_id=?, comment=? WHERE name=?", encriptData, userID, comment, name)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
