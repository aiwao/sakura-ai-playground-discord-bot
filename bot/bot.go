package bot

import (
	"database/sql"
	"log"
	"os"
	"sakura_ai_bot/api"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"

	_ "github.com/mattn/go-sqlite3"
)

var sakuraIDList []api.SakuraID
var historyDB *sql.DB
var (
	mu sync.Mutex
	sessionList = []*api.SakuraSession{}
)

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

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		handleCommand(s, i)
	})

	if err := s.Open(); err != nil {
		log.Fatalln(err)
	}

	go func() {
		for _, id := range sakuraIDList {
			session, err := id.NewSakuraSession()
			if err != nil {
				log.Println(err)
				continue
			}
			mu.Lock()
			sessionList = append(sessionList, session)
			mu.Unlock()
			time.Sleep(30*time.Second)
		}
	}()

	registerCommand(s, AskCommand())
}

func reply(message string, s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Content: message,
	})
}
