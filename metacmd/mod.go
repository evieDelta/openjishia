package metacmd

import (
	"time"

	"codeberg.org/eviedelta/drc"
	"github.com/bwmarrin/discordgo"
	"github.com/eviedelta/openjishia/module"
)

// InitTime contains the exact time the bot was completely restarted
var InitTime time.Time

// ReadyTime contains the time of the last onready event
var ReadyTime time.Time

var config = struct {
	Invite string
}{}

// Module contains stuff (bug me to fix comments later)
var Module = &module.Module{
	Name: "metacmd",
	Commands: []*drc.Command{
		hello,
		invite,
		cList,
		ping,
		uptime,
	},
	DgoHandlers: []interface{}{
		onReady,
	},
	OpenFunc: func(m *module.Module) error {
		InitTime = time.Now().UTC()

		if !m.ConfigFound {
			config.Invite = "No invite set"
		}
		return nil
	},
	Config: &config,
}

func onReady(s *discordgo.Session, r *discordgo.Ready) {
	ReadyTime = time.Now().UTC()
}
