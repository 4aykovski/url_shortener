package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/4aykovski/learning/golang/rest/internal/config"
	"github.com/4aykovski/learning/golang/rest/internal/http-server/v1"
	"github.com/4aykovski/learning/golang/rest/internal/lib/hasher"
	"github.com/4aykovski/learning/golang/rest/internal/lib/logger/slogHelper"
	tokenManager "github.com/4aykovski/learning/golang/rest/internal/lib/token-manager"
	"github.com/4aykovski/learning/golang/rest/internal/repository/Postgres"
	"github.com/4aykovski/learning/golang/rest/internal/services"
	"github.com/natefinch/lumberjack"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// TODO: разобраться с prettyslog
	// init config: cleanenv
	cfg := config.MustLoad()

	// init logger: slog ? grafana ? kibana ? grep
	log := setupLogger(cfg.Env)
	log.Info("starting url-shortener", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	// init db: Postgres
	pq, err := Postgres.New(cfg.Postgres)
	if err != nil {
		log.Error("failed to init Postgres database", slogHelper.Err(err))
		os.Exit(1)
	}

	// init repositories
	urlRepo := Postgres.NewUrlRepository(pq)
	userRepo := Postgres.NewUserRepository(pq)
	refreshRepo := Postgres.NewRefreshSessionRepository(pq)

	// init additional stuff
	h := hasher.NewBcryptHasher()
	tM := tokenManager.New(cfg.Secret)

	// init services
	refreshService := services.NewRefreshSessionService(refreshRepo, tM, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	userService := services.NewUserService(userRepo, refreshService, h, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)

	// init router: chi, "chi render"
	mux := v1.NewMux(log, urlRepo, userService, tM)

	// run server
	log.Info("starting server", slog.String("address", cfg.HTTPServer.Address))

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      mux,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")
}

func setupLogger(env string) *slog.Logger {

	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		lumber := &lumberjack.Logger{
			Filename:   "logs/app.log",
			MaxSize:    10,
			MaxBackups: 3,
			MaxAge:     7,
		}
		log = slog.New(slog.NewJSONHandler(lumber, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		lumber := &lumberjack.Logger{
			Filename:   "logs/app.log",
			MaxSize:    10,
			MaxBackups: 3,
			MaxAge:     7,
		}
		log = slog.New(slog.NewJSONHandler(lumber, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
