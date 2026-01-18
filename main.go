package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sakura_ai_bot/bot"
	"sakura_ai_bot/utility"
	"strconv"

	_ "github.com/lib/pq"
)

func main() {
	loadSD, err := strconv.Atoi(os.Getenv("LOAD_SESSION_DELAY"))
	if err == nil {
		utility.LoadSessionDelay = loadSD
	}
	checkMD, err := strconv.Atoi(os.Getenv("CHECK_MAIL_DELAY"))
	if err == nil {
		utility.CheckMailDelay = checkMD
	}

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
