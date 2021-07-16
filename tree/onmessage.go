package tree

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/eviedelta/openjishia/wlog"
)

// the thing that does the on message stuff
func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	// debug purposes only, logs all messages to the terminal
	// ### *** DO NOT ENABLE IN PRODUCTION ***
	if Conf.Bot.Debugm {
		log.Printf("New message by %v (%v) in %v : %v\n> %v\n", m.Author.String(), m.Author.ID, m.ChannelID, m.GuildID, m.Content)
	}

	if m.Author.ID == "192909357585793024" && !strings.HasPrefix(m.Content, "<@") {
		return
	}

	//	if m.Author.ID != "603416910998274059" {
	//		return
	//	}

	// Call message handler and print if it errors
	err := Hn.OnMessage(s, m)
	if err != nil {
		wlog.Err.Print("Error Running command by ", m.Author.String(), " (`", m.Author.ID, "`) in `", m.ChannelID, "` g`", m.GuildID, "`\n>>> ```\n", err, "\n```")
	}
}
