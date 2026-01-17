package bot

import (
	"database/sql"
	"log"
	"os"
	"sakura_ai_bot/api"
	"time"

	"github.com/bwmarrin/discordgo"

	_ "github.com/mattn/go-sqlite3"
)

var sakuraIDList []api.SakuraID
var historyDB *sql.DB
var sessionChan chan []*api.SakuraSession

func Setup(idList []api.SakuraID, db *sql.DB) {
	sakuraIDList = idList
	historyDB = db

	s, err := discordgo.New("Bot "+os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatalln(err)
	}	

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v\n", s.State.User.Username, s.State.User.Discriminator)
	})

	if err := s.Open(); err != nil {
		log.Fatalln(err)
	}

	sessionChan = make(chan []*api.SakuraSession)
	go func() {
		sessionList := []*api.SakuraSession{}
		for _, id := range sakuraIDList {
			session, err := id.NewSakuraSession()
			if err != nil {
				log.Println(err)
				continue
			}
			sessionList = append(sessionList, session)
			sessionChan <- sessionList
			time.Sleep(30*time.Second)
		}
	}()

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		handleCommand(s, i)
	})

	registerCommand(s, AskCommand())
}

func reply(message string, s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: message},
	})
}
