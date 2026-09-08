package config

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
	env "github.com/sethvargo/go-envconfig"
)

type Config struct {
	Postgres PostgresConfig
	Service  ServiceConfig
}

type ServiceConfig struct {
	AdminToken string `env:"ADMIN_TOKEN"`
	HTTPHost   string `env:"SERVICE_HTTP_HOST, default=localhost"`
	HTTPPort   string `env:"SERVICE_HTTP_PORT, default=8080"`
}

type PostgresConfig struct {
	Name     string `env:"POSTGRES_DB, default=car_damps"`
	Host     string `env:"POSTGRES_HOST, default=localhost"`
	Port     int    `env:"POSTGRES_PORT, default=5432"`
	User     string `env:"POSTGRES_USER"`
	Password string `env:"POSTGRES_PASSWORD"`
	SSLMode  string `env:"SSL_MODE"`
}

func (c *Config) DatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Europe/Moscow",
		c.Postgres.Host,
		c.Postgres.User,
		c.Postgres.Password,
		c.Postgres.Name,
		c.Postgres.Port,
		c.Postgres.SSLMode,
	)
}

func (c *Config) ServiceHTTPConnString() string {
	return fmt.Sprintf("%s:%s", c.Service.HTTPHost, c.Service.HTTPPort)
}

var conf *Config

func Load() (*Config, error) {
	if conf != nil {
		return conf, nil
	}

	conf = &Config{}

	_, filename, _, ok := runtime.Caller(0)

	if !ok {
		return nil, errors.New("caller error")
	}

	dir := filepath.Dir(filename)
	envPath := filepath.Join(dir, "..", "..", ".env")

	envFile, err := filepath.Abs(envPath)
	if err != nil {
		return nil, fmt.Errorf("err to get abs path - %s", envPath)
	}

	if err := godotenv.Load(envFile); err != nil {
		return nil, errors.New("err loading .env file")
	}

	ctx := context.Background()
	if err := env.Process(ctx, conf); err != nil {
		log.Fatalf("%s", err)
	}

	return conf, nil
}

func AppConfig() *Config {
	conf, err := Load()
	if err != nil {
		slog.Error("config load error", "err", err)
		os.Exit(1)
	}

	return conf
}
