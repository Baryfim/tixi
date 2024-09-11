package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tixiby/internal/config"
	"github.com/tixiby/internal/db"
	"github.com/tixiby/internal/server"
	"github.com/tixiby/pkg/sql"
)

func main() {
	config.InitConfig()

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

	if err := server.RunGRPCServer(); err != nil {
		log.Fatalf("Ошибка запуска gRPC сервера: %v", err)
	}
}
