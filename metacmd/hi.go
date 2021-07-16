package metacmd

import (
	"codeberg.org/eviedelta/drc"
	"codeberg.org/eviedelta/trit"
	"github.com/bwmarrin/discordgo"
)

// hello say hi
var hello = &drc.Command{
	Name:         "hello",
	Manual:       []string{"say hi"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.Unset,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		Listable:    true,
		MinimumArgs: 0,
	},
	Exec: cfHello,
}

func cfHello(ctx *drc.Context) error {
	return ctx.Reply("Hello~")
}
