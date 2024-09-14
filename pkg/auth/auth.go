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
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

var jwtSecret = []byte(config.Cfg.JWTSecret)

type AuthServiceServer struct {
	authpb.UnimplementedAuthServiceServer
}

type Claims struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
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

	_, err = db.DBConn.Exec(ctx, query, req.Email, nil, fmt.Sprintf("%v", code))
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

func (s *AuthServiceServer) ValidateCode(ctx context.Context, req *authpb.ValidateCodeRequest) (*authpb.ValidateCodeResponse, error) {
	var code string
	var codeExpiration time.Time

	// Загружаем SQL-запрос
	query, err := sql.LoadSQLFile("auth/code-validate.sql")
	if err != nil {
		log.Println("Ошибка загрузки SQL-запроса: " + err.Error())
		return nil, err
	}

	// Извлекаем код и время его истечения
	err = db.DBConn.QueryRow(ctx, query, req.Email, req.Phone).Scan(&code, &codeExpiration)
	if err != nil {
		log.Printf("Ошибка запроса к базе данных: %v", err)
		return nil, errors.New("пользователь не найден")
	}

	// Проверяем, истёк ли срок действия кода
	if time.Now().After(codeExpiration) {
		log.Println("Время действия кода истекло")
		return nil, errors.New("время действия кода истекло")
	}

	// Сравниваем коды
	if fmt.Sprintf("%v", code) != fmt.Sprintf("%v", req.Code) {
		log.Println("неверный код")
		return nil, errors.New("неверный код")
	}

	// Создаём JWT-токен при успешной валидации кода
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Email: req.Email,
		Phone: req.Phone,
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
	return &authpb.ValidateCodeResponse{Token: tokenString}, nil
}

func (s *AuthServiceServer) LoginByPhoneNumber(ctx context.Context, req *authpb.LoginByPhoneNumberRequest) (*authpb.LoginByPhoneNumberResponse, error) {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username:   config.Cfg.TwilioAccountSid, // AccountSid as Username
		Password:   config.Cfg.TwilioAuthToken,  // AuthToken as Password
		AccountSid: config.Cfg.TwilioAccountSid, // Also provide AccountSid
	})

	code := generateCode(1000, 9999)

	params := &openapi.CreateMessageParams{}
	params.SetTo(req.Phone)
	params.SetFrom(config.Cfg.TwilioPhoneNumber)
	params.SetBody(fmt.Sprint(code))

	query, err := sql.LoadSQLFile("auth/code-generate.sql")
	if err != nil {
		log.Println("Ошибка загрузки SQL-запроса: " + err.Error())
		return &authpb.LoginByPhoneNumberResponse{Success: false}, err
	}

	_, err = db.DBConn.Exec(ctx, query, nil, req.Phone, fmt.Sprintf("%v", code))
	if err != nil {
		fmt.Printf("Не сохранить записи: %v", err)
		return &authpb.LoginByPhoneNumberResponse{Success: false}, err
	}

	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		fmt.Printf("Failed to send SMS: %v", err)
		return &authpb.LoginByPhoneNumberResponse{Success: false}, err
	}
	fmt.Printf("SMS sent successfully! SID: %s\n", *resp.Sid)
	return &authpb.LoginByPhoneNumberResponse{Success: true}, nil
}
