package bot

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type Command struct {
	Action func(s *discordgo.Session, i *discordgo.InteractionCreate)
	*discordgo.ApplicationCommand
}

var commandRegistry = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){}

func registerCommand(s *discordgo.Session, c *Command) {
	_, err := s.ApplicationCommandCreate(
		s.State.User.ID,
		"",
		c.ApplicationCommand,
	)
	if err != nil {
		log.Println(err)
		return
	}
	commandRegistry[c.Name] = c.Action
}

func handleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if cmd, ok := commandRegistry[i.ApplicationCommandData().Name]; ok {
		cmd(s, i)
	}
}
