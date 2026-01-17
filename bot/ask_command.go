package bot

import (
	"log"
	"math/rand/v2"
	"sakura_ai_bot/api"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

func AskCommand() *Command {
	return &Command{
		Action: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			sessionList := <-sessionChan
			if sessionList == nil || len(sessionList) == 0 {
				reply("Application is starting now", s, i)
				return
			}

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
				reply("Message is must to be not empty", s, i)	
				return
			}

			rows, err := historyDB.Query(
				"SELECT (content, id, role) FROM histories WHERE user_id = $1",
				i.Member.User.ID,
			)
			if err != nil {
				log.Println(err)
				reply("Internal server error", s, i)
				return
			}
			defer rows.Close()

			messages := []api.Message{}
			for rows.Next() {
				var message api.Message
				if err := rows.Scan(&message.Content, &message.ID, &message.Role); err != nil {
					log.Println(err)
					reply("Internal server error", s, i)
					return
				}
				messages = append(messages, message)
			}
			minID := 1000000000
			maxID := 9999999999
			msgID := rand.IntN(maxID-minID+1)+minID
			messages = append(messages, api.Message{ID: strconv.Itoa(msgID), Content: msg, Role: "user"})

			for idx := range min(20, len(sessionList)) {
				session := sessionList[idx]
				c, err := session.Chat(api.ChatPayload{Messages: messages, Model: model.Name()})
				if err != nil {
					log.Println(err)
					continue
				}
				reply(c.Content, s, i)
				break
			}	
		},
		ApplicationCommand: discordgo.ApplicationCommand{
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
	}
}
