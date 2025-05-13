package grpcserver

import (
	"context"

	"os/signal"
	"syscall"

	"github.com/developerc/GophKeeper/internal/config"
	"github.com/developerc/GophKeeper/internal/repositories"
	"github.com/developerc/GophKeeper/internal/repositories/datarepository"
	"github.com/developerc/GophKeeper/internal/repositories/userrepository"
	"github.com/developerc/GophKeeper/internal/security"
	"github.com/developerc/GophKeeper/internal/service/dataservice"
	"github.com/developerc/GophKeeper/internal/service/userservice"
)

var beforeStop chan struct{}

// Run метод запускает работу сервера и мягко останавливает.
func Run() error {
	beforeStop = make(chan struct{})
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	//ctx := context.Background()

	settings, err := config.NewServerSettings()
	if err != nil {
		return err
	}

	//fmt.Println(settings)
	db, err := repositories.InitDB(ctx, settings.DataBaseDsn)
	if err != nil {
		return err
	}

	userRepository := userrepository.New(db)
	rawDataRepository := datarepository.New(db)
	userService := userservice.New(userRepository)
	jwtManager, err := security.NewJWTManager(settings.Key, settings.TokenDuration)
	if err != nil {
		return err
	}
	//fmt.Println(jwtManager)
	cipherManager, err := security.NewCipherManager(settings.Key)
	if err != nil {
		return err
	}
	storageService := dataservice.New(rawDataRepository, cipherManager)

	NewGRPCserver(ctx, settings, userService, jwtManager, storageService, db)
	<-beforeStop
	return nil
}
