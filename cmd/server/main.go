package main

import (
	"BookingGo/pkg/db"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal("DB_URl в .env не задан")
	}
	db.InitDB(dbUrl)
}
