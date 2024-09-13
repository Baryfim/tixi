package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tixiby/internal/config"
	"github.com/tixiby/internal/db"
	"github.com/tixiby/internal/server/grpc"
	"github.com/tixiby/internal/server/rest"
	"github.com/tixiby/pkg/sql"
)

func main() {
	config.InitConfig()

	// Згружаем базу данных и выполняем миграции
	db.InitDB()
	ctx := context.Background()
	query, err := sql.LoadSQLFile("init-tables.sql")
	if err != nil {
		log.Fatalf("Ошибка загрузки SQL-запроса: %v", err)
	}

	_, err = db.DBConn.Exec(ctx, query)
	if err != nil {
		fmt.Printf("Не удалось запустить базу данных: %v", err)
	}

	go func() {
		if err := grpc.RunGRPCServer(); err != nil {
			log.Fatalf("Ошибка запуска gRPC сервера: %v", err)
		}
	}()

	go func() {
		if err = rest.RunRESTServer(); err != nil {
			log.Fatalf("Ошибка запуска REST сервера: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Завершается работа сервера...")
}
