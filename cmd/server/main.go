package main

import (
	"log"

	"github.com/tixiby/internal/config"
	"github.com/tixiby/internal/db"
	"github.com/tixiby/internal/server"
)

func main() {
	config.InitConfig()

	db.InitDB()

	if err := server.RunGRPCServer(); err != nil {
		log.Fatalf("Ошибка запуска gRPC сервера: %v", err)
	}
}
