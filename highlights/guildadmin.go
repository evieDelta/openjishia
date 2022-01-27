package highlights

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"codeberg.org/eviedelta/drc"
	"codeberg.org/eviedelta/trit"
	"github.com/bwmarrin/discordgo"
)

// cConf change settings about the highlights
var cConf = &drc.Command{
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
	ok := guildIsEnabled(ctx.Mes.GuildID)
	list := ctx.Com.Subcommands.ListRecursiveStructured(false, 1)
	ls := "SubCommands:\n"
	for _, x := range list {
		ls += " - " + x.Name + "\n"
	}
	return ctx.Reply("Highlights enabled? ", ok, "\n```\n", ls, "```")
}

func init() {
	cConf.Subcommands.Add(confEnable)
	cConf.Subcommands.Add(confDisable)

	cConf.Subcommands.Add(confBlock)
	cConf.Subcommands.AddAliasString("blk", "hlmod", "block")
	cConf.Subcommands.AddAliasString("b+", "hlmod", "block")
	cConf.Subcommands.Add(confUnblock)
	cConf.Subcommands.AddAliasString("ublk", "hlmod", "unblock")
	cConf.Subcommands.AddAliasString("unblk", "hlmod", "unblock")
	cConf.Subcommands.AddAliasString("b-", "hlmod", "unblock")
	cConf.Subcommands.Add(confBlocking)
	cConf.Subcommands.AddAliasString("bls", "hlmod", "blocking")
	cConf.Subcommands.AddAliasString("blocklist", "hlmod", "blocking")
	cConf.Subcommands.AddAliasString("b=", "hlmod", "blocking")
}

// confEnable highlights in a server
var confEnable = &drc.Command{
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
	Exec: cfConfEnable,
}

func cfConfEnable(ctx *drc.Context) error {
	err := guildSetEnabled(ctx.Mes.GuildID, true)
	if err != nil {
		return err
	}
	return ctx.Reply("Highlights are now on")
}

// confDisable highlights in a server
var confDisable = &drc.Command{
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
	Exec: cfConfDisable,
}

func cfConfDisable(ctx *drc.Context) error {
	err := guildSetEnabled(ctx.Mes.GuildID, false)
	if err != nil {
		return err
	}
	return ctx.Reply("Highlights are now off")
}

// cMod moderation commands for highlights
var cMod = &drc.Command{
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
	ok := guildIsEnabled(ctx.Mes.GuildID)
	list := ctx.Com.Subcommands.ListRecursiveStructured(false, 1)
	ls := "SubCommands:\n"
	for _, x := range list {
		ls += " - " + x.Name + "\n"
	}
	return ctx.Reply("Highlights enabled? ", ok, "\n```\n", ls, "```")
}

func init() {
	cMod.Subcommands.Add(modView)
	cMod.Subcommands.Add(modRemove)
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
	list, err := userListHighlights(user.User.ID, ctx.Mes.GuildID)
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
	err = userRemoveHighlight(user.User.ID, ctx.Mes.GuildID, tem)
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

// confBlock a channel or category globally for a guild
var confBlock = &drc.Command{
	Name:         "block",
	Manual:       []string{"blocks a channel or category globally for a guild"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.Unset,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		Listable:    true,
		MinimumArgs: 1,
	},
	Exec: cfConfBlock,
}

func cfConfBlock(ctx *drc.Context) error {
	if !guildIsEnabled(ctx.Mes.GuildID) {
		return ctx.Reply(notEnabledMessage)
	}

	arg := ""
	res := ""

	{
		c, _, err := ctx.Args[0].Channel(ctx)
		if err != nil {
			return err
		}
		arg = c.ID
		res = c.Name
	}

	// check if its already blocked
	{
		list, err := guildBlockedChannels(ctx.Mes.GuildID)
		if err != nil {
			return err
		}
		for _, x := range list {
			if x == arg {
				return drc.NewFailure(nil, "that channel is already blocked")
			}
		}
	}

	err := guildBlockChannel(ctx.Mes.GuildID, arg)
	if err != nil {
		return err
	}

	g, err := ctx.Ses.State.Guild(ctx.Mes.GuildID)
	if err != nil {
		return err
	}

	_, err = ctx.Ses.ChannelMessageSendEmbed(ctx.Mes.ChannelID, &discordgo.MessageEmbed{
		Description: "Added ``" + res + "`` to the server blocklist,",
		Color:       ctx.Ses.State.UserColor(ctx.Mes.Author.ID, ctx.Mes.ChannelID),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			IconURL: g.IconURL(),
		},
	})
	return err
}

// confUnblock a channel or category globally for a guild
var confUnblock = &drc.Command{
	Name:         "unblock",
	Manual:       []string{"unblocks a channel from the guild level"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.Unset,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		Listable:    true,
		MinimumArgs: 1,
		BoolFlags: map[string]bool{
			"index": false,
		},
	},
	Exec: cfConfUnblock,
}

func cfConfUnblock(ctx *drc.Context) error {
	arg := ""
	res := ""

	if ctx.BoolFlags["index"] {
		blocklist, err := guildBlockedChannels(ctx.Mes.GuildID)
		if err != nil {
			return err
		}
		sort.Strings(blocklist)

		i, err := ctx.Args[0].Int()
		if err != nil {
			return err
		}
		if i >= len(blocklist) {
			return drc.NewParseError(nil, "invalid index (use `hl blocking` to view)")
		}

		arg = blocklist[i]
		res = "<#" + arg + "> / " + ctx.RawArgs[0]
	} else {
		c, _, err := ctx.Args[0].Channel(ctx)
		if err != nil {
			return err
		}
		arg = c.ID
		res = "`" + c.Name + "`"
	}

	err := guildUnblockChannel(ctx.Mes.GuildID, arg)
	if err != nil {
		return err
	}

	g, err := ctx.Ses.State.Guild(ctx.Mes.GuildID)
	if err != nil {
		return err
	}

	_, err = ctx.Ses.ChannelMessageSendEmbed(ctx.Mes.ChannelID, &discordgo.MessageEmbed{
		Description: "Removed " + res + " from the server blocklist,",
		Color:       ctx.Ses.State.UserColor(ctx.Mes.Author.ID, ctx.Mes.ChannelID),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			IconURL: g.IconURL(),
		},
	})
	return err
}

// confBlocking views all the channels blocked globally for a guild
var confBlocking = &drc.Command{
	Name:         "blocking",
	Manual:       []string{"views all guild blocked channels"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.Unset,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		Listable:    true,
		MinimumArgs: 0,
	},
	Exec: cfConfBlocking,
}

func cfConfBlocking(ctx *drc.Context) error {
	blocklist, err := guildBlockedChannels(ctx.Mes.GuildID)
	if err != nil {
		return err
	}
	sort.Strings(blocklist)
	resp := ""

	for i, id := range blocklist {
		resp += fmt.Sprintf("%2.1v. <#%v>\n", i, id)
	}
	if len(blocklist) == 0 {
		resp = "No channels are blocked"
	}

	g, err := ctx.Ses.State.Guild(ctx.Mes.GuildID)
	if err != nil {
		return err
	}

	_, err = ctx.Ses.ChannelMessageSendEmbed(ctx.Mes.ChannelID, &discordgo.MessageEmbed{
		Description: "\u200B" + resp,
		Color:       ctx.Ses.State.UserColor(ctx.Mes.Author.ID, ctx.Mes.ChannelID),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			IconURL: g.IconURL(),
		},
	})
	return err
}
