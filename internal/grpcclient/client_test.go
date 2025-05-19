// main пакет для клиента
// Интеграционный тест. Запускается после запуска основного серверного приложения.
// Перед запуском серверного приложения удалить таблицы raw_data и users что бы при запуске они создались пустыми.
// Проверяется работа всех функций клиента.
package main

import (
	"context"
	"fmt"

	"log"
	"os"
	"testing"
	"time"

	pb "github.com/developerc/GophKeeper/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestClient(t *testing.T) {
	var err error
	cm, err := NewClientManager()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	conn, err := grpc.NewClient(cm.ServerSettings.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println("could not connect to grpc server: ", err)
		os.Exit(1)
	}
	defer conn.Close()
	cm.GrpcClient = pb.NewGrpcServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("#1_RegisterUserTest", func(t *testing.T) {
		cm.Lgn = "myLogin"
		cm.Psw = "myPassword"
		err = cm.CreateUser(ctx)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Регистрация прошла успешно, userID: ", cm.UserID)
		}
	})

	t.Run("#2_AuthTest", func(t *testing.T) {
		cm.Lgn = "myLogin"
		cm.Psw = "myPassword"
		err = cm.LoginUser(ctx)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Аутентификация прошла успешно, userID: ", cm.UserID)
		}
	})

	t.Run("#3_SaveRawDataTest", func(t *testing.T) {
		name := "RawData1"
		myData := "my_Raw_Data"
		comment := "my_Comment"
		errorResponse, err := cm.SaveRawData(ctx, name, myData, comment)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Сырые данные сохранены успешно: ", errorResponse.Error)
		}
	})

	t.Run("#4_GetRawDataTest", func(t *testing.T) {
		name := "RawData1"
		getRawDataResponse, err := cm.GetRawData(ctx, name)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Получены сырые данные: ", getRawDataResponse.Data)
		}
	})

	t.Run("#5_SaveLgnPswTest", func(t *testing.T) {
		name := "LgnPsw1"
		login := "myLogin1"
		password := "myPassword1"
		comment := "my_comment"
		errorResponse, err := cm.SaveLoginWithPassword(ctx, name, login, password, comment)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Логин, пароль сохранены успешно: ", errorResponse.Error)
		}
	})

	t.Run("#6_GetLgnPswTest", func(t *testing.T) {
		name := "LgnPsw1"
		getLoginWithPasswordResponse, err := cm.GetLoginWithPassword(ctx, name)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Получены логин, пароль: ", getLoginWithPasswordResponse.Login, getLoginWithPasswordResponse.Password)
		}
	})

	t.Run("#7_SaveBinaryTest", func(t *testing.T) {
		myBinary := []byte("my_binary_data")
		name := "binData1"
		comment := "my_comment"
		errorResponse, err := cm.SaveBinaryData(ctx, name, myBinary, comment)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Бинарные данные сохранены успешно: ", errorResponse.Error)
		}
	})

	t.Run("#8_GetBinaryTest", func(t *testing.T) {
		name := "binData1"
		getBinaryDataResponse, err := cm.GetBinaryData(ctx, name)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Получены бинарные данные, строка: ", string(getBinaryDataResponse.Data))
		}
	})

	t.Run("#9_SaveCardTest", func(t *testing.T) {
		name := "myCard1"
		number := "1234-5670-8910-3451"
		month := "May"
		year := "2025"
		cardHolder := "МИР"
		cvv := "123"
		comment := "my_comment"
		errorResponse, err := cm.SaveCardData(ctx, name, number, month, year, cardHolder, cvv, comment)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Данные карты сохранены успешно: ", errorResponse.Error)
		}
	})

	t.Run("#10_GetCardTest", func(t *testing.T) {
		name := "myCard1"
		getCardDataResponse, err := cm.GetCardData(ctx, name)
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("Получены данные карты: номер %s, месяц %s, год %s, держатель карты %s, CVV %s\n", getCardDataResponse.Number, getCardDataResponse.Month, getCardDataResponse.Year, getCardDataResponse.CardHolder, getCardDataResponse.Cvv)
		}
	})

	t.Run("#11_GetNamesTest", func(t *testing.T) {
		getAllSavedDataNamesResponse, err := cm.GetAllSavedDataNames(ctx)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("список сохраненных имен:")
			for _, n := range getAllSavedDataNamesResponse.SavedDataNames {
				fmt.Println(n)
			}
		}
	})

}
