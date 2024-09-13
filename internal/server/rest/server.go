package rest

import (
	"net/http"
	"time"

	"github.com/tixiby/internal/config"
	"github.com/tixiby/pkg/auth"
)

func RunRESTServer() error {
	// Настройка HTTP сервера
	mux := http.NewServeMux()

	// Init Handles
	auth.InitAuthHandles(mux)

	// Server Setup
	srv := &http.Server{
		Addr:         config.Cfg.RESTPort, // порт REST сервера
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return srv.ListenAndServe()
}
