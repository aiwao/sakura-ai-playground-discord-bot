package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sakura_ai_bot/api"
	"sakura_ai_bot/bot"

	_ "github.com/lib/pq"
)

func main() {
	dbHost := os.Getenv("DB_HOST")	
	dbPort := os.Getenv("DB_PORT")	
	dbUser := os.Getenv("DB_USER")	
	dbPass := os.Getenv("DB_PASS")	
	dbName := os.Getenv("DB_NAME")
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName)
	log.Println("Database: " + dsn)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()	

	bot.Setup(db)
}
