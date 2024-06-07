package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/natefinch/lumberjack"
	"github.com/rs/cors"

	v1 "github.com/4aykovski/url_shortener/internal/adapters/http-server/v1"
	"github.com/4aykovski/url_shortener/internal/adapters/repository/postgres"
	"github.com/4aykovski/url_shortener/internal/config"
	"github.com/4aykovski/url_shortener/internal/services"
	"github.com/4aykovski/url_shortener/pkg/hasher"
	"github.com/4aykovski/url_shortener/pkg/logger/slogHelper"
	"github.com/4aykovski/url_shortener/pkg/manager/token"
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
	log.Debug("Postgres configuration", slog.String("dbname", cfg.Postgres.DatabaseName), slog.String("user", cfg.Postgres.User), slog.String("host", cfg.Postgres.Host), slog.Int("port", cfg.Postgres.Port))

	// init db: Postgres
	pq, err := postgres.New(cfg.Postgres)
	if err != nil {
		log.Error("failed to init Postgres database", slogHelper.Err(err))
		os.Exit(1)
	}

	// init repositories
	urlRepo := postgres.NewUrlRepository(pq)
	userRepo := postgres.NewUserRepository(pq)
	refreshRepo := postgres.NewRefreshSessionRepository(pq)

	// init additional stuff
	h := hasher.NewBcryptHasher()
	tM := token.NewManager(cfg.Secret)

	// init services
	urlService := services.NewUrlService(urlRepo)
	refreshService := services.NewRefreshSessionService(refreshRepo, tM, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	userService := services.NewAuthService(userRepo, refreshService, h, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)

	// init router: chi, "chi render"
	mux := v1.NewMux(log, urlService, userService, tM)

	c := cors.New(cors.Options{
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodOptions,
		},
		AllowedOrigins: []string{
			"http://localhost:3000"},
		AllowCredentials: true,
		AllowedHeaders: []string{
			"Authorization",
			"Content-Type",
		},
		OptionsPassthrough: true,
		ExposedHeaders:     []string{},
		Debug:              true,
	})

	handler := c.Handler(mux)

	// run server
	log.Info("starting server", slog.String("address", cfg.HTTPServer.Address))

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      handler,
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
