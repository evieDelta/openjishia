package highlights

import (
	"strings"
	"time"

	"codeberg.org/eviedelta/drc"
	"codeberg.org/eviedelta/trit"
	"github.com/bwmarrin/discordgo"
)

// hlconf change settings about the highlights
var hlconf = &drc.Command{
	Name:         "hlconf",
	Manual:       []string{"change settings about the highlights"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.Unset,
		Discord:  discordgo.PermissionManageServer,
	},
	Config: drc.CfgCommand{
		Listable:    true,
		MinimumArgs: 0,
	},
	Exec: cfHlconf,
}

func cfHlconf(ctx *drc.Context) error {
	ok := isguildenabled(ctx.Mes.GuildID)
	list := ctx.Com.Subcommands.ListRecursiveStructured(false, 1)
	ls := "SubCommands:\n"
	for _, x := range list {
		ls += " - " + x.Name + "\n"
	}
	return ctx.Reply("Highlights enabled? ", ok, "\n```\n", ls, "```")
}

func init() {
	hlconf.Subcommands.Add(hlconfEnable)
	hlconf.Subcommands.Add(hlconfDisable)
}

// enable enables highlights in a server
var hlconfEnable = &drc.Command{
	Name:         "enable",
	Manual:       []string{"enables highlights in a server"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.Unset,
		Discord:  discordgo.PermissionManageServer,
	},
	Config: drc.CfgCommand{
		Listable:    true,
		MinimumArgs: 0,
	},
	Exec: cfEnable,
}

func cfEnable(ctx *drc.Context) error {
	err := setguildenable(ctx.Mes.GuildID, true)
	if err != nil {
		return err
	}
	return ctx.Reply("Highlights are now on")
}

// disable disables highlights in a server
var hlconfDisable = &drc.Command{
	Name:         "disable",
	Manual:       []string{"disables highlights in a server"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.Unset,
		Discord:  discordgo.PermissionManageServer,
	},
	Config: drc.CfgCommand{
		Listable:    true,
		MinimumArgs: 0,
	},
	Exec: cfDisable,
}

func cfDisable(ctx *drc.Context) error {
	err := setguildenable(ctx.Mes.GuildID, false)
	if err != nil {
		return err
	}
	return ctx.Reply("Highlights are now off")
}

// hlmod moderation commands for highlights
var hlmod = &drc.Command{
	Name:         "hlmod",
	Manual:       []string{"moderation commands for highlights"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.Unset,
		Discord:  discordgo.PermissionManageNicknames,
	},
	Config: drc.CfgCommand{
		Listable:    false,
		MinimumArgs: 0,
	},
	Exec: cfHlmod,
}

func cfHlmod(ctx *drc.Context) error {
	ok := isguildenabled(ctx.Mes.GuildID)
	list := ctx.Com.Subcommands.ListRecursiveStructured(false, 1)
	ls := "SubCommands:\n"
	for _, x := range list {
		ls += " - " + x.Name + "\n"
	}
	return ctx.Reply("Highlights enabled? ", ok, "\n```\n", ls, "```")
}

func init() {
	hlmod.Subcommands.Add(modView)
	hlmod.Subcommands.Add(modRemove)
}

// view views another users highlights
var modView = &drc.Command{
	Name:         "view",
	Manual:       []string{"views another users highlights"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.Unset,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		Listable:    false,
		MinimumArgs: 1,
	},
	Exec: cfView,
}

func cfView(ctx *drc.Context) error {
	user, _, err := ctx.Args[0].Member(ctx)
	if err != nil {
		return err
	}

	var ls string
	list, err := userlshl(user.User.ID, ctx.Mes.GuildID)
	if err != nil {
		return err
	}
	for _, x := range list {
		ls += x + "\n"
	}

	return ctx.ReplyEmbed(&discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    user.User.String() + "'s Highlights",
			IconURL: user.User.AvatarURL("128"),
		},
		Description: "```\n" + ls + "```",
		Color:       ctx.Ses.State.UserColor(ctx.Mes.Author.ID, ctx.Mes.ChannelID),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	})
}

// remove removes a highlight from a user
var modRemove = &drc.Command{
	Name:         "remove",
	Manual:       []string{"removes a highlight from a user"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.Unset,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		Listable:    false,
		MinimumArgs: 2,
	},
	Exec: cfModRemove,
}

func cfModRemove(ctx *drc.Context) error {
	user, _, err := ctx.Args[0].Member(ctx)
	if err != nil {
		return err
	}

	tem := strings.Join(ctx.RawArgs[1:], " ")
	err = userremhl(user.User.ID, ctx.Mes.GuildID, tem)
	if err != nil {
		return err
	}

	_, err = ctx.Ses.ChannelMessageSendEmbed(ctx.Mes.ChannelID, &discordgo.MessageEmbed{
		//		Author: &discordgo.MessageEmbedAuthor{
		//			Name:    ctx.Mes.Author.Username + "'s Highlights",
		//			IconURL: ctx.Mes.Author.AvatarURL("128"),
		//		},
		Description: "Removed ``" + tem + "`` from their highlights,",
		Color:       ctx.Ses.State.UserColor(ctx.Mes.Author.ID, ctx.Mes.ChannelID),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text:    user.User.Username + "'s Highlights",
			IconURL: user.User.AvatarURL("128"),
		},
	})
	return err
}
