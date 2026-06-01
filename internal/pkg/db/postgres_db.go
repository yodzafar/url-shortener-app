package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/yodzafar/url-shortener-app/internal/config"
)

func NewPostgresDb(cfg *config.Config) (*sqlx.DB, func(), error) {
	db, err := sqlx.Connect("postgres", cfg.DB.URL)

	if err != nil {
		return nil, nil, fmt.Errorf("connect db: %v", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("ping db: %w", err)
	}

	db.SetMaxOpenConns(int(cfg.DB.MaxConns))
	db.SetMaxIdleConns(int(cfg.DB.MinConns))
	db.SetConnMaxIdleTime(cfg.DB.ConnTimeout)

	return db, func() { db.Close() }, nil

}
