package bot

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func ClearHistoryCommand() *Command {
	return &Command{
		Action: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			thinkingEphemeral(s, i)
			go func() {
				id, err := getUserID(i)
				if err != nil {
					log.Println(err)
					reply("Internal server error", s, i)
					return
				}
				_, err = botDB.Exec("DELETE FROM histories WHERE user_id = $1", id)
				if err != nil {
					log.Println(err)
					reply("Internal server error", s, i)
					return
				}
				reply("Cleared", s, i)
			}()
		},
		ApplicationCommand: discordgo.ApplicationCommand{
			Name: "clear_history",
			Description: "Clear your chat history",
		},
	}
}
