package bot

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"sakura_ai_bot/api"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
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
	defer s.Close()

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
			time.Sleep(1*time.Second)
		}
	}()

	registerCommand(s, AskCommand())
	registerCommand(s, ClearHistoryCommand())
	registerCommand(s, ShowHistoryCommand())

	select {}
}

func getUserID(i *discordgo.InteractionCreate) (string, error) {
	if i.Member != nil {
		return i.Member.User.ID, nil
	} else if i.User != nil {
		return i.User.ID, nil
	}
	return "", errors.New("failed to get user ID")
}

func thinkingFlag(s *discordgo.Session, i *discordgo.InteractionCreate, f int) {
	data := &discordgo.InteractionResponseData{}
	if f != -1 {
		data.Flags = discordgo.MessageFlags(f)
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: data,
	})
}

func thinking(s *discordgo.Session, i *discordgo.InteractionCreate) {
	thinkingFlag(s, i, -1)	
}

func thinkingEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate) {
	thinkingFlag(s, i, int(discordgo.MessageFlagsEphemeral))
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
