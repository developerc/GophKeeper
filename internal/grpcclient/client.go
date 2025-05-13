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
	//"github.com/developerc/GophKeeper/internal/security"
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
	"11 Получение всех сохраненных имен\n"

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
		log.Println(err)
		os.Exit(1)
	}

	// создадим клиент grpc //с перехватчиком
	conn, err := grpc.NewClient(cm.ServerSettings.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println("could not connect to grpc server: ", err)
		os.Exit(1)
	}
	defer conn.Close()
	cm.GrpcClient = pb.NewGrpcServiceClient(conn)

	fmt.Print(menu)

	for {
		fmt.Println("Сделайте выбор меню:")
		fmt.Scan(&choice)
		//fmt.Println(choice)
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
		}
	}
	//-- вводим логин, пароль
	/*cm.Lgn = "myLogin"
	cm.Psw = "myPassword"
	//-- регистрируем юзера
	//fmt.Println(cm)
	err = cm.CreateUser(ctx)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Регистрация прошла успешно, userID: ", cm.UserID)
	}

	//-- аутентификация
	cm.Lgn = "myLogin"
	cm.Psw = "myPassword"
	err = cm.LoginUser(ctx)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Аутентификация прошла успешно, userID: ", cm.UserID)
	}

	// сохраняем сырые данные строка
	name := "RawData1"
	myData := "my_Raw_Data"
	errorResponse, err := cm.SaveRawData(ctx, name, myData)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Сырые данные сохранены успешно")
		fmt.Println("Сырые данные сохранены успешно: ", errorResponse.Error)
	}

	// получаем сырые данные строка
	name = "RawData1"
	getRawDataResponse, err := cm.GetRawData(ctx, name)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Получены сырые данные: ", getRawDataResponse.Data)
	}

	// сохраняем логин, пароль
	name = "LgnPsw1"
	login := "myLogin1"
	password := "myPassword1"
	errorResponse, err = cm.SaveLoginWithPassword(ctx, name, login, password)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Логин, пароль сохранены успешно: ", errorResponse.Error)
	}

	// получаем логин, пароль
	name = "LgnPsw1"
	getLoginWithPasswordResponse, err := cm.GetLoginWithPassword(ctx, name)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Получены логин, пароль: ", getLoginWithPasswordResponse.Login, getLoginWithPasswordResponse.Password)
	}

	// сохраняем бинарные данные
	myBinary := []byte("my_binary_data")
	name = "binData1"
	errorResponse, err = cm.SaveBinaryData(ctx, name, myBinary)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Бинарные данные сохранены успешно: ", errorResponse.Error)
	}

	// получаем бинарные данные
	name = "binData1"
	getBinaryDataResponse, err := cm.GetBinaryData(ctx, name)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Получены бинарные данные, строка: ", string(getBinaryDataResponse.Data))
	}

	// сохраняем данные карты
	//{Name: "myCard1", Number: "1234-5670-8910-3451", Month: "May", Year: "2025", CardHolder: "МИР"}
	name = "myCard1"
	number := "1234-5670-8910-3451"
	month := "May"
	year := "2025"
	cardHolder := "МИР"
	errorResponse, err = cm.SaveCardData(ctx, name, number, month, year, cardHolder)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Данные карты сохранены успешно: ", errorResponse.Error)
	}

	//получаем данные карты
	name = "myCard1"
	getCardDataResponse, err := cm.GetCardData(ctx, name)
	if err != nil {
		log.Println(err)
	} else {
		//log.Println(getCardDataResponse)
		log.Printf("Получены данные карты: номер %s, месяц %s, год %s, держатель карты %s\n", getCardDataResponse.Number, getCardDataResponse.Month, getCardDataResponse.Year, getCardDataResponse.CardHolder)
	}

	// получаем все сохраненные имена
	getAllSavedDataNamesResponse, err := cm.GetAllSavedDataNames(ctx)
	if err != nil {
		log.Println(err)
	} else {
		//log.Println(getCardDataResponse)
		//log.Printf("Получены данные карты: номер %s, месяц %s, год %s, держатель карты %s\n", getCardDataResponse.Number, getCardDataResponse.Month, getCardDataResponse.Year, getCardDataResponse.CardHolder)
		log.Println("список сохраненных имен:")
		for _, n := range getAllSavedDataNamesResponse.SavedDataNames {
			fmt.Println(n)
		}
	}*/

	/*var userClaims *UserClaims
	var err error
	var lgn string
	var psw string
	var userID string
	var clientJWTManager *JWTManager
	var grpcClient pb.GrpcServiceClient
	serverSettings, err = config.NewServerSettings()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	clientJWTManager, err = NewJWTManager(serverSettings.Key, serverSettings.TokenDuration)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	//fmt.Println(serverSettings)
	addr := serverSettings.Host //"localhost:5000"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// создадим клиент grpc //с перехватчиком
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println("could not connect to grpc server: ", err)
		os.Exit(1)
	}
	defer conn.Close()

	grpcClient = pb.NewGrpcServiceClient(conn)

	// Зарегистрируем юзера, отправим логин, пассворд
	lgn = "myLogin"
	psw = "myPassword"
	authorizedResponse, err := grpcClient.CreateUser(ctx, &pb.UserRegisterRequest{Login: lgn, Password: psw})
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(authorizedResponse.Token)
	userClaims, err = getLoginPassword(authorizedResponse.Token)
	if err != nil {
		fmt.Println("ERROR: token not valid!")
	}
	if userClaims.Login != lgn {
		fmt.Println("ERROR: login not valid!")
	} else {
		userID = userClaims.UserID
		fmt.Println("Регистрация прошла успешно, userID: ", userID)
	}
	//fmt.Println(userClaims)

	// авторизуемся, отправим логин, пассворд, получим токен
	lgn = "myLogin"
	psw = "myPassword"
	authorizedResponse, err = grpcClient.LoginUser(ctx, &pb.UserAuthorizedRequest{Login: lgn, Password: psw})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(authorizedResponse.Token)
	}

	// отправляем сырые данные
	//jwtManager,err := &security.NewJWTManager(serverSettings.Key, serverSettings.TokenDuration)
	//jwtManager := security.JWTManager{&ServerSettings.Key, &ServerSettings.TokenDuration}
	jwtToken, err := clientJWTManager.GenerateJWT(userID, lgn)
	if err != nil {
		fmt.Println(err)
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	errorResponse, err := grpcClient.SaveRawData(ctx, &pb.SaveRawDataRequest{Name: "rawData1", Data: "my_raw_data_1"})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(errorResponse.Error)
	}

	// получаем сырые данные
	jwtToken, err = clientJWTManager.GenerateJWT(userID, lgn)
	if err != nil {
		fmt.Println(err)
	}
	md = metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	getRawDataResponse, err := grpcClient.GetRawData(ctx, &pb.GetRawDataRequest{Name: "rawData1"})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(getRawDataResponse.Data)
	}

	// сохраняем логин, пароль
	jwtToken, err = clientJWTManager.GenerateJWT(userID, lgn)
	if err != nil {
		fmt.Println(err)
	}
	md = metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	errorResponse, err = grpcClient.SaveLoginWithPassword(ctx, &pb.SaveLoginWithPasswordRequest{Name: "LgnPsw1", Login: "login1", Password: "password1"})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(errorResponse.Error)
	}

	// получаем сохраненные логин, пароль
	jwtToken, err = clientJWTManager.GenerateJWT(userID, lgn)
	if err != nil {
		fmt.Println(err)
	}
	md = metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	getLoginWithPasswordResponse, err := grpcClient.GetLoginWithPassword(ctx, &pb.GetLoginWithPasswordRequest{Name: "LgnPsw1"})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("login: %s, password: %s\n", getLoginWithPasswordResponse.Login, getLoginWithPasswordResponse.Password)
	}

	// сохраняем бинарные данные
	myBinary := []byte("my_binary_data")
	jwtToken, err = clientJWTManager.GenerateJWT(userID, lgn)
	if err != nil {
		fmt.Println(err)
	}
	md = metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	errorResponse, err = grpcClient.SaveBinaryData(ctx, &pb.SaveBinaryDataRequest{Name: "myBinary1", Data: myBinary})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(errorResponse.Error)
	}

	// получаем бинарные данные
	jwtToken, err = clientJWTManager.GenerateJWT(userID, lgn)
	if err != nil {
		fmt.Println(err)
	}
	md = metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	getBinaryDataResponse, err := grpcClient.GetBinaryData(ctx, &pb.GetBinaryDataRequest{Name: "myBinary1"})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(getBinaryDataResponse.Data))
	}

	// сохраняем данные карты
	jwtToken, err = clientJWTManager.GenerateJWT(userID, lgn)
	if err != nil {
		fmt.Println(err)
	}
	md = metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	errorResponse, err = grpcClient.SaveCardData(ctx, &pb.SaveCardDataRequest{Name: "myCard1", Number: "1234-5670-8910-3451", Month: "May", Year: "2025", CardHolder: "МИР"})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(errorResponse.Error)
	}

	// получаем данные карты
	jwtToken, err = clientJWTManager.GenerateJWT(userID, lgn)
	if err != nil {
		fmt.Println(err)
	}
	md = metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	getCardDataResponse, err := grpcClient.GetCardData(ctx, &pb.GetCardDataRequest{Name: "myCard1"})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Номер карты: %s, год: %s, месяц: %s, держатель карты: %s\n", getCardDataResponse.Number, getCardDataResponse.Year, getCardDataResponse.Month, getCardDataResponse.CardHolder)
	}

	// получаем все сохраненные имена
	jwtToken, err = clientJWTManager.GenerateJWT(userID, lgn)
	if err != nil {
		fmt.Println(err)
	}
	md = metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	getAllSavedDataNamesResponse, err := grpcClient.GetAllSavedDataNames(ctx, &pb.GetAllSavedDataNamesRequest{})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Список сохраненных имен:")
		for _, n := range getAllSavedDataNamesResponse.SavedDataNames {
			fmt.Println(n)
		}
	}*/
}
func exitApp() {
	var confirm string
	fmt.Println("Подтвердите выход y/n")
	fmt.Scan(&confirm)
	if confirm == "y" {
		os.Exit(0)
	}
}

func CreateUser(cm *ClientManager) {
	var lgn string
	var psw string

	fmt.Println("Регистрируем пользователя. Введите логин:")
	fmt.Scan(&lgn)
	fmt.Println("Введите пароль:")
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

func LoginUser(cm *ClientManager) {
	var lgn string
	var psw string

	fmt.Println("Проводим аутентификацию. Введите логин:")
	fmt.Scan(&lgn)
	fmt.Println("Введите пароль:")
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

func SaveRawData(cm *ClientManager) {
	var name string
	var data string

	fmt.Println("Добавляем сырые данные, строка. Введите имя в хранилище:")
	fmt.Scan(&name)
	fmt.Println("Введите сохраняемую строку:")
	fmt.Scan(&data)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	errorResponse, err := cm.SaveRawData(ctx, name, data)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Сырые данные сохранены успешно: ", errorResponse.Error)
	}
}

func GetRawData(cm *ClientManager) {
	var name string
	fmt.Println("Получаем сырые данные, строка. Введите имя в хранилище:")
	fmt.Scan(&name)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	getRawDataResponse, err := cm.GetRawData(ctx, name)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Получены сырые данные: ", getRawDataResponse.Data)
	}
}

func SaveLoginWithPassword(cm *ClientManager) {
	var name string
	var login string
	var password string
	fmt.Println("Добавляем логин, пароль. Введите имя в хранилище:")
	fmt.Scan(&name)
	fmt.Println("Введите сохраняемый логин:")
	fmt.Scan(&login)
	fmt.Println("Введите сохраняемый пароль:")
	fmt.Scan(&password)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	errorResponse, err := cm.SaveLoginWithPassword(ctx, name, login, password)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Логин, пароль сохранены успешно: ", errorResponse.Error)
	}
}

func GetLoginWithPassword(cm *ClientManager) {
	var name string
	fmt.Println("Получаем логин, пароль. Введите имя в хранилище:")
	fmt.Scan(&name)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	getLoginWithPasswordResponse, err := cm.GetLoginWithPassword(ctx, name)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Получены логин, пароль: ", getLoginWithPasswordResponse.Login, getLoginWithPasswordResponse.Password)
	}
}

func SaveBinaryData(cm *ClientManager) {
	var name string
	var myBinaryStr string
	var myBinary []byte
	//myBinary := []byte("my_binary_data")
	//name = "binData1"
	fmt.Println("Добавляем бинарные данные. Введите имя в хранилище:")
	fmt.Scan(&name)
	fmt.Println("Введите сохраняемые бинарные данные как строку:")
	fmt.Scan(&myBinaryStr)
	myBinary = []byte(myBinaryStr)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	errorResponse, err := cm.SaveBinaryData(ctx, name, myBinary)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Бинарные данные сохранены успешно: ", errorResponse.Error)
	}
}

func GetBinaryData(cm *ClientManager) {
	var name string
	fmt.Println("Получаем бинарные данные. Введите имя в хранилище:")
	fmt.Scan(&name)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	getBinaryDataResponse, err := cm.GetBinaryData(ctx, name)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Получены бинарные данные, строка: ", string(getBinaryDataResponse.Data))
	}
}

func SaveCardData(cm *ClientManager) {
	var name string
	var number string
	var month string
	var year string
	var cardHolder string

	fmt.Println("Добавляем данные карты. Введите имя в хранилище:")
	fmt.Scan(&name)
	fmt.Println("Введите номер карты:")
	fmt.Scan(&number)
	fmt.Println("Введите месяц выдачи карты:")
	fmt.Scan(&month)
	fmt.Println("Введите год выдачи карты:")
	fmt.Scan(&year)
	fmt.Println("Введите держателя карты:")
	fmt.Scan(&cardHolder)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	errorResponse, err := cm.SaveCardData(ctx, name, number, month, year, cardHolder)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Данные карты сохранены успешно: ", errorResponse.Error)
	}
}

func GetCardData(cm *ClientManager) {
	var name string
	//name = "myCard1"
	fmt.Println("Получаем данные карты. Введите имя в хранилище:")
	fmt.Scan(&name)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	getCardDataResponse, err := cm.GetCardData(ctx, name)
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("Получены данные карты: номер %s, месяц %s, год %s, держатель карты %s\n", getCardDataResponse.Number, getCardDataResponse.Month, getCardDataResponse.Year, getCardDataResponse.CardHolder)
	}
}

func GetAllSavedDataNames(cm *ClientManager) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
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

// func getLoginPassword(tokenString string) (*security.UserClaims, error) {
func getLoginPassword(tokenString, key string) (*UserClaims, error) {
	//userClaims := &security.UserClaims{}
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

func (cm *ClientManager) CreateUser(ctx context.Context) error {
	authorizedResponse, err := cm.GrpcClient.CreateUser(ctx, &pb.UserRegisterRequest{Login: cm.Lgn, Password: cm.Psw})
	//fmt.Println(authorizedResponse.Token)
	if err != nil {
		return err
	}
	cm.UserClaims, err = getLoginPassword(authorizedResponse.Token, cm.ServerSettings.Key)
	if err != nil {
		return errors.New("ERROR: token not valid")
		//fmt.Println("ERROR: token not valid!")
	}
	if cm.UserClaims.Login != cm.Lgn {
		return errors.New("ERROR: login not valid")
		//fmt.Println("ERROR: login not valid!")
	} else {
		cm.UserID = cm.UserClaims.UserID
		//fmt.Println("Регистрация прошла успешно, userID: ", userID)
	}
	return nil
}

func (cm *ClientManager) LoginUser(ctx context.Context) error {
	authorizedResponse, err := cm.GrpcClient.LoginUser(ctx, &pb.UserAuthorizedRequest{Login: cm.Lgn, Password: cm.Psw})
	if err != nil {
		return err
		//fmt.Println(err)
	}
	//fmt.Println(authorizedResponse.Token)
	cm.UserClaims, err = getLoginPassword(authorizedResponse.Token, cm.ServerSettings.Key)
	if err != nil {
		return errors.New("ERROR: token not valid")
		//fmt.Println("ERROR: token not valid!")
	}
	if cm.UserClaims.Login != cm.Lgn {
		return errors.New("ERROR: login not valid")
		//fmt.Println("ERROR: login not valid!")
	} else {
		cm.UserID = cm.UserClaims.UserID
		//fmt.Println("Регистрация прошла успешно, userID: ", userID)
	}
	return nil
}

func (cm *ClientManager) SaveRawData(ctx context.Context, name, data string) (*pb.ErrorResponse, error) {
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
		//fmt.Println(err)
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	errorResponse, err := cm.GrpcClient.SaveRawData(ctx, &pb.SaveRawDataRequest{Name: name, Data: data})
	if err != nil {
		return nil, err
		//fmt.Println(err)
	} /*else {
		fmt.Println("from SaveRawData: ", errorResponse.Error)
	}*/
	return errorResponse, nil
}

func (cm *ClientManager) GetRawData(ctx context.Context, name string) (*pb.GetRawDataResponse, error) {
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
		//fmt.Println(err)
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	getRawDataResponse, err := cm.GrpcClient.GetRawData(ctx, &pb.GetRawDataRequest{Name: name})
	if err != nil {
		return nil, err
	} /*else {
		fmt.Println(getRawDataResponse.Data)
	}*/
	return getRawDataResponse, nil
}

func (cm *ClientManager) SaveLoginWithPassword(ctx context.Context, name, lgn, psw string) (*pb.ErrorResponse, error) {
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
		//fmt.Println(err)
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	errorResponse, err := cm.GrpcClient.SaveLoginWithPassword(ctx, &pb.SaveLoginWithPasswordRequest{Name: name, Login: lgn, Password: psw})
	if err != nil {
		return nil, err
	}
	return errorResponse, nil
}

func (cm *ClientManager) GetLoginWithPassword(ctx context.Context, name string) (*pb.GetLoginWithPasswordResponse, error) {
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
		//fmt.Println(err)
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	getLoginWithPasswordResponse, err := cm.GrpcClient.GetLoginWithPassword(ctx, &pb.GetLoginWithPasswordRequest{Name: name})
	if err != nil {
		return nil, err
	}
	return getLoginWithPasswordResponse, nil
}

func (cm *ClientManager) SaveBinaryData(ctx context.Context, name string, binData []byte) (*pb.ErrorResponse, error) {
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
		//fmt.Println(err)
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	errorResponse, err := cm.GrpcClient.SaveBinaryData(ctx, &pb.SaveBinaryDataRequest{Name: name, Data: binData})
	if err != nil {
		return nil, err
	}
	return errorResponse, nil
}

func (cm *ClientManager) GetBinaryData(ctx context.Context, name string) (*pb.GetBinaryDataResponse, error) {
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
		//fmt.Println(err)
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	getBinaryDataResponse, err := cm.GrpcClient.GetBinaryData(ctx, &pb.GetBinaryDataRequest{Name: name})
	if err != nil {
		return nil, err
	}
	return getBinaryDataResponse, nil
}

func (cm *ClientManager) SaveCardData(ctx context.Context, name, number, month, year, cardHolder string) (*pb.ErrorResponse, error) {
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
		//fmt.Println(err)
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	errorResponse, err := cm.GrpcClient.SaveCardData(ctx, &pb.SaveCardDataRequest{Name: name, Number: number, Month: month, Year: year, CardHolder: cardHolder})
	if err != nil {
		return nil, err
	}
	return errorResponse, nil
}

func (cm *ClientManager) GetCardData(ctx context.Context, name string) (*pb.GetCardDataResponse, error) {
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
		//fmt.Println(err)
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	getCardDataResponse, err := cm.GrpcClient.GetCardData(ctx, &pb.GetCardDataRequest{Name: name})
	if err != nil {
		return nil, err
	}
	return getCardDataResponse, nil
}

func (cm *ClientManager) GetAllSavedDataNames(ctx context.Context) (*pb.GetAllSavedDataNamesResponse, error) {
	jwtToken, err := cm.ClientJWTManager.GenerateJWT(cm.UserID, cm.Lgn)
	if err != nil {
		return nil, err
		//fmt.Println(err)
	}
	md := metadata.New(map[string]string{"authorization": jwtToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	getAllSavedDataNamesResponse, err := cm.GrpcClient.GetAllSavedDataNames(ctx, &pb.GetAllSavedDataNamesRequest{})
	if err != nil {
		return nil, err
	} /*else {
		fmt.Println("Список сохраненных имен:")
		for _, n := range getAllSavedDataNamesResponse.SavedDataNames {
			fmt.Println(n)
		}
	}*/
	return getAllSavedDataNamesResponse, nil
}
