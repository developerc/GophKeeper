package main

import (
	"log"

	"github.com/developerc/GophKeeper/internal/server"
)

// main главная функция запуска приложения
func main() {
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
