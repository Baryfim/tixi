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
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", config.Cfg.DBUser, config.Cfg.DBPassword, config.Cfg.DBHost, config.Cfg.DBPort, config.Cfg.DBName)

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}

	log.Println("Подключение к базе данных успешно установлено")
	DBConn = conn
}
