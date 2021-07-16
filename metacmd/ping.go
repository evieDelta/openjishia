package metacmd

import (
	"fmt"
	"time"

	"codeberg.org/eviedelta/drc"
	"codeberg.org/eviedelta/trit"
	"github.com/bwmarrin/discordgo"
)

// ping returns the heartbeat pong
var ping = &drc.Command{
	Name:         "ping",
	Manual:       []string{"returns the heartbeat pong"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.False,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		Listable:    true,
		MinimumArgs: 0,
	},
	Exec: cfPing,
}

func cfPing(ctx *drc.Context) error {
	timer := time.Now().UTC()
	hbeat := ctx.Ses.HeartbeatLatency().Round(time.Millisecond).String()
	msg, err := ctx.XReply("Pong!\n", hbeat)
	if err != nil {
		return err
	}
	sendt := time.Now().UTC().Sub(timer)
	_, err = ctx.Ses.ChannelMessageEdit(msg.ChannelID, msg.ID,
		fmt.Sprint("Pong!\nHeartbeat: ", hbeat, "\nMessage: ", sendt.Round(time.Millisecond).String()),
	)
	return err
}
