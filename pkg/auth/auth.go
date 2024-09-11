package auth

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"time"

	crand "crypto/rand"
	"math/big"

	"github.com/golang-jwt/jwt"
	"github.com/tixiby/api/proto/authpb"
	"github.com/tixiby/internal/config"
	"github.com/tixiby/internal/db"
	"github.com/tixiby/pkg/sql"
)

var jwtSecret = []byte(config.Cfg.JWTSecret)

type AuthServiceServer struct {
	authpb.UnimplementedAuthServiceServer
}

type Claims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

func generateCode(min, max int) int {
	n, err := crand.Int(crand.Reader, big.NewInt(int64(max-min)))
	if err != nil {
		log.Println("Ошибка генерации кода: " + err.Error())
	}
	return int(n.Int64()) + min
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

func (s *AuthServiceServer) LoginByEmail(ctx context.Context, req *authpb.LoginByEmailRequest) (*authpb.LoginByEmailResponse, error) {
	to := []string{
		req.Email,
	}

	auth := smtp.PlainAuth("", config.Cfg.MailFrom, config.Cfg.MailPassword, config.Cfg.SMTPHost)
	temp, _ := template.ParseFiles("./templates/email.html")

	var body bytes.Buffer

	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body.Write([]byte(fmt.Sprintf("Subject: This is a test subject \n%s\n\n", mimeHeaders)))

	code := generateCode(1000, 9999)
	temp.Execute(&body, struct {
		Name    string
		Message string
	}{
		Name:    "TIXI Company",
		Message: fmt.Sprintf("%v", code),
	})

	query, err := sql.LoadSQLFile("auth/code-generate.sql")
	if err != nil {
		log.Println("Ошибка загрузки SQL-запроса: " + err.Error())
		return &authpb.LoginByEmailResponse{Success: false}, err
	}

	_, err = db.DBConn.Exec(ctx, query, req.Email, fmt.Sprintf("%v", code))
	if err != nil {
		fmt.Printf("Не сохранить записи: %v", err)
		return &authpb.LoginByEmailResponse{Success: false}, err
	}

	if err = smtp.SendMail(fmt.Sprintf("%s:587", config.Cfg.SMTPHost), auth, config.Cfg.MailFrom, to, body.Bytes()); err != nil {
		fmt.Println("Error sending email:", err)
		return &authpb.LoginByEmailResponse{Success: false}, err
	}

	return &authpb.LoginByEmailResponse{Success: true}, nil
}

func (s *AuthServiceServer) ValidateCodeEmail(ctx context.Context, req *authpb.ValidateCodeEmailRequest) (*authpb.ValidateCodeEmailResponse, error) {
	var code string

	query, err := sql.LoadSQLFile("auth/code-validate.sql")
	if err != nil {
		log.Println("Ошибка загрузки SQL-запроса: " + err.Error())
		return nil, err
	}

	err = db.DBConn.QueryRow(ctx, query, req.Email).Scan(&code)
	if err != nil {
		log.Printf("Ошибка запроса к базе данных: %v", err)
		return nil, errors.New("пользователь не найден")
	}

	if fmt.Sprintf("%v", code) != fmt.Sprintf("%v", req.Code) {
		log.Println("неверный код")
		return nil, errors.New("неверный код")
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Email: req.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return nil, err
	}

	// Возвращаем токен
	return &authpb.ValidateCodeEmailResponse{Token: tokenString}, nil
}
