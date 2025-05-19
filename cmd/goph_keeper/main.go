package main

import (
	"log"

	"github.com/developerc/GophKeeper/internal/grpcserver"
)

// main главная функция запуска приложения
func main() {
	if err := grpcserver.Run(); err != nil {
		log.Fatal(err)
	}
}
