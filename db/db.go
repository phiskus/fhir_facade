package db

import (
	"database/sql"
	"fmt"
	"time"
	"fhir_facade/config"
	_ "github.com/lib/pq"

)

func New(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
    )

	db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, fmt.Errorf("sql.Open: %w", err)
    }
	
	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
    db.SetMaxIdleConns(cfg.DBMaxIdleConns)
    db.SetConnMaxLifetime(5 * time.Minute)

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("db.Ping: %w", err)
    }

    return db, nil
}