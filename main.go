package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sakura_ai_bot/bot"
	"sakura_ai_bot/environment"

	"github.com/aiwao/envar"
	_ "github.com/lib/pq"
)

func main() {
	envar.GetIntv("LOAD_SESSION_DELAY", &environment.LoadSessionDelay)
	envar.GetIntv("CHECK_MAIL_DELAY", &environment.CheckMailDelay)
	envar.GetIntv("MAX_SAKURA_SESSIONS", &environment.MaxSessions)
	envar.GetIntv("MAX_INVALID_REQUEST_COUNT", &environment.MaxInvalid)
	
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
	environment.DB = db
	
	bot.Setup()
}
