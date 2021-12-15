package highlights

import (
	"strconv"
	"strings"
	"time"

	"codeberg.org/eviedelta/drc"
	"codeberg.org/eviedelta/trit"
	"github.com/bwmarrin/discordgo"
)

// cmdhighlight some commands for highlighting words and stuff
var cmdhighlight = &drc.Command{
	Name:         "highlight",
	Manual:       []string{"some commands for highlighting words and stuff"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.Unset,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		DMSettings:  drc.CommandDMsBlock,
		Listable:    true,
		MinimumArgs: 0,
	},
	Exec: cfHighlight,
}

func cfHighlight(ctx *drc.Context) error {
	ok := isguildenabled(ctx.Mes.GuildID)
	list := ctx.Com.Subcommands.ListRecursiveStructured(false, 1)
	ls := "SubCommands:\n"
	for _, x := range list {
		ls += " - " + x.Name + "\n"
	}
	return ctx.Reply("Highlights enabled? ", ok, "\n```\n", ls, "```")
}

func init() {
	cmdhighlight.Subcommands.Add(hladd)
	cmdhighlight.Subcommands.Add(hlremove)
	cmdhighlight.Subcommands.AddAliasString("rem", "highlight", "remove")
	cmdhighlight.Subcommands.AddAliasString("delete", "highlight", "remove")
	cmdhighlight.Subcommands.Add(hlList)
	cmdhighlight.Subcommands.Add(hlBlock)
	cmdhighlight.Subcommands.Add(hlUnblock)
	cmdhighlight.Subcommands.Add(hlBlocking)
	cmdhighlight.Subcommands.AddAliasString("blocklist", "highlight", "blocking")
	cmdhighlight.Subcommands.Add(hlTest)
}

// add adds a new word to highlight
var hladd = &drc.Command{
	Name:         "add",
	Manual:       []string{"adds a new word to highlight"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.Unset,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		DMSettings:  drc.CommandDMsBlock,
		Listable:    true,
		MinimumArgs: 1,
	},
	Exec: cfAdd,
}

const notEnabledMessage = "Highlights are currently not enabled on this server\nUse ``hlconf enable`` (req: Manage Server) to enable"

func cfAdd(ctx *drc.Context) error {
	if !isguildenabled(ctx.Mes.GuildID) {
		return ctx.Reply(notEnabledMessage)
	}
	tem := strings.Join(ctx.RawArgs, " ")
	useraddhl(ctx.Mes.Author.ID, ctx.Mes.GuildID, tem)
	_, err := ctx.Ses.ChannelMessageSendEmbed(ctx.Mes.ChannelID, &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    ctx.Mes.Author.Username + "'s Highlights",
			IconURL: ctx.Mes.Author.AvatarURL("128"),
		},
		Description: "Added ``" + tem + "`` to your highlights,",
		Color:       ctx.Ses.State.UserColor(ctx.Mes.Author.ID, ctx.Mes.ChannelID),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		//Footer: &discordgo.MessageEmbedFooter{
		//	Text:    ctx.Mes.Author.Username + "'s Highlights",
		//	IconURL: ctx.Mes.Author.AvatarURL("128"),
		//},
	})
	return err
}

// remove stop highlighting a word
var hlremove = &drc.Command{
	Name:         "remove",
	Manual:       []string{"stop highlighting a word"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.Unset,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		DMSettings:  drc.CommandDMsBlock,
		Listable:    true,
		MinimumArgs: 1,
	},
	Exec: cfRemove,
}

func cfRemove(ctx *drc.Context) error {
	tem := strings.Join(ctx.RawArgs, " ")
	userremhl(ctx.Mes.Author.ID, ctx.Mes.GuildID, tem)

	_, err := ctx.Ses.ChannelMessageSendEmbed(ctx.Mes.ChannelID, &discordgo.MessageEmbed{
		//		Author: &discordgo.MessageEmbedAuthor{
		//			Name:    ctx.Mes.Author.Username + "'s Highlights",
		//			IconURL: ctx.Mes.Author.AvatarURL("128"),
		//		},
		Description: "Removed ``" + tem + "`` from your highlights,",
		Color:       ctx.Ses.State.UserColor(ctx.Mes.Author.ID, ctx.Mes.ChannelID),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text:    ctx.Mes.Author.Username + "'s Highlights",
			IconURL: ctx.Mes.Author.AvatarURL("128"),
		},
	})
	return err
}

// List lists your highlights
var hlList = &drc.Command{
	Name:         "list",
	Manual:       []string{"lists your highlights"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.Unset,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		DMSettings:  drc.CommandDMsBlock,
		Listable:    true,
		MinimumArgs: 0,
	},
	Exec: cfList,
}

func cfList(ctx *drc.Context) error {
	ok := isguildenabled(ctx.Mes.GuildID)
	ls := ""
	list := userlshl(ctx.Mes.Author.ID, ctx.Mes.GuildID)
	for _, x := range list {
		ls += x + "\n"
	}

	_, err := ctx.Ses.ChannelMessageSendEmbed(ctx.Mes.ChannelID, &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    ctx.Mes.Author.Username + "'s Highlights",
			IconURL: ctx.Mes.Author.AvatarURL("128"),
		},
		Description: "Highlights enabled? " + strconv.FormatBool(ok) + "\n```\n" + ls + "```",
		Color:       ctx.Ses.State.UserColor(ctx.Mes.Author.ID, ctx.Mes.ChannelID),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	})
	return err
}

// block block a channel or user from showing up in your highlights
var hlBlock = &drc.Command{
	Name:         "block",
	Manual:       []string{"block a channel or user from showing up in your highlights"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.Unset,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		DMSettings:  drc.CommandDMsBlock,
		Listable:    true,
		MinimumArgs: 1,
		BoolFlags: map[string]bool{
			"user":    false,
			"channel": false,
		},
	},
	Exec: cfBlock,
}

func cfBlock(ctx *drc.Context) error {
	if !isguildenabled(ctx.Mes.GuildID) {
		return ctx.Reply(notEnabledMessage)
	}
	arg := ""
	res := ""
	if ctx.BoolFlags["user"] {
		m, _, err := ctx.Args[0].Member(ctx)
		if err != nil {
			return err
		}
		arg = m.User.ID
		res = m.User.String()
	} else if ctx.BoolFlags["channel"] {
		c, _, err := ctx.Args[0].Channel(ctx)
		if err != nil {
			return err
		}
		arg = c.ID
		res = c.Name
	} else {
		m, _, err := ctx.Args[0].Member(ctx)
		if err == nil {
			arg = m.User.ID
			res = m.User.String()
		} else {
			c, _, err := ctx.Args[0].Channel(ctx)
			if err == nil {
				arg = c.ID
				res = c.Name
			} else {
				return drc.NewFailure(nil, "404 Not found", "Could not find channel or member")
			}
		}
	}
	useraddhlblock(ctx.Mes.Author.ID, ctx.Mes.GuildID, arg)
	_, err := ctx.Ses.ChannelMessageSendEmbed(ctx.Mes.ChannelID, &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    ctx.Mes.Author.Username + "'s Highlights",
			IconURL: ctx.Mes.Author.AvatarURL("128"),
		},
		Description: "Added ``" + res + "`` to your blocklist,",
		Color:       ctx.Ses.State.UserColor(ctx.Mes.Author.ID, ctx.Mes.ChannelID),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		//Footer: &discordgo.MessageEmbedFooter{
		//	Text:    ctx.Mes.Author.Username + "'s Highlights",
		//	IconURL: ctx.Mes.Author.AvatarURL("128"),
		//},
	})
	return err
}

// unblock unblock a channel or user from showing up in your highlights
var hlUnblock = &drc.Command{
	Name:         "unblock",
	Manual:       []string{"unblock a channel or user from showing up in your highlights"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.Unset,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		DMSettings:  drc.CommandDMsBlock,
		Listable:    true,
		MinimumArgs: 1,
		BoolFlags: map[string]bool{
			"user":    false,
			"channel": false,
		},
	},
	Exec: cfUnblock,
}

func cfUnblock(ctx *drc.Context) error {
	arg := ""
	res := ""
	if ctx.BoolFlags["user"] {
		m, _, err := ctx.Args[0].Member(ctx)
		if err != nil {
			return err
		}
		arg = m.User.ID
		res = m.User.String()
	} else if ctx.BoolFlags["channel"] {
		c, _, err := ctx.Args[0].Channel(ctx)
		if err != nil {
			return err
		}
		arg = c.ID
		res = c.Name
	} else {
		m, _, err := ctx.Args[0].Member(ctx)
		if err == nil {
			arg = m.User.ID
			res = m.User.String()
		} else {
			c, _, err := ctx.Args[0].Channel(ctx)
			if err == nil {
				arg = c.ID
				res = c.Name
			} else {
				return drc.NewFailure(nil, "404 Not found", "Could not find channel or member")
			}
		}
	}
	userremhlblock(ctx.Mes.Author.ID, ctx.Mes.GuildID, arg)
	_, err := ctx.Ses.ChannelMessageSendEmbed(ctx.Mes.ChannelID, &discordgo.MessageEmbed{
		//Author: &discordgo.MessageEmbedAuthor{
		//	Name:    ctx.Mes.Author.Username + "'s Highlights",
		//	IconURL: ctx.Mes.Author.AvatarURL("128"),
		//},
		Description: "Removed ``" + res + "`` from your blocklist",
		Color:       ctx.Ses.State.UserColor(ctx.Mes.Author.ID, ctx.Mes.ChannelID),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text:    ctx.Mes.Author.Username + "'s Highlights",
			IconURL: ctx.Mes.Author.AvatarURL("128"),
		},
	})
	return err
}

// blocking lists who you are blocking
var hlBlocking = &drc.Command{
	Name:         "blocking",
	Manual:       []string{"lists who you are blocking"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.Unset,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		DMSettings:  drc.CommandDMsBlock,
		Listable:    true,
		MinimumArgs: 0,
	},
	Exec: cfBlocking,
}

func cfBlocking(ctx *drc.Context) error {
	blls := userlshlblock(ctx.Mes.Author.ID, ctx.Mes.GuildID)
	usls := ""
	chls := ""

	guildChannels, err := ctx.Ses.GuildChannels(ctx.Mes.GuildID)
	if err != nil {
		return err
	}
	for _, id := range blls {
		isChannel := false
		for _, ch := range guildChannels {
			if id == ch.ID {
				isChannel = true
				break
			}
		}
		if isChannel {
			chls += "<#" + id + ">\n"
		} else {
			usls += "<@!" + id + ">\n"
		}
	}

	_, err = ctx.Ses.ChannelMessageSendEmbed(ctx.Mes.ChannelID, &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    ctx.Mes.Author.Username + "'s Blocking",
			IconURL: ctx.Mes.Author.AvatarURL("128"),
		},
		Fields: []*discordgo.MessageEmbedField{{
			Name:   "Channels",
			Value:  "\u200B" + chls,
			Inline: true,
		}, {
			Name:   "Users",
			Value:  "\u200B" + usls,
			Inline: true,
		}},
		Color:     ctx.Ses.State.UserColor(ctx.Mes.Author.ID, ctx.Mes.ChannelID),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
	return err
}

// test tests if a string matches any of your highlights
var hlTest = &drc.Command{
	Name:         "test",
	Manual:       []string{"tests if a string matches any of your highlights"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.Unset,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		DMSettings:  drc.CommandDMsBlock,
		Listable:    true,
		MinimumArgs: 1,
	},
	Exec: cfTest,
}

var truefalseemote = map[bool]string{
	true:  "✅",
	false: "❌",
}

func cfTest(ctx *drc.Context) error {
	if !isguildenabled(ctx.Mes.GuildID) {
		return ctx.Reply(notEnabledMessage)
	}

	tem := strings.Join(ctx.RawArgs, " ")

	highlights := userlshl(ctx.Mes.Author.ID, ctx.Mes.GuildID)
	if len(highlights) == 0 {
		return ctx.Reply("You don't seem to have any highlights")
	}

	doots := make(map[string]bool, len(highlights))

	for _, word := range highlights {
		if start, end := checkHighlight(tem, word, ctx.Mes.Author.ID, ctx.Mes); start >= 0 && end >= 0 {
			doots[word] = true
		} else {
			doots[word] = false
		}
	}

	list := ""
	for y, x := range doots {
		list += truefalseemote[x] + " " + y + "\n"
	}

	return ctx.ReplyEmbed(&discordgo.MessageEmbed{
		Fields: []*discordgo.MessageEmbedField{{
			Name:  "Words",
			Value: "\u200B" + list,
		}, {
			Name:  "Message",
			Value: "> " + tem,
		}},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    ctx.Mes.Author.Username + "'s Highlights",
			IconURL: ctx.Mes.Author.AvatarURL("128"),
		},
		Color:     ctx.Ses.State.UserColor(ctx.Mes.Author.ID, ctx.Mes.ChannelID),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}
