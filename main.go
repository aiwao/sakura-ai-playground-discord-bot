package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sakura_ai_bot/api"
	"sakura_ai_bot/bot"
	"strconv"

	"github.com/bwmarrin/discordgo"
	_ "github.com/lib/pq"
)

const baseURL = "https://playground.aipf.sakura.ad.jp/api"
const chatURL = baseURL+"/chat"

func main() {
	dbHost := os.Getenv("ID_DB_HOST")	
	dbPort := os.Getenv("ID_DB_PORT")	
	dbUser := os.Getenv("ID_DB_USER")	
	dbPass := os.Getenv("ID_DB_PASS")	
	dbName := os.Getenv("ID_DB_NAME")
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName)
	log.Println("Database: " + dsn)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	amountSakuraIDToUseEnv := os.Getenv("AMOUNT_SAKURA_ID_TO_USE")
	amountSakuraIDToUse, err := strconv.Atoi(amountSakuraIDToUseEnv)
	if err != nil {
		log.Println(err)
		amountSakuraIDToUse = 10
	}
	log.Printf("Use %d Sakura IDs\n", amountSakuraIDToUse)

	rows, err := db.Query(
		"SELECT * FROM accounts ORDER BY created_at DESC LIMIT ?",
		amountSakuraIDToUse,
	)
	if err != nil {
		log.Fatalln(err)
	}
	defer rows.Close()
	sakuraIDCount := 0
	sakuraIDList := []api.SakuraID{}
	for rows.Next() {
		var sakuraID api.SakuraID
		if err := rows.Scan(&sakuraID.Email, &sakuraID.Password, &sakuraID.CreatedAt); err != nil {
			log.Println(err)
			continue
		}
		sakuraIDList = append(sakuraIDList, sakuraID)
		sakuraIDCount++
		if sakuraIDCount >= amountSakuraIDToUse {
			break
		}
	}
	db.Close()
	if len(sakuraIDList) == 0 {
		log.Fatalln("No Sakura ID is found.")
	}
	log.Printf("Sakura ID count: %d\n", len(sakuraIDList))

	s, err := discordgo.New("Bot "+os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatalln(err)
	}

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v\n", s.State.User.Username, s.State.User.Discriminator)
	})

	bot.SetupCommand(s)	
}
