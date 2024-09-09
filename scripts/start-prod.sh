#!/bin/bash

echo "Запуск сервера в продакшене..."

# Запуск сервера с использованием переменных окружения
export GO_ENV=production
go run ./cmd/server
