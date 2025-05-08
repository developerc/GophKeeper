// config пакет получения конфигурационных данных для сервера
package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/developerc/GophKeeper/internal/logger"
	"go.uber.org/zap"
)

// Config интерфейс ServerSettings
type Config interface {
	GetServerSettings() *ServerSettings
}

// ServerSettings структура для хранения настроечных данных сервера
type ServerSettings struct {
	Logger        *zap.Logger
	Host          string
	Key           string
	DataBaseDsn   string
	TokenDuration time.Duration
}

// ConfigJSON структура для получения настроечных данных из JSON файла.
type ConfigJSON struct {
	Host          string `json:"host_address"`
	Key           string `json:"secret_key"`
	DataBaseDsn   string `json:"database_dsn"`
	TokenDuration int    `json:"token_duration"`
}

// NewServerSettings конструктор ServerSettings
func NewServerSettings() (*ServerSettings, error) {
	const (
		defaultConfig = "internal/config/configJSON.txt"
		usage         = "configuration by JSON file"
	)
	var fileJSON string
	var err error
	serverSettings := &ServerSettings{}
	serverSettings.Logger, err = logger.Initialize("Info")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	flag.StringVar(&fileJSON, "c", defaultConfig, usage)
	hostAddress := flag.String("h", "", "host IP, port")
	secretKey := flag.String("s", "", "secret key")
	databaseDsn := flag.String("d", "", "data base DSN")
	tokenDuration := flag.Int("t", 0, "token duration")
	flag.Parse()

	configJSON := getConfigJSON(fileJSON)
	if configJSON == nil {
		fileJSON = "../config/configJSON.txt"
		configJSON = getConfigJSON(fileJSON)
		if configJSON == nil {
			fileJSON = "../../config/configJSON.txt"
			configJSON = getConfigJSON(fileJSON)
		}
	}

	val, ok := os.LookupEnv("HOST_ADDRESS")
	if !ok || val == "" {
		if !isFlagPassed("h") {
			serverSettings.Host = configJSON.Host
			serverSettings.Logger.Info("host address from fileJSON:", zap.String("host_address", serverSettings.Host))
		} else {
			serverSettings.Host = *hostAddress
			serverSettings.Logger.Info("host address from flag:", zap.String("host_address", serverSettings.Host))
		}
	} else {
		serverSettings.Host = val
		serverSettings.Logger.Info("host address from env:", zap.String("host_address", serverSettings.Host))
	}

	val, ok = os.LookupEnv("SECRET_KEY")
	if !ok || val == "" {
		if !isFlagPassed("s") {
			serverSettings.Key = configJSON.Key
			serverSettings.Logger.Info("secret key from fileJSON:", zap.String("secret_key", serverSettings.Key))
		} else {
			serverSettings.Key = *secretKey
			serverSettings.Logger.Info("secret key from flag:", zap.String("secret_key", serverSettings.Key))
		}
	} else {
		serverSettings.Key = val
		serverSettings.Logger.Info("secret key from env:", zap.String("secret_key", serverSettings.Key))
	}

	val, ok = os.LookupEnv("DATABASE_DSN")
	if !ok || val == "" {
		if !isFlagPassed("d") {
			serverSettings.DataBaseDsn = configJSON.DataBaseDsn
			serverSettings.Logger.Info("data base DSN from fileJSON:", zap.String("database_dsn", serverSettings.DataBaseDsn))
		} else {
			serverSettings.DataBaseDsn = *databaseDsn
			serverSettings.Logger.Info("data base DSN from flag:", zap.String("database_dsn", serverSettings.DataBaseDsn))
		}
	} else {
		serverSettings.DataBaseDsn = val
		serverSettings.Logger.Info("data base DSN from env:", zap.String("database_dsn", serverSettings.DataBaseDsn))
	}

	val, ok = os.LookupEnv("TOKEN_DURATION")
	if !ok || val == "" {
		if !isFlagPassed("t") {
			serverSettings.TokenDuration = time.Duration(time.Duration(configJSON.TokenDuration) * time.Minute)
			serverSettings.Logger.Info("token duration from fileJSON:", zap.String("token_duration", serverSettings.TokenDuration.String()))
		} else {
			serverSettings.TokenDuration = time.Duration(time.Duration(*tokenDuration) * time.Minute)
			serverSettings.Logger.Info("token duration from flag:", zap.String("token_duration", serverSettings.TokenDuration.String()))
		}
	} else {
		if n, err := strconv.Atoi(val); err == nil {
			serverSettings.TokenDuration = time.Duration(time.Duration(n) * time.Minute)
			serverSettings.Logger.Info("token duration from env:", zap.String("token_duration", serverSettings.TokenDuration.String()))
		} else {
			log.Println("error token duration: ", err)
			serverSettings.Logger.Info("token duration from env:", zap.String("error:", err.Error()))
			return nil, err
		}
	}

	return serverSettings, nil
}

// GetServerSettings возвращает экземпляр ServerSettings
func (s *ServerSettings) GetServerSettings() *ServerSettings {
	return s
}

// getConfigJSON получает конфигурационные данные из JSON файла
func getConfigJSON(fileJSON string) *ConfigJSON {
	var configJSON ConfigJSON
	b, err := os.ReadFile(fileJSON)
	if err != nil {
		return nil
	}
	if err = json.Unmarshal(b, &configJSON); err != nil {
		return nil
	}
	return &configJSON
}

// isFlagPassed проверяет был ли применен флаг
func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
