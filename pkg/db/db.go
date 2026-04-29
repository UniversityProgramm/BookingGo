package db

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func InitDB(dbUrl string) {
	DB, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		log.Fatal("Не удалось подключиться к БД:", err)
	}

	if err := DB.Ping(context.Background()); err != nil {
		log.Fatal("Не удалось пропинговать БД:", err)
	}
	log.SetOutput(os.Stdout)
	log.Println("Подключились к базе PostgreSQL")
}
