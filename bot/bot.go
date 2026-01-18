package bot

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"sakura_ai_bot/api"
	"sakura_ai_bot/utility"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

var sakuraIDList = []api.SakuraID{}
var botDB *sql.DB
var (
	mu sync.Mutex
	sessionList = []*api.SakuraSession{}
)
var cancelLoadSessions context.CancelFunc

func Setup(db *sql.DB) {
	botDB = db	

	loadSessions()

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
	defer s.Close()	

	registerCommand(s, AskCommand())
	registerCommand(s, ClearHistoryCommand())
	registerCommand(s, ShowHistoryCommand())
	registerCommand(s, ReloadSessionsCommand())

	select {}
}

func loadSessions() {
	sakuraIDList = []api.SakuraID{}
	sessionList = []*api.SakuraSession{}

	rows, err := botDB.Query("SELECT * FROM accounts")
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var idScan api.SakuraID
		if err := rows.Scan(&idScan.Email, &idScan.Password, &idScan.CreatedAt, &idScan.InstaddrID, &idScan.InstaddrPassword); err != nil {
			log.Println(err)
			continue
		}
		sakuraIDList = append(sakuraIDList, idScan)
	}
	log.Printf("Sakura ID count: %d\n", len(sakuraIDList))

	ctx, cancel := context.WithCancel(context.Background())
	cancelLoadSessions = cancel
	go func(ctx context.Context, cancel context.CancelFunc) {
		defer cancel()
		for _, id := range sakuraIDList {
			select {
			case <-ctx.Done():
				return
			default:
				session, err := id.NewSakuraSession()
				if err != nil {
					log.Println(err)
					continue
				}
				mu.Lock()
				sessionList = append(sessionList, session)
				mu.Unlock()
				time.Sleep(time.Duration(utility.LoadSessionDelay) * time.Millisecond)
			}
		}
	}(ctx, cancel)
}

func getUserID(i *discordgo.InteractionCreate) (string, error) {
	if i.Member != nil {
		return i.Member.User.ID, nil
	} else if i.User != nil {
		return i.User.ID, nil
	}
	return "", errors.New("failed to get user ID")
}

func thinkingFlag(s *discordgo.Session, i *discordgo.InteractionCreate, f discordgo.MessageFlags) {
	data := &discordgo.InteractionResponseData{}
	if f != discordgo.MessageFlags(-1) {
		data.Flags = discordgo.MessageFlags(f)
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: data,
	})
}

func thinking(s *discordgo.Session, i *discordgo.InteractionCreate) {
	thinkingFlag(s, i, discordgo.MessageFlags(-1))
}

func thinkingEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate) {
	thinkingFlag(s, i, discordgo.MessageFlagsEphemeral)
}

func reply(message string, s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Content: message,
	})
}

type OptionMap map[string]*discordgo.ApplicationCommandInteractionDataOption

func mapOption(i *discordgo.InteractionCreate) OptionMap {
	options := i.ApplicationCommandData().Options
	optionMap := make(OptionMap, len(options))
	for _, o  := range options {
		optionMap[o.Name] = o
	}
	return optionMap
}

func getOptionString(key string, optionMap OptionMap) (string, error) {
	if value, ok := optionMap[key]; ok {
		return value.StringValue(), nil
	}
	return "", errors.New("failed to get option")
}

func getOptionBool(key string, optionMap OptionMap) (bool, error) {
	if value, ok := optionMap[key]; ok {
		return value.BoolValue(), nil
	}
	return false, errors.New("failed to get option")
}

func getOptionInt(key string, optionMap OptionMap) (int64, error) {
	if value, ok := optionMap[key]; ok {
		return value.IntValue(), nil
	}
	return -1, errors.New("failed to get option")
}
