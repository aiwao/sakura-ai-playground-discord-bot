package bot

import (
	"log"
	"os"
	"sakura_ai_bot/sessionmanager"
	"sakura_ai_bot/utility"

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

				newIDList := utility.LoadSessionIDList()
				_, err = sessionmanager.Request(sessionmanager.RequestBody{
					Method: sessionmanager.Reload,
					IDListReload: newIDList,
				})
				if err != nil {
					reply("Internal server error", s, i)
					return
				}

				reply("Reloaded", s, i)
			}()
		},
		ApplicationCommand: discordgo.ApplicationCommand{
			Name: "reload_sessions",
			Description: "[Owner only] Reload sessions",
		},
	}
}
