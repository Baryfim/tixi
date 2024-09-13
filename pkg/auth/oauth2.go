package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// var (
// 	clientID     = config.Cfg.GoogleClientID
// 	clientSecret = config.Cfg.GoogleClientSecret
// )

// Google OAuth2 конфигурация
var oauthGoogleConfig = &oauth2.Config{
	ClientID:     "697503104115-ui1i318hgo5h9d439tanq51s6gtc3pd8.apps.googleusercontent.com",
	ClientSecret: "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",
	RedirectURL:  "http://localhost:8080/callback",
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
}

// Обработчик для перенаправления пользователя на страницу авторизации Google
func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	url := oauthGoogleConfig.AuthCodeURL("randomstate", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Обработчик для получения OAuth-кода и обмена его на токен
func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Получить код авторизации от Google
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code", http.StatusBadRequest)
		return
	}

	// Обменять код на токен
	token, err := oauthGoogleConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Token exchange failed", http.StatusInternalServerError)
		return
	}

	// Получить пользовательскую информацию с помощью полученного токена
	client := oauthGoogleConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Email: fmt.Sprint(userInfo["email"]),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	tokenJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := tokenJWT.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Email not found", http.StatusBadRequest)
	}

	response := map[string]string{
		"token": tokenString,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Инициализация обработчиков аутентификации
func InitAuthHandles(mux *http.ServeMux) {
	mux.HandleFunc("/login", GoogleLoginHandler)       // Обработчик для входа через Google
	mux.HandleFunc("/callback", GoogleCallbackHandler) // Обработчик для callback от Google
}
