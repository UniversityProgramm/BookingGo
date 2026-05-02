package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Пул подключений к БД
var DB *pgxpool.Pool

func InitDB(dbUrl string) error {
	var err error
	DB, err = pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		return err
	}

	if err := DB.Ping(context.Background()); err != nil {
		return err
	}

	return nil
}
