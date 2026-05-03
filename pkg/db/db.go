package db

import (
	"errors"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Пул подключений к БД
var DB *gorm.DB

func InitDB() error {
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		return errors.New("env DB_URL is not set")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dbUrl), &gorm.Config{})
	if err != nil {
		return err
	}
	return nil
}
