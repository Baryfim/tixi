package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/tixiby/internal/config"
)

var DBConn *pgx.Conn

func InitDB() {
	cfg := config.Cfg
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}

	log.Println("Подключение к базе данных успешно установлено")
	DBConn = conn
}
