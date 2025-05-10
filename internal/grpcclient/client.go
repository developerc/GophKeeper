package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/developerc/GophKeeper/internal/config"
	//"github.com/developerc/GophKeeper/internal/security"
	pb "github.com/developerc/GophKeeper/proto"
	"github.com/golang-jwt/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var serverSettings *config.ServerSettings

// main запускает клиента gRPC
func main() {
	//var userClaims *security.UserClaims
	var userClaims *UserClaims
	var err error
	var lgn string
	var psw string
	var userID string
	var clientJWTManager *JWTManager
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

	grpcClient := pb.NewGrpcServiceClient(conn)

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
	errorResponse, err := grpcClient.SaveRawData(ctx, &pb.SaveRawDataRequest{Name: "myName", Data: "my  raw data"})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(errorResponse.Error)
	}
}

// func getLoginPassword(tokenString string) (*security.UserClaims, error) {
func getLoginPassword(tokenString string) (*UserClaims, error) {
	//userClaims := &security.UserClaims{}
	userClaims := &UserClaims{}
	token, err := jwt.ParseWithClaims(tokenString, userClaims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(serverSettings.Key), nil
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
