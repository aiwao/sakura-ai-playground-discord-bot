package bot

import (
	"log"
	"math/rand/v2"
	"sakura_ai_bot/api"
	"sakura_ai_bot/utility"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

func AskCommand() *Command {
	modelChoices := []*discordgo.ApplicationCommandOptionChoice{}
	for _, m := range api.AIModelList {
		modelChoices = append(modelChoices, &discordgo.ApplicationCommandOptionChoice{
			Name: m,
			Value: m,
		})
	}

	return &Command{
		Action: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			ephemeral, err := getOptionBool("ephemeral", mapOption(i))
			if err != nil {
				reply("Ephemeral is required", s, i)
				return
			}
			if ephemeral {
				thinkingEphemeral(s, i)
			} else {
				thinking(s, i)
			}
			go func() {
				mu.Lock()
				sessionListCopy := append([]*api.SakuraSession(nil), sessionList...)
				mu.Unlock()
				if len(sessionListCopy) == 0 {
					reply("Application is preparing now", s, i)
					return
				}

				optionMap := mapOption(i)
				msg, err := getOptionString("message", optionMap)
				if err != nil || msg == "" {
					reply("Message is required", s, i)
					return
				}
			
				model, err := getOptionString("model", optionMap)
				if err != nil || model == "" {
					reply("Model is required", s, i)
					return
				}

				id, err := getUserID(i)
				if err != nil {
					log.Println(err)
					reply("Internal server error", s, i)
					return
				}

				rows, err := botDB.Query(
					"SELECT content, id, role FROM histories WHERE user_id = $1 ORDER BY message_order ASC",
					id,
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
				userMSG := api.Message{ID: strconv.Itoa(msgID), Content: msg, Role: "user"}
				messages = append(messages, userMSG)

				for _, session := range sessionListCopy {
					if session.InvalidRequestCount >= utility.MaxInvalid {
						chAddSessions <- 1	
						continue
					}

					c, err := session.Chat(api.ChatPayload{Messages: messages, Model: model})
					if err != nil {
						log.Println(err)
						continue
					}
	
					replyBigString(c.Content, s, i)

					dbQuery := "INSERT INTO histories(user_id, content, id, role) VALUES ($1, $2, $3, $4)"
					_, err = botDB.Exec(
						dbQuery,
						id,
						userMSG.Content,
						userMSG.ID,
						userMSG.Role,
					)
					if err != nil {
						log.Println(err)
						return
					}
					_, err = botDB.Exec(
						dbQuery,
						id,
						c.Content,
						c.ID,
						c.Role,
					)
					if err != nil {
						log.Println(err)
					}
					return
				}

				reply("Internal server error", s, i)
			}()
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
					Type: discordgo.ApplicationCommandOptionString,
					Name: "model",
					Description: "AI model to ask",
					Required: true,
					Choices: modelChoices,
				},
				{
					Name: "ephemeral",
					Description: "Only for you",
					Type: discordgo.ApplicationCommandOptionBoolean,
					Required: true,
				},
			},
		},
	}
}
