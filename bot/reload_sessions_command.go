package bot

import (
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

func ReloadSessionsCommand() *Command {
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

				if id != os.Getenv("OWNER_ID") {
					reply("You can't access to this command", s, i)
					return
				}

				if cancelLoadSessions != nil {
					cancelLoadSessions()
				}
				loadSessions()

				reply("Reloaded", s, i)
			}()
		},
		ApplicationCommand: discordgo.ApplicationCommand{

		},
	}
}
