package grpcserver

import (
	"context"

	//"os/signal"
	//"syscall"

	"github.com/developerc/GophKeeper/internal/config"
	//"github.com/developerc/GophKeeper/internal/grpcserver"
)

// Run метод запускает работу сервера и мягко останавливает.
func Run() error {
	/*ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()*/
	ctx := context.Background()

	settings, err := config.NewServerSettings()
	if err != nil {
		return err
	}
	//fmt.Println(settings)

	NewGRPCserver(ctx, settings)

	return nil
}
