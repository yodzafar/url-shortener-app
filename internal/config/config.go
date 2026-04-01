package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App     AppConfig
	DB      DBConfig
	Server  ServerConfig
	Session SessionConfig
}

type AppConfig struct {
	Env   string
	Debug bool
}

type DBConfig struct {
	URL           string
	MaxConns      int32
	MinConns      int32
	ConnTimeout   time.Duration
	MigrationsDir string
}

type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type SessionConfig struct {
	TTL        time.Duration
	CookieName string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	cfg := &Config{
		App: AppConfig{
			Env:   getEnv("APP_ENV", "development"),
			Debug: getBoolEnv("APP_DEBUG", false),
		},
		DB: DBConfig{
			URL:           mustEnv("DB_URL"),
			MaxConns:      int32(getIntEnv("DB_MAX_CONNS", 20)),
			MinConns:      int32(getIntEnv("DB_MIN_CONNS", 2)),
			ConnTimeout:   getDurationEnv("DB_CONN_TIMEOUT", 5*time.Second),
			MigrationsDir: getEnv("DB_MIGRATIONS_DIR", "./migrations"),
		},
		Server: ServerConfig{
			Port:            getEnv("SERVER_PORT", "8080"),
			ReadTimeout:     getDurationEnv("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getDurationEnv("SERVER_WRITE_TIMEOUT", 30*time.Second),
			ShutdownTimeout: getDurationEnv("SERVER_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		Session: SessionConfig{
			TTL:        getDurationEnv("SESSION_TTL", 7*24*time.Hour),
			CookieName: getEnv("SESSION_COOKIE_NAME", "session_id"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	return cfg, nil
}

func Mustload() *Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	return cfg
}

func (c *Config) validate() error {
	var errs []string

	if c.DB.URL == "" {
		errs = append(errs, "DB_URL is required")
	}

	if c.DB.MaxConns <= c.DB.MinConns {
		errs = append(errs, "DB_MAX_CONNS must be greater than or equal to DB_MIN_CONNS")
	}

	if c.Session.TTL <= time.Minute {
		errs = append(errs, "SESSION_TTL must be greater than 1 minute")
	}

	valid := map[string]bool{"development": true, "production": true, "test": true}

	if valid[c.App.Env] == false {
		errs = append(errs, "APP_ENV must be one of: development, production, test")
	}

	if len(errs) > 0 {
		return errors.New("config validation errors: " + fmt.Sprintf("%v", errs))
	}

	return nil
}

func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}
