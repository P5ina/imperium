package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Connect(databaseURL string) error {
	var err error
	Pool, err = pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	return Pool.Ping(context.Background())
}

func Migrate() error {
	sql, err := os.ReadFile("db/migrations/001_init.sql")
	if err != nil {
		return fmt.Errorf("reading migration: %w", err)
	}
	_, err = Pool.Exec(context.Background(), string(sql))
	if err != nil {
		return fmt.Errorf("running migration: %w", err)
	}
	return nil
}

func Close() {
	if Pool != nil {
		Pool.Close()
	}
}
