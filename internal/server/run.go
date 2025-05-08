package server

import (
	"fmt"

	"github.com/developerc/GophKeeper/internal/config"
)

// Run метод запускает работу сервера и мягко останавливает.
func Run() error {

	settings, err := config.NewServerSettings()
	if err != nil {
		return err
	}
	fmt.Println(settings)

	return nil
}
