package bot

import (
	"database/sql"
	"log"
	"os"
	"sakura_ai_bot/api"

	"github.com/bwmarrin/discordgo"
)

var sakuraIDList []api.SakuraID
var historyDB *sql.DB
var s *discordgo.Session

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

	setupCommand()
}

func setupCommand() {
	s.ApplicationCommandCreate(
		s.State.Application.ID,
		"",
		&discordgo.ApplicationCommand{
			Name:        "ask",
			Description: "Ask the AI",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "message",
					Description: "Message",
					Required:    true,
				},
				{
					Type: discordgo.ApplicationCommandOptionInteger,
					Name: "model",
					Description: "AI model to ask",
					Required: true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  api.GPT_OSS_120b.Name(),
							Value: api.GPT_OSS_120b,
						},
						{
							Name:  api.Qwen3_Coder_30B_A3B_Instruct.Name(),
							Value: api.Qwen3_Coder_30B_A3B_Instruct,
						},
						{
							Name:  api.Qwen3_Coder_480B_A35B_Instruct_FP8.Name(),
							Value: api.Qwen3_Coder_480B_A35B_Instruct_FP8,
						},
						{
							Name: api.LLM_JP_3_1_8x13b_instruct4.Name(), 
							Value: api.LLM_JP_3_1_8x13b_instruct4,
						},
						{
							Name:  api.Phi_4_mini_instruct_cpu.Name(),
							Value: api.Phi_4_mini_instruct_cpu,
						},
						{
							Name: api.Phi_4_multimodal_instruct.Name(),
							Value: api.Phi_4_multimodal_instruct,
						},
						{
							Name:  api.Qwen3_0_6B_cpu.Name(),
							Value: api.Qwen3_0_6B_cpu,
						},
						{
							Name:  api.Qwen3_VL_30B_A3B_Instruct.Name(),
							Value: api.Qwen3_VL_30B_A3B_Instruct,
						},
					},
				},
			},
		},	
	)
	
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.ApplicationCommandData().Name == "ask" {
			askCommand(s, i)	
		}
	})		
}

func askCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, o  := range options {
		optionMap[o.Name] = o
	}

	msg := ""
	if o, ok := optionMap["message"]; ok {
		msg = o.StringValue()
	}
	model := api.GPT_OSS_120b
	if o, ok := optionMap["model"]; ok {
		model = api.AIModel(o.IntValue())
	}
	if msg == "" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "Message is must to be not empty"},
		})
		return
	}	
}
