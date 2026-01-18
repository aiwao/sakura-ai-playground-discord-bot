package bot

import (
	"log"
	"sakura_ai_bot/utility"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func ShowHistoryCommand() *Command {
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
				id, err := getUserID(i)
				if err != nil {
					log.Println(err)
					reply("Internal server error", s, i)
					return
				}

				rows, err := botDB.Query(
					"SELECT content, role FROM histories WHERE user_id = $1 ORDER BY message_order ASC",
					id,
				)
				if err != nil {
					log.Println(err)
					reply("Internal server error", s, i)
					return
				}
				defer rows.Close()

				messages := []string{}
				for rows.Next() {
					var history struct {
						Content string `json:"content"`
						Role string `json:"role"`
					}
					if err := rows.Scan(&history.Content, &history.Role); err != nil {
						log.Println(err)
						reply("Internal server error", s, i)
						return
					}

					contentTrim := history.Content
					runes := []rune(history.Content)
					if len(runes) > 10 {
						contentTrim = string(runes[:10])+"..."
					}
					messages = append(messages, history.Role+": "+contentTrim)
				}
				
				replyMSG := "History is empty"
				if len(messages) > 0 {
					replyMSG = strings.Join(messages, "\n")
				}
				
				for _, spl := range utility.SplitByN(replyMSG, 900) {
					reply(spl, s, i)
				}	
			}()	
		},
		ApplicationCommand: discordgo.ApplicationCommand{
			Name: "show_history",
			Description: "Show your chat history",
			Options: []*discordgo.ApplicationCommandOption{
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
