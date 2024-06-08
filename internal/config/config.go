package config

import (
	"fmt"
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env             string        `env:"ENV" env-required:"true"`
	Postgres        Postgres      `env-required:"true"`
	HTTPServer      HTTPServer    `env-required:"true"`
	Secret          string        `env:"SECRET" env-required:"true" env:"SECRET"`
	AccessTokenTTL  time.Duration `env:"ACCESS_TOKEN_TTL" env-required:"true"`
	RefreshTokenTTL time.Duration `env:"REFRESH_TOKEN_TTL" env-required:"true"`
}

type Postgres struct {
	Host         string `env:"POSTGRES_HOST"`
	Port         int    `env:"POSTGRES_PORT"`
	User         string `env:"POSTGRES_USER"`
	Password     string `env:"POSTGRES_PASSWORD" env-required:"true"`
	DatabaseName string `env:"POSTGRES_DB"`
	DSNTemplate  string
}

type HTTPServer struct {
	Address     string        `env:"HTTP_ADDRESS" env-default:"localhost:8080"`
	Timeout     time.Duration `env:"TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `env:"IDLE_TIMEOUT" env-default:"60s"`
}

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal("can't load .env")
	}

	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	cfg.Postgres.DSNTemplate = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.DatabaseName)

	return &cfg
}
