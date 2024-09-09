package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/tixiby/api/proto/authpb"
	"github.com/tixiby/internal/config"
	"github.com/tixiby/internal/db"
	"golang.org/x/crypto/bcrypt"
)

var Cfg config.Config
var jwtSecret = []byte(Cfg.JWTSecret)

type AuthServiceServer struct {
	authpb.UnimplementedAuthServiceServer
}

// Claims структура для JWT
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// Функция для хэширования пароля
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// Проверка пароля с хэшем
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Регистрация нового пользователя
func (s *AuthServiceServer) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	_, err = db.DBConn.Exec(ctx, "INSERT INTO users (username, password_hash) VALUES ($1, $2)", req.Username, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("не удалось сохранить пользователя: %v", err)
	}

	return &authpb.RegisterResponse{Success: true}, nil
}

func (s *AuthServiceServer) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	var passwordHash string

	err := db.DBConn.QueryRow(ctx, "SELECT password_hash FROM users WHERE username = $1", req.Username).Scan(&passwordHash)
	if err != nil {
		return nil, errors.New("пользователь не найден")
	}

	if !CheckPasswordHash(req.Password, passwordHash) {
		return nil, errors.New("неверный пароль")
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: req.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return nil, err
	}

	return &authpb.LoginResponse{Token: tokenString}, nil
}

func (s *AuthServiceServer) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	token, err := jwt.ParseWithClaims(req.Token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return &authpb.ValidateTokenResponse{IsValid: false}, nil
	}

	return &authpb.ValidateTokenResponse{IsValid: true}, nil
}
