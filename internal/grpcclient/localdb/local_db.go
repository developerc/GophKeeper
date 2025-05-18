package localdb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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

var DB *sql.DB
var DBPath string
var Key string
var cipherManager *security.CipherManager

func InitDB(dbPath string, key string) error {
	Key = key
	DBPath = dbPath
	_, err := os.Stat(dbPath)
	dbExists := !os.IsNotExist(err)
	if !dbExists {
		os.Create(dbPath)
	}

	DB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer DB.Close()
	if !dbExists {
		_, err = DB.Exec(`
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
		return err
	}

	cipherManager, err = security.NewCipherManager(Key)
	if err != nil {
		return err
	}

	return nil
}

// SaveRawData сохраняет в локальную базу сырые данные
func SaveRawData(ctx context.Context, name, data, comment, userID string) error {
	fmt.Println("from local_db SaveRawData")
	DB, err := sql.Open("sqlite", DBPath)
	if err != nil {
		return err
	}
	defer DB.Close()
	encriptData, err := cipherManager.Encrypt([]byte(data))
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = DB.ExecContext(ctx, "INSERT INTO raw_data (name, data_type, data, user_id, comment) VALUES (?,?,?,?,?)", name, 1, encriptData, userID, comment)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// GetRawData получает из локальной базы сырые данные
func GetRawData(ctx context.Context, name string) (string, string, error) {
	fmt.Println("from local_db GetRawData")
	DB, err := sql.Open("sqlite", DBPath)
	if err != nil {
		return "", "", err
	}
	defer DB.Close()
	row := DB.QueryRowContext(context.Background(), "SELECT data, comment from raw_data WHERE name = ?", name)
	var data []byte
	var comment string
	err = row.Scan(&data, &comment)
	if err != nil {
		fmt.Println(err)
		return "", "", err
	}
	fmt.Println(string(data), comment)
	decriptData, err := cipherManager.Decrypt(data)
	if err != nil {
		fmt.Println(err)
		return "", "", err
	}

	return string(decriptData), comment, nil
}

func SaveLoginWithPassword(ctx context.Context, name, lgn, psw, comment, userID string) error {
	fmt.Println("from local_db SaveLoginWithPassword")
	cred := CredentialsDTO{
		Login:    lgn,
		Password: psw,
	}
	marshalledCred, err := json.Marshal(cred)
	if err != nil {
		return err
	}

	DB, err := sql.Open("sqlite", DBPath)
	if err != nil {
		return err
	}
	defer DB.Close()

	encriptData, err := cipherManager.Encrypt(marshalledCred)
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = DB.ExecContext(ctx, "INSERT INTO raw_data (name, data_type, data, user_id, comment) VALUES (?,?,?,?,?)", name, 2, encriptData, userID, comment)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func GetLoginWithPassword(ctx context.Context, name string) (string, string, string, error) {
	fmt.Println("from local_db GetLoginWithPassword")
	DB, err := sql.Open("sqlite", DBPath)
	if err != nil {
		return "", "", "", err
	}
	defer DB.Close()
	row := DB.QueryRowContext(context.Background(), "SELECT data, comment from raw_data WHERE name = ?", name)
	var data []byte
	var comment string
	err = row.Scan(&data, &comment)
	if err != nil {
		fmt.Println(err)
		return "", "", "", err
	}

	decriptData, err := cipherManager.Decrypt(data)
	if err != nil {
		fmt.Println(err)
		return "", "", "", err
	}

	cred := CredentialsDTO{}
	err = json.Unmarshal(decriptData, &cred)
	if err != nil {
		fmt.Println(err)
		return "", "", "", err
	}

	return cred.Login, cred.Password, comment, nil
}

func SaveBinaryData(ctx context.Context, name string, data []byte, comment, userID string) error {
	fmt.Println("from local_db SaveBinaryData")
	DB, err := sql.Open("sqlite", DBPath)
	if err != nil {
		return err
	}
	defer DB.Close()
	encriptData, err := cipherManager.Encrypt(data)
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = DB.ExecContext(ctx, "INSERT INTO raw_data (name, data_type, data, user_id, comment) VALUES (?,?,?,?,?)", name, 3, encriptData, userID, comment)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func GetBinaryData(ctx context.Context, name string) (string, string, error) {
	fmt.Println("from local_db GetBinaryData")
	DB, err := sql.Open("sqlite", DBPath)
	if err != nil {
		return "", "", err
	}
	defer DB.Close()
	row := DB.QueryRowContext(context.Background(), "SELECT data, comment from raw_data WHERE name = ?", name)
	var data []byte
	var comment string
	err = row.Scan(&data, &comment)
	if err != nil {
		fmt.Println(err)
		return "", "", err
	}
	fmt.Println(string(data), comment)
	decriptData, err := cipherManager.Decrypt(data)
	if err != nil {
		fmt.Println(err)
		return "", "", err
	}

	return string(decriptData), comment, nil
}

func SaveCardData(ctx context.Context, name, number, month, year, cardHolder, cvv, comment, userID string) error {
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

	DB, err := sql.Open("sqlite", DBPath)
	if err != nil {
		return err
	}
	defer DB.Close()

	encriptData, err := cipherManager.Encrypt(marshalledCard)
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = DB.ExecContext(ctx, "INSERT INTO raw_data (name, data_type, data, user_id, comment) VALUES (?,?,?,?,?)", name, 4, encriptData, userID, comment)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func GetCardData(ctx context.Context, name string) (string, string, string, string, string, string, error) {
	fmt.Println("from local_db GetCardData")
	DB, err := sql.Open("sqlite", DBPath)
	if err != nil {
		return "", "", "", "", "", "", err
	}
	defer DB.Close()
	row := DB.QueryRowContext(context.Background(), "SELECT data, comment from raw_data WHERE name = ?", name)
	var data []byte
	var comment string
	err = row.Scan(&data, &comment)
	if err != nil {
		fmt.Println(err)
		return "", "", "", "", "", "", err
	}

	decriptData, err := cipherManager.Decrypt(data)
	if err != nil {
		fmt.Println(err)
		return "", "", "", "", "", "", err
	}

	card := CardDataDTO{}
	err = json.Unmarshal(decriptData, &card)
	if err != nil {
		fmt.Println(err)
		return "", "", "", "", "", "", err
	}

	return card.Number, card.Month, card.Year, card.CardHolder, card.Cvv, comment, nil
}
