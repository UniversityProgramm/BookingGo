package db

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Пул подключений к БД
var DB *pgxpool.Pool

func InitDB(dbUrl string) {
	var err error
	DB, err = pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		log.Fatal("Не удалось подключиться к БД:", err)
	}

	if err := DB.Ping(context.Background()); err != nil {
		log.Fatal("Не удалось пропинговать БД:", err)
	}
	log.SetOutput(os.Stdout)
	log.Println("Подключились к базе PostgreSQL")
}
