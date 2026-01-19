package bot

import (
	"errors"
	"log"
	"os"
	"sakura_ai_bot/sessionmanager"
	"sakura_ai_bot/utility"
	"time"

	"github.com/bwmarrin/discordgo"
)

func Setup() {
	sakuraIDList := utility.LoadSessionIDList()	
	sessionmanager.StartServer(sakuraIDList)

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

func replyBigString(message string, s *discordgo.Session, i *discordgo.InteractionCreate) {
	splitBy900 := utility.SplitByN(message, 900)
	for idx, spl := range splitBy900 {
		reply(spl, s, i)
		if idx != len(splitBy900)-1 {
			time.Sleep(1*time.Second)
		}
	}
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
