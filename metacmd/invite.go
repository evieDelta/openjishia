package metacmd

import (
	"codeberg.org/eviedelta/drc"
	"codeberg.org/eviedelta/trit"
	"github.com/bwmarrin/discordgo"
)

// invite get the information about inviting the bot, if applicable
var invite = &drc.Command{
	Name:         "invite",
	Manual:       []string{"get the information about inviting the bot, if applicable"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.Unset,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		Listable:    true,
		MinimumArgs: 0,
	},
	Exec: cfInvite,
}

func cfInvite(ctx *drc.Context) error {
	return ctx.Reply(config.Invite)
}
