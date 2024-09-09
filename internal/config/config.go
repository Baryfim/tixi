package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	GRPCPort   string `mapstructure:"grpc_port"`
	JWTSecret  string `mapstructure:"jwt_secret"`
	SSLCert    string `mapstructure:"ssl_cert"`
	SSLKey     string `mapstructure:"ssl_key"`
	DBHost     string `mapstructure:"db_host"`
	DBPort     string `mapstructure:"db_port"`
	DBUser     string `mapstructure:"db_user"`
	DBPassword string `mapstructure:"db_password"`
	DBName     string `mapstructure:"db_name"`
}

var Cfg Config

func InitConfig() {
	viper.SetConfigName("config-dev")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Ошибка чтения конфигурации: %v", err)
	}

	if err := viper.Unmarshal(&Cfg); err != nil {
		log.Fatalf("Ошибка разбора конфигурации: %v", err)
	}

	log.Printf("Конфигурация загружена: %+v", Cfg) // Добавьте это для отладки
}
