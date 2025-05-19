package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/developerc/GophKeeper/internal/config"
	"github.com/developerc/GophKeeper/internal/grpcclient/localdb"
	pb "github.com/developerc/GophKeeper/proto"
	"github.com/golang-jwt/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type ClientManager struct {
	ServerSettings   *config.ServerSettings
	UserClaims       *UserClaims
	Err              error
	Lgn              string
	Psw              string
	UserID           string
	ClientJWTManager *JWTManager
	GrpcClient       pb.GrpcServiceClient
}

var localMode bool

func NewClientManager() (*ClientManager, error) {
	clientManager := ClientManager{}
	ss, err := config.NewServerSettings()
	if err != nil {
		return nil, err
	}
	clientManager.ServerSettings = ss
	cJWTm, err := NewJWTManager(ss.Key, ss.TokenDuration)
	if err != nil {
		return nil, err
	}
	clientManager.ClientJWTManager = cJWTm
	return &clientManager, nil
}

var (
	//go:embed version.txt
	buildVersion string
	//go:embed date.txt
	buildDate string
)

const menu = "" +
	"МЕНЮ:\n" +
	"0 Выход из приложения\n" +
	"1 Регистрация пользователя\n" +
	"2 Аутентификация\n" +
	"3 Сохранение сырых данных \"строка\"\n" +
	"4 Получение сырых данных \"строка\"\n" +
	"5 Сохранение логин, пароля\n" +
	"6 Получение логин, пароля\n" +
	"7 Сохранение бинарных данных\n" +
	"8 Получение бинарных данных\n" +
	"9 Сохранение данных карты\n" +
	"10 Получение данных карты\n" +
	"11 Получение всех сохраненных имен\n" +
	"12 Удаление сырых данных\n" +
	"13 Удаление логин, пароля\n" +
	"14 Удаление бинарных данных\n" +
	"15 Удаление данных карты\n" +
	"16 Обновление сырых данных\n" +
	"17 Обновление данных логин, пароль\n" +
	"18 Обновление бинарных данных\n" +
	"19 Обновление данных карты\n" +
	"20 Проверка соединения с gRPC сервером\n"

// main запускает клиента gRPC
func main() {

	log.SetFlags(0)
	BuildVersion := strings.TrimSpace(buildVersion)
	if len(BuildVersion) > 0 {
		log.Printf("Build version: %q\n", BuildVersion)
	} else {
		log.Printf("Build version: N/A\n")
	}

	BuildDate := strings.TrimSpace(buildDate)
	if len(BuildDate) > 0 {
		log.Printf("Build date: %q\n", BuildDate)
	} else {
		log.Printf("Build date: N/A\n")
	}
	log.SetFlags(3)
	var choice int
	cm, err := NewClientManager()
	if err != nil {
		log.Fatal(err)
	}
	err = localdb.InitDB("raw_data.db", cm.ServerSettings.Key)
	if err != nil {
		log.Fatal(err)
	}
	// создадим клиент grpc
	conn, err := grpc.NewClient(cm.ServerSettings.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	cm.GrpcClient = pb.NewGrpcServiceClient(conn)

	fmt.Print(menu)

	for {
		fmt.Println("Сделайте выбор меню:")
		fmt.Scan(&choice)
		switch choice {
		case 0:
			exitApp()
		case 1:
			CreateUser(cm)
		case 2:
			LoginUser(cm)
		case 3:
			SaveRawData(cm)
		case 4:
			GetRawData(cm)
		case 5:
			SaveLoginWithPassword(cm)
		case 6:
			GetLoginWithPassword(cm)
		case 7:
			SaveBinaryData(cm)
		case 8:
			GetBinaryData(cm)
		case 9:
			SaveCardData(cm)
		case 10:
			GetCardData(cm)
		case 11:
			GetAllSavedDataNames(cm)
		case 12:
			DelRawData(cm)
		case 13:
			DelLoginWithPassword(cm)
		case 14:
			DelBinaryData(cm)
		case 15:
			DelCardData(cm)
		case 16:
			UpdRawData(cm)
		case 17:
			UpdLoginWithPassword(cm)
		case 18:
			UpdBinaryData(cm)
		case 19:
			UpdCardData(cm)
		case 20:
			errorResponse, err := cm.CheckConnectCall()
			if err != nil {
				fmt.Println(err)
				fmt.Println("Проверка соединения неуспешна. Работа только в локальном режиме.")
				localMode = true
				//fmt.Println(localMode)
			} else {
				fmt.Println("Проверка соединения успешна: ", errorResponse.Error, "Работа в штатном режиме.")
				localMode = false
				//fmt.Println(localMode)
			}
		}
	}

}

// exitApp завершает приложение
func exitApp() {
	var confirm string
	fmt.Print("Подтвердите выход y/n ")
	fmt.Scan(&confirm)
	if confirm == "y" {
		os.Exit(0)
	}
}

// CreateUser сохраняет нового пользователя
func CreateUser(cm *ClientManager) {
	var lgn string
	var psw string

	fmt.Print("Регистрируем пользователя. Введите логин: ")
	fmt.Scan(&lgn)
	fmt.Print("Введите пароль: ")
	fmt.Scan(&psw)
	cm.Lgn = lgn
	cm.Psw = psw
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := cm.CreateUser(ctx)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Регистрация прошла успешно, userID: ", cm.UserID)
	}
}

// LoginUser авторизация существующего пользователя
func LoginUser(cm *ClientManager) {
	var lgn string
	var psw string

	fmt.Print("Проводим аутентификацию. Введите логин: ")
	fmt.Scan(&lgn)
	fmt.Print("Введите пароль: ")
	fmt.Scan(&psw)
	cm.Lgn = lgn
	cm.Psw = psw
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cm.Lgn = lgn
	cm.Psw = psw
	err := cm.LoginUser(ctx)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Аутентификация прошла успешно, userID: ", cm.UserID)
	}
}

// SaveRawData сохранения произвольную текстовую информацию для авторизованного пользователя
func SaveRawData(cm *ClientManager) {
	if localMode {
		fmt.Println("Удаленный сервер недоступен. Возможно только чтение из локальной БД.")
		return
	}

	var name string
	var data string
	var comment string
	fmt.Print("Добавляем сырые данные, строка. Введите имя в хранилище: ")
	fmt.Scan(&name)
	fmt.Print("Введите сохраняемую строку: ")
	fmt.Scan(&data)
	fmt.Print("Введите комментарий: ")
	fmt.Scan(&comment)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	errorResponse, err := cm.SaveRawData(ctx, name, data, comment)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Сырые данные сохранены успешно: ", errorResponse.Error)
	}
}

// GetRawData получение текстовой информации по названию для авторизованного пользователя
func GetRawData(cm *ClientManager) {
	var name string
	fmt.Print("Получаем сырые данные, строка. Введите имя в хранилище: ")
	fmt.Scan(&name)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if localMode {
		data, comment, err := localdb.GetRawData(ctx, name)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("Получены из локальной БД сырые данные %s, комментарий %s\n", data, comment)
		return
	}
	getRawDataResponse, err := cm.GetRawData(ctx, name)
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("Получены сырые данные %s, комментарий %s\n", getRawDataResponse.Data, getRawDataResponse.Comment)
	}
}

// SaveLoginWithPassword сохраняет логин и пароль для авторизованного пользователя
func SaveLoginWithPassword(cm *ClientManager) {
	if localMode {
		fmt.Println("Удаленный сервер недоступен. Возможно только чтение из локальной БД.")
		return
	}

	var name string
	var login string
	var password string
	var comment string
	fmt.Print("Добавляем логин, пароль. Введите имя в хранилище: ")
	fmt.Scan(&name)
	fmt.Print("Введите сохраняемый логин: ")
	fmt.Scan(&login)
	fmt.Print("Введите сохраняемый пароль: ")
	fmt.Scan(&password)
	fmt.Print("Введите комментарий: ")
	fmt.Scan(&comment)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	errorResponse, err := cm.SaveLoginWithPassword(ctx, name, login, password, comment)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Логин, пароль сохранены успешно: ", errorResponse.Error)
	}
}

// GetLoginWithPassword получает логин и пароль по названию для авторизованного пользователя
func GetLoginWithPassword(cm *ClientManager) {
	var name string
	fmt.Print("Получаем логин, пароль. Введите имя в хранилище: ")
	fmt.Scan(&name)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if localMode {
		login, password, comment, err := localdb.GetLoginWithPassword(ctx, name)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("Получены из локальной базы логин %s, пароль %s, комментарий %s\n", login, password, comment)
		return
	}
	getLoginWithPasswordResponse, err := cm.GetLoginWithPassword(ctx, name)
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("Получены логин %s, пароль %s, комментарий %s\n", getLoginWithPasswordResponse.Login, getLoginWithPasswordResponse.Password, getLoginWithPasswordResponse.Comment)
	}
}

// SaveBinaryData сохранение произвольных бинарных данных для авторизованного пользователя
func SaveBinaryData(cm *ClientManager) {
	if localMode {
		fmt.Println("Удаленный сервер недоступен. Возможно только чтение из локальной БД.")
		return
	}

	var name string
	var filePath string
	var myBinary []byte
	var comment string
	fmt.Print("Добавляем бинарные данные. Введите имя в хранилище: ")
	fmt.Scan(&name)
	fmt.Print("Введите полный путь к файлу: ")
	fmt.Scan(&filePath)
	fmt.Print("Введите комментарий: ")
	fmt.Scan(&comment)

	myBinary, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	errorResponse, err := cm.SaveBinaryData(ctx, name, myBinary, comment)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Бинарные данные сохранены успешно: ", errorResponse.Error)
	}
}

// GetBinaryData получение произвольных бинарных данных по названию для авторизованного пользователя
func GetBinaryData(cm *ClientManager) {
	var name string
	fmt.Print("Получаем бинарные данные. Введите имя в хранилище: ")
	fmt.Scan(&name)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if localMode {
		data, comment, err := localdb.GetBinaryData(ctx, name)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("Получены из локальной БД бинарные данные %s, комментарий %s\n", data, comment)
		return
	}
	getBinaryDataResponse, err := cm.GetBinaryData(ctx, name)
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("Получены бинарные данные строка %s, комментарий %s\n", string(getBinaryDataResponse.Data), getBinaryDataResponse.Comment)
	}
}

// SaveCardData сохранение данных банковской карты для авторизованного пользователя
func SaveCardData(cm *ClientManager) {
	if localMode {
		fmt.Println("Удаленный сервер недоступен. Возможно только чтение из локальной БД.")
		return
	}

	var name string
	var number string
	var month string
	var year string
	var cardHolder string
	var cvv string
	var comment string

	fmt.Println("Добавляем данные карты. Введите имя в хранилище: ")
	fmt.Scan(&name)
	fmt.Print("Введите номер карты: ")
	fmt.Scan(&number)
	fmt.Print("Введите месяц выдачи карты: ")
	fmt.Scan(&month)
	fmt.Print("Введите год выдачи карты: ")
	fmt.Scan(&year)
	fmt.Print("Введите держателя карты: ")
	fmt.Scan(&cardHolder)
	fmt.Print("Введите CVV карты: ")
	fmt.Scan(&cvv)
	fmt.Print("Введите комментарий: ")
	fmt.Scan(&comment)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	errorResponse, err := cm.SaveCardData(ctx, name, number, month, year, cardHolder, cvv, comment)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Данные карты сохранены успешно: ", errorResponse.Error)
	}
}

// GetCardData получение данных банковской карты по названию для авторизованного пользователя
func GetCardData(cm *ClientManager) {
	var name string

	fmt.Print("Получаем данные карты. Введите имя в хранилище: ")
	fmt.Scan(&name)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if localMode {
		number, month, year, cardHolder, cvv, comment, err := localdb.GetCardData(ctx, name)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("Получены из локальной базы данные карты: номер %s, месяц %s, год %s, держатель карты %s, CVV %s, комментарий %s\n", number, month, year, cardHolder, cvv, comment)
		return
	}
	getCardDataResponse, err := cm.GetCardData(ctx, name)
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("Получены данные карты: номер %s, месяц %s, год %s, держатель карты %s, CVV %s, комментарий %s\n", getCardDataResponse.Number, getCardDataResponse.Month, getCardDataResponse.Year, getCardDataResponse.CardHolder, getCardDataResponse.Cvv, getCardDataResponse.Comment)
	}
}

// GetAllSavedDataNames получение всех названий сохранений
func GetAllSavedDataNames(cm *ClientManager) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer cancel()
	if localMode {
		names, err := localdb.GetAllSavedDataNames(ctx)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Получены из локальной БД names:")
		for _, name := range names {
			fmt.Println(name)
		}
		return
	}
	getAllSavedDataNamesResponse, err := cm.GetAllSavedDataNames(ctx)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("список сохраненных имен:")
		for _, n := range getAllSavedDataNamesResponse.SavedDataNames {
			fmt.Println(n)
		}
	}
}

// DelRawData удаляет сырые данные
func DelRawData(cm *ClientManager) {
	if localMode {
		fmt.Println("Удаленный сервер недоступен. Возможно только чтение из локальной БД.")
		return
	}

	var name string
	fmt.Print("Удаляем сырые данные. Введите имя в хранилище: ")
	fmt.Scan(&name)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	errorResponse, err := cm.DelRawData(ctx, name)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Сырые данные удалены успешно: ", errorResponse.Error)
	}
}

// DelLoginWithPassword удаляет данные логин, пароль
func DelLoginWithPassword(cm *ClientManager) {
	if localMode {
		fmt.Println("Удаленный сервер недоступен. Возможно только чтение из локальной БД.")
		return
	}

	var name string
	fmt.Print("Удаляем данные логин, пароль. Введите имя в хранилище: ")
	fmt.Scan(&name)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	errorResponse, err := cm.DelLoginWithPassword(ctx, name)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Данные логин, пароль удалены успешно: ", errorResponse.Error)
	}
}

// DelBinaryData удаляет бинарные данные
func DelBinaryData(cm *ClientManager) {
	if localMode {
		fmt.Println("Удаленный сервер недоступен. Возможно только чтение из локальной БД.")
		return
	}

	var name string
	fmt.Print("Удаляем бинарные данные. Введите имя в хранилище: ")
	fmt.Scan(&name)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	errorResponse, err := cm.DelBinaryData(ctx, name)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Бинарные данные удалены успешно: ", errorResponse.Error)
	}
}

// DelCardData удаляет данные карты
func DelCardData(cm *ClientManager) {
	if localMode {
		fmt.Println("Удаленный сервер недоступен. Возможно только чтение из локальной БД.")
		return
	}

	var name string
	fmt.Print("Удаляем данные карты. Введите имя в хранилище: ")
	fmt.Scan(&name)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	errorResponse, err := cm.DelCardData(ctx, name)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Данные карты удалены успешно: ", errorResponse.Error)
	}
}

// UpdRawData обновляет произвольную текстовую информацию для авторизованного пользователя
func UpdRawData(cm *ClientManager) {
	if localMode {
		fmt.Println("Удаленный сервер недоступен. Возможно только чтение из локальной БД.")
		return
	}

	var name string
	var data string
	var comment string

	fmt.Print("Обновляем сырые данные, строка. Введите имя в хранилище: ")
	fmt.Scan(&name)
	fmt.Print("Введите сохраняемую строку: ")
	fmt.Scan(&data)
	fmt.Print("Введите комментарий: ")
	fmt.Scan(&comment)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	errorResponse, err := cm.UpdRawData(ctx, name, data, comment)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Сырые данные обновлены успешно: ", errorResponse.Error)
	}
}

// UpdLoginWithPassword обновляет логин и пароль для авторизованного пользователя
func UpdLoginWithPassword(cm *ClientManager) {
	if localMode {
		fmt.Println("Удаленный сервер недоступен. Возможно только чтение из локальной БД.")
		return
	}

	var name string
	var login string
	var password string
	var comment string
	fmt.Print("Обновляем логин, пароль. Введите имя в хранилище: ")
	fmt.Scan(&name)
	fmt.Print("Введите сохраняемый логин: ")
	fmt.Scan(&login)
	fmt.Print("Введите сохраняемый пароль: ")
	fmt.Scan(&password)
	fmt.Print("Введите комментарий: ")
	fmt.Scan(&comment)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	errorResponse, err := cm.UpdLoginWithPassword(ctx, name, login, password, comment)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Логин, пароль обновлены успешно: ", errorResponse.Error)
	}
}

// SaveBinaryData обновляет произвольных бинарных данных для авторизованного пользователя
func UpdBinaryData(cm *ClientManager) {
	if localMode {
		fmt.Println("Удаленный сервер недоступен. Возможно только чтение из локальной БД.")
		return
	}

	var name string
	var filePath string
	var myBinary []byte
	var comment string
	fmt.Print("Обновляем бинарные данные. Введите имя в хранилище: ")
	fmt.Scan(&name)
	fmt.Print("Введите полный путь к файлу: ")
	fmt.Scan(&filePath)
	fmt.Print("Введите комментарий: ")
	fmt.Scan(&comment)

	myBinary, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	errorResponse, err := cm.UpdBinaryData(ctx, name, myBinary, comment)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Бинарные данные обновлены успешно: ", errorResponse.Error)
	}
}

// UpdCardData обновляет данных банковской карты для авторизованного пользователя
func UpdCardData(cm *ClientManager) {
	if localMode {
		fmt.Println("Удаленный сервер недоступен. Возможно только чтение из локальной БД.")
		return
	}

	var name string
	var number string
	var month string
	var year string
	var cardHolder string
	var cvv string
	var comment string

	fmt.Print("Обновляем данные карты. Введите имя в хранилище: ")
	fmt.Scan(&name)
	fmt.Print("Введите номер карты: ")
	fmt.Scan(&number)
	fmt.Print("Введите месяц выдачи карты: ")
	fmt.Scan(&month)
	fmt.Print("Введите год выдачи карты: ")
	fmt.Scan(&year)
	fmt.Print("Введите держателя карты: ")
	fmt.Scan(&cardHolder)
	fmt.Print("Введите CVV карты: ")
	fmt.Scan(&cvv)
	fmt.Print("Введите комментарий: ")
	fmt.Scan(&comment)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	errorResponse, err := cm.UpdCardData(ctx, name, number, month, year, cardHolder, cvv, comment)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Данные карты обновлены успешно: ", errorResponse.Error)
	}
}

// getLoginPassword получение экземпляра структуры UserClaims из JWT токена
func getLoginPassword(tokenString, key string) (*UserClaims, error) {
	userClaims := &UserClaims{}
	token, err := jwt.ParseWithClaims(tokenString, userClaims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(key), nil
		})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		fmt.Println("Token is not valid")
		return nil, err
	}

	return userClaims, nil
}

// CreateUser метод сохранения нового пользователя
func (cm *ClientManager) CreateUser(ctx context.Context) error {
	authorizedResponse, err := cm.GrpcClient.CreateUser(ctx, &pb.UserRegisterRequest{Login: cm.Lgn, Password: cm.Psw})
	if err != nil {
		return err
	}
	cm.UserClaims, err = getLoginPassword(authorizedResponse.Token, cm.ServerSettings.Key)
	if err != nil {
		return errors.New("ERROR: token not valid")
	}
	if cm.UserClaims.Login != cm.Lgn {
		return errors.New("ERROR: login not valid")
	} else {
		cm.UserID = cm.UserClaims.UserID
	}
	return nil
}

// LoginUser метод авторизации существующего пользователя
func (cm *ClientManager) LoginUser(ctx context.Context) error {
	authorizedResponse, err := cm.GrpcClient.LoginUser(ctx, &pb.UserAuthorizedRequest{Login: cm.Lgn, Password: cm.Psw})
	if err != nil {
		return err
	}
	cm.UserClaims, err = getLoginPassword(authorizedResponse.Token, cm.ServerSettings.Key)
	if err != nil {
		return errors.New("ERROR: token not valid")
	}
	if cm.UserClaims.Login != cm.Lgn {
		return errors.New("ERROR: login not valid")
	} else {
		cm.UserID = cm.UserClaims.UserID
	}
	return nil
}

// SaveRawData метод сохранения произвольной текстовой информации для авторизованного пользователя
func (cm *ClientManager) SaveRawData(ctx context.Context, name, data, comment string) (*pb.ErrorResponse, error) {
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	errorResponse, err := cm.GrpcClient.SaveRawData(ctx, &pb.SaveRawDataRequest{Name: name, Data: data, Comment: comment})
	if err != nil {
		return nil, err
	}
	err = localdb.SaveRawData(ctx, name, data, comment, cm.UserID)
	if err != nil {
		return nil, err
	}

	return errorResponse, nil
}

// GetRawData метод получения текстовой информации по названию для авторизованного пользователя
func (cm *ClientManager) GetRawData(ctx context.Context, name string) (*pb.GetRawDataResponse, error) {
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	getRawDataResponse, err := cm.GrpcClient.GetRawData(ctx, &pb.GetRawDataRequest{Name: name})
	if err != nil {
		return nil, err
	}
	return getRawDataResponse, nil
}

// SaveLoginWithPassword метод сохранения логина и пароля для авторизованного пользователя
func (cm *ClientManager) SaveLoginWithPassword(ctx context.Context, name, lgn, psw, comment string) (*pb.ErrorResponse, error) {
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	errorResponse, err := cm.GrpcClient.SaveLoginWithPassword(ctx, &pb.SaveLoginWithPasswordRequest{Name: name, Login: lgn, Password: psw, Comment: comment})
	if err != nil {
		return nil, err
	}

	err = localdb.SaveLoginWithPassword(ctx, name, lgn, psw, comment, cm.UserID)
	if err != nil {
		return nil, err
	}
	return errorResponse, nil
}

// GetLoginWithPassword метод получения логина и пароля по названию для авторизованного пользователя
func (cm *ClientManager) GetLoginWithPassword(ctx context.Context, name string) (*pb.GetLoginWithPasswordResponse, error) {
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	getLoginWithPasswordResponse, err := cm.GrpcClient.GetLoginWithPassword(ctx, &pb.GetLoginWithPasswordRequest{Name: name})
	if err != nil {
		return nil, err
	}
	return getLoginWithPasswordResponse, nil
}

// SaveBinaryData метод сохранения произвольных бинарных данных для авторизованного пользователя
func (cm *ClientManager) SaveBinaryData(ctx context.Context, name string, binData []byte, comment string) (*pb.ErrorResponse, error) {
	//return nil, nil
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	errorResponse, err := cm.GrpcClient.SaveBinaryData(ctx, &pb.SaveBinaryDataRequest{Name: name, Data: binData, Comment: comment})
	if err != nil {
		return nil, err
	}
	err = localdb.SaveBinaryData(ctx, name, binData, comment, cm.UserID)
	if err != nil {
		return nil, err
	}
	return errorResponse, nil
}

// GetBinaryData получения произвольных бинарных данных по названию для авторизованного пользователя
func (cm *ClientManager) GetBinaryData(ctx context.Context, name string) (*pb.GetBinaryDataResponse, error) {
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	getBinaryDataResponse, err := cm.GrpcClient.GetBinaryData(ctx, &pb.GetBinaryDataRequest{Name: name})
	if err != nil {
		return nil, err
	}
	return getBinaryDataResponse, nil
}

// SaveCardData метод сохранения данных банковской карты для авторизованного пользователя
func (cm *ClientManager) SaveCardData(ctx context.Context, name, number, month, year, cardHolder, cvv, comment string) (*pb.ErrorResponse, error) {
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	errorResponse, err := cm.GrpcClient.SaveCardData(ctx, &pb.SaveCardDataRequest{Name: name, Number: number, Month: month, Year: year, CardHolder: cardHolder, Cvv: cvv, Comment: comment})
	if err != nil {
		return nil, err
	}
	err = localdb.SaveCardData(ctx, name, number, month, year, cardHolder, cvv, comment, cm.UserID)
	if err != nil {
		return nil, err
	}
	return errorResponse, nil
}

// GetCardData метод получения данных банковской карты по названию для авторизованного пользователя
func (cm *ClientManager) GetCardData(ctx context.Context, name string) (*pb.GetCardDataResponse, error) {
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	getCardDataResponse, err := cm.GrpcClient.GetCardData(ctx, &pb.GetCardDataRequest{Name: name})
	if err != nil {
		return nil, err
	}
	return getCardDataResponse, nil
}

// GetAllSavedDataNames метод для получения всех названий сохранений
func (cm *ClientManager) GetAllSavedDataNames(ctx context.Context) (*pb.GetAllSavedDataNamesResponse, error) {
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	getAllSavedDataNamesResponse, err := cm.GrpcClient.GetAllSavedDataNames(ctx, &pb.GetAllSavedDataNamesRequest{})
	if err != nil {
		return nil, err
	}
	return getAllSavedDataNamesResponse, nil
}

// DelRawData удаляет сырые данные
func (cm *ClientManager) DelRawData(ctx context.Context, name string) (*pb.ErrorResponse, error) {
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	errorResponse, err := cm.GrpcClient.DelRawData(ctx, &pb.DelRequest{Name: name})
	if err != nil {
		return nil, err
	}
	err = localdb.DelRawData(ctx, name)
	if err != nil {
		return nil, err
	}
	return errorResponse, nil
}

// DelLoginWithPassword удаляет данные с логин, паролем
func (cm *ClientManager) DelLoginWithPassword(ctx context.Context, name string) (*pb.ErrorResponse, error) {
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	errorResponse, err := cm.GrpcClient.DelLoginWithPassword(ctx, &pb.DelRequest{Name: name})
	if err != nil {
		return nil, err
	}
	err = localdb.DelLoginWithPassword(ctx, name)
	if err != nil {
		return nil, err
	}
	return errorResponse, nil
}

// DelBinaryData удаляет бинарные данные
func (cm *ClientManager) DelBinaryData(ctx context.Context, name string) (*pb.ErrorResponse, error) {
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	errorResponse, err := cm.GrpcClient.DelBinaryData(ctx, &pb.DelRequest{Name: name})
	if err != nil {
		return nil, err
	}
	err = localdb.DelBinaryData(ctx, name)
	if err != nil {
		return nil, err
	}
	return errorResponse, nil
}

// DelCardData удаляет данные карты
func (cm *ClientManager) DelCardData(ctx context.Context, name string) (*pb.ErrorResponse, error) {
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	errorResponse, err := cm.GrpcClient.DelCardData(ctx, &pb.DelRequest{Name: name})
	if err != nil {
		return nil, err
	}
	err = localdb.DelCardData(ctx, name)
	if err != nil {
		return nil, err
	}
	return errorResponse, nil
}

// UpdRawData метод обновления произвольной текстовой информации для авторизованного пользователя
func (cm *ClientManager) UpdRawData(ctx context.Context, name, data, comment string) (*pb.ErrorResponse, error) {
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	errorResponse, err := cm.GrpcClient.UpdRawData(ctx, &pb.SaveRawDataRequest{Name: name, Data: data, Comment: comment})
	if err != nil {
		return nil, err
	}
	err = localdb.UpdRawData(ctx, name, data, comment, cm.UserID)
	if err != nil {
		return nil, err
	}
	return errorResponse, nil
}

// UpdLoginWithPassword метод обновления логина и пароля для авторизованного пользователя
func (cm *ClientManager) UpdLoginWithPassword(ctx context.Context, name, lgn, psw, comment string) (*pb.ErrorResponse, error) {
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	errorResponse, err := cm.GrpcClient.UpdLoginWithPassword(ctx, &pb.SaveLoginWithPasswordRequest{Name: name, Login: lgn, Password: psw, Comment: comment})
	if err != nil {
		return nil, err
	}
	err = localdb.UpdLoginWithPassword(ctx, name, lgn, psw, comment, cm.UserID)
	if err != nil {
		return nil, err
	}
	return errorResponse, nil
}

// UpdBinaryData метод обновления произвольных бинарных данных для авторизованного пользователя
func (cm *ClientManager) UpdBinaryData(ctx context.Context, name string, binData []byte, comment string) (*pb.ErrorResponse, error) {
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	errorResponse, err := cm.GrpcClient.UpdBinaryData(ctx, &pb.SaveBinaryDataRequest{Name: name, Data: binData, Comment: comment})
	if err != nil {
		return nil, err
	}
	err = localdb.UpdBinaryData(ctx, name, binData, comment, cm.UserID)
	if err != nil {
		return nil, err
	}
	return errorResponse, nil
}

// UpdCardData метод обновления данных банковской карты для авторизованного пользователя
func (cm *ClientManager) UpdCardData(ctx context.Context, name, number, month, year, cardHolder, cvv, comment string) (*pb.ErrorResponse, error) {
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	errorResponse, err := cm.GrpcClient.UpdCardData(ctx, &pb.SaveCardDataRequest{Name: name, Number: number, Month: month, Year: year, CardHolder: cardHolder, Cvv: cvv, Comment: comment})
	if err != nil {
		return nil, err
	}
	err = localdb.UpdCardData(ctx, name, number, month, year, cardHolder, cvv, comment, cm.UserID)
	if err != nil {
		return nil, err
	}
	return errorResponse, nil
}

func (cm *ClientManager) CheckConnectCall() (*pb.ErrorResponse, error) {
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, md)
	errorResponse, err := cm.GrpcClient.CheckConnectCall(ctx, &pb.CheckConnectRequest{})
	if err != nil {
		return nil, err
	}
	return errorResponse, nil
}

/*func checkConnection(conn *grpc.ClientConn) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	state := conn.GetState()
	switch state {
	case connectivity.Idle, connectivity.Connecting:
		// Connection is being established
		fmt.Println(state)
		return conn.WaitForStateChange(ctx, state)
	case connectivity.Ready:
		// Connection is active
		fmt.Println(state)
		return true
	case connectivity.TransientFailure, connectivity.Shutdown:
		// Connection failed or is shutting down
		fmt.Println(state)
		return false
	default:
		return false
	}
}

func checkHealth(conn *grpc.ClientConn) bool {
	client := grpc_health_v1.NewHealthClient(conn)
	resp, err := client.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{
		Service: "", // empty for overall server health
	})

	if err != nil {
		log.Printf("Health check failed: %v", err)
		return false
	}
	fmt.Println(resp.Status)
	return resp.Status == grpc_health_v1.HealthCheckResponse_SERVING
}*/
