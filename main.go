package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sakura_ai_bot/api"
	"sakura_ai_bot/bot"
	"syscall"

	_ "github.com/lib/pq"
)

func main() {
	sakuraIDJSONFile := os.Getenv("SAKURA_ID_JSON")
	if sakuraIDJSONFile == "" {
		log.Fatalln("SAKURA_ID_JSON is empty. please define a JSON file path of Sakura-ID list.")
	}
	
	b, err := os.ReadFile(sakuraIDJSONFile)
	if err != nil {
		log.Fatalln(err)
	}
	sakuraIDList := []api.SakuraID{}
	if err := json.Unmarshal(b, &sakuraIDList); err != nil {
		log.Fatalln(err)
	}
	idCnt := len(sakuraIDList)
	if idCnt == 0 {
		log.Fatalln("Sakura-ID list is empty")
	}
	log.Printf("Sakura-ID count: %d\n", idCnt)

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

	bot.Setup(sakuraIDList, db)
}
