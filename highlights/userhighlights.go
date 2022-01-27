package highlights

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"codeberg.org/eviedelta/drc"
	"codeberg.org/eviedelta/trit"
	"github.com/bwmarrin/discordgo"
)

// cHighlight some commands for highlighting words and stuff
var cHighlight = &drc.Command{
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
		BoolFlags: map[string]bool{
			"old": true,
		},
	},
	Exec: cfHighlight,
}

func cfHighlight(ctx *drc.Context) error {
	ok := guildIsEnabled(ctx.Mes.GuildID)
	if ctx.BoolFlags["old"] {
		list := ctx.Com.Subcommands.ListRecursiveStructured(false, 1)
		ls := "SubCommands:\n"
		for _, x := range list {
			ls += " - " + x.Name + "\n"
		}
		return ctx.Reply("Highlights enabled? ", ok, "\n```\n", ls, "```")
	}

	//	toggle := "Highlights are Disabled here"
	//	if ok {
	//		toggle = "Highlights are Enabled"
	//	}
	//
	//	doc, ok := helputil.Get("highlights")
	//	if !ok {
	//		return drc.NewFailure(nil, "the help appears to be missing")
	//	}

	return ctx.Reply("reworked help command under construction")
}

func init() {
	cHighlight.Subcommands.Add(hlAdd)
	cHighlight.Subcommands.AddAliasString("a", "highlight", "add")
	cHighlight.Subcommands.AddAliasString("+", "highlight", "add")

	cHighlight.Subcommands.Add(hlRemove)
	cHighlight.Subcommands.AddAliasString("rm", "highlight", "remove")
	cHighlight.Subcommands.AddAliasString("rem", "highlight", "remove")
	cHighlight.Subcommands.AddAliasString("d", "highlight", "remove")
	cHighlight.Subcommands.AddAliasString("del", "highlight", "remove")
	cHighlight.Subcommands.AddAliasString("delete", "highlight", "remove")
	cHighlight.Subcommands.AddAliasString("-", "highlight", "remove")

	cHighlight.Subcommands.Add(hlClear)
	cHighlight.Subcommands.AddAliasString("c", "highlight", "clear")
	cHighlight.Subcommands.AddAliasString("cl", "highlight", "clear")
	cHighlight.Subcommands.AddAliasString("dall", "highlight", "clear")
	cHighlight.Subcommands.AddAliasString("rall", "highlight", "clear")

	cHighlight.Subcommands.Add(hlList)
	cHighlight.Subcommands.AddAliasString("l", "highlight", "list")
	cHighlight.Subcommands.AddAliasString("ls", "highlight", "list")
	cHighlight.Subcommands.AddAliasString("=", "highlight", "list")

	cHighlight.Subcommands.Add(hlBlock)
	cHighlight.Subcommands.AddAliasString("blk", "highlight", "block")
	cHighlight.Subcommands.AddAliasString("b+", "highlight", "block")
	cHighlight.Subcommands.Add(hlUnblock)
	cHighlight.Subcommands.AddAliasString("ublk", "highlight", "unblock")
	cHighlight.Subcommands.AddAliasString("unblk", "highlight", "unblock")
	cHighlight.Subcommands.AddAliasString("b-", "highlight", "unblock")
	cHighlight.Subcommands.Add(hlBlocking)
	cHighlight.Subcommands.AddAliasString("bls", "highlight", "blocking")
	cHighlight.Subcommands.AddAliasString("blocklist", "highlight", "blocking")
	cHighlight.Subcommands.AddAliasString("b=", "highlight", "blocking")
	cHighlight.Subcommands.Add(hlTest)
}

// hlAdd adds a new word to highlight
var hlAdd = &drc.Command{
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
	Exec: cfHlAdd,
}

func cfHlAdd(ctx *drc.Context) error {
	if !guildIsEnabled(ctx.Mes.GuildID) {
		return ctx.Reply(notEnabledMessage)
	}

	tem := strings.Join(ctx.RawArgs, " ")
	tem = NormaliseString(tem)

	list, err := userListHighlights(ctx.Mes.Author.ID, ctx.Mes.GuildID)
	if err != nil {
		return err
	}
	for _, x := range list {
		if x == tem {
			return drc.NewParseError(nil, "you are already highlighting that")
		}
	}

	if len(list) >= maxHighlights {
		return drc.NewFailure(nil, "you have reached the highlight limit (max: "+maxHighlightsString+")")
	}
	if len([]rune(tem)) > maxHighlightLength {
		return drc.NewFailure(nil, "highlight text too long (max: "+maxHighlightLengthString+" char)")
	}

	err = userAddHighlight(ctx.Mes.Author.ID, ctx.Mes.GuildID, tem)
	if err != nil {
		return err
	}

	_, err = ctx.Ses.ChannelMessageSendEmbed(ctx.Mes.ChannelID, &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    ctx.Mes.Author.Username + "'s Highlights",
			IconURL: ctx.Mes.Author.AvatarURL("128"),
		},
		Description: "Added ``" + tem + "`` to your highlights",
		Color:       ctx.Ses.State.UserColor(ctx.Mes.Author.ID, ctx.Mes.ChannelID),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		//Footer: &discordgo.MessageEmbedFooter{
		//	Text:    ctx.Mes.Author.Username + "'s Highlights",
		//	IconURL: ctx.Mes.Author.AvatarURL("128"),
		//},
	})
	return err
}

// hlRemove stops highlighting a word
var hlRemove = &drc.Command{
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
		BoolFlags: map[string]bool{
			"index": false,
		},
	},
	Exec: cfHlRemove,
}

func cfHlRemove(ctx *drc.Context) error {
	tem := strings.Join(ctx.RawArgs, " ")

	if ctx.BoolFlags["index"] {
		hls, err := userListHighlights(ctx.Mes.Author.ID, ctx.Mes.GuildID)
		if err != nil {
			return err
		}
		sort.Strings(hls)

		i, err := ctx.Args[0].Int()
		if err != nil {
			return err
		}
		if i >= len(hls) {
			return drc.NewParseError(nil, "index too high (`hl list` to view)")
		}

		tem = hls[i]
	}

	err := userRemoveHighlight(ctx.Mes.Author.ID, ctx.Mes.GuildID, tem)
	if err != nil {
		return err
	}

	_, err = ctx.Ses.ChannelMessageSendEmbed(ctx.Mes.ChannelID, &discordgo.MessageEmbed{
		//		Author: &discordgo.MessageEmbedAuthor{
		//			Name:    ctx.Mes.Author.Username + "'s Highlights",
		//			IconURL: ctx.Mes.Author.AvatarURL("128"),
		//		},
		Description: "Removed ``" + tem + "`` from your highlights",
		Color:       ctx.Ses.State.UserColor(ctx.Mes.Author.ID, ctx.Mes.ChannelID),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text:    ctx.Mes.Author.Username + "'s Highlights",
			IconURL: ctx.Mes.Author.AvatarURL("128"),
		},
	})
	return err
}

// hlClear all highlights
var hlClear = &drc.Command{
	Name:         "clear",
	Manual:       []string{"clears all highlights"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.Unset,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		Listable:    true,
		MinimumArgs: 0,
	},
	Exec: cfHlClear,
}

func cfHlClear(ctx *drc.Context) error {
	hls, err := userListHighlights(ctx.Mes.Author.ID, ctx.Mes.GuildID)
	if err != nil {
		return err
	}
	sort.Strings(hls)

	str := "Removed " + strconv.Itoa(len(hls)) + " Highlights\n```\n"
	for _, x := range hls {
		str += x + "\n"
	}
	str += "```"

	err = userClearHighlights(ctx.Mes.Author.ID, ctx.Mes.GuildID)
	if err != nil {
		return err
	}

	_, err = ctx.Ses.ChannelMessageSendEmbed(ctx.Mes.ChannelID, &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    ctx.Mes.Author.Username + "'s Highlights",
			IconURL: ctx.Mes.Author.AvatarURL("128"),
		},
		Description: str,
		Color:       ctx.Ses.State.UserColor(ctx.Mes.Author.ID, ctx.Mes.ChannelID),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	})
	return err
}

// hlList your highlights
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
	Exec: cfHlList,
}

func cfHlList(ctx *drc.Context) error {
	ok := guildIsEnabled(ctx.Mes.GuildID)
	ls := ""
	list, err := userListHighlights(ctx.Mes.Author.ID, ctx.Mes.GuildID)
	if err != nil {
		return err
	}
	sort.Strings(list)
	for i, x := range list {
		ls += fmt.Sprintf("%2.1v. %v\n", i, x)
	}

	_, err = ctx.Ses.ChannelMessageSendEmbed(ctx.Mes.ChannelID, &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    ctx.Mes.Author.Username + "'s Highlights",
			IconURL: ctx.Mes.Author.AvatarURL("128"),
		},
		Description: "```\n" + ls + "\n```",
		Color:       ctx.Ses.State.UserColor(ctx.Mes.Author.ID, ctx.Mes.ChannelID),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: func() string {
				if ok {
					return ""
				}
				return "Highlights Disabled"
			}(),
		}})
	return err
}

// hlBlock a channel or user from showing up in your highlights
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
	Exec: cfHlBlock,
}

func cfHlBlock(ctx *drc.Context) error {
	if !guildIsEnabled(ctx.Mes.GuildID) {
		return ctx.Reply(notEnabledMessage)
	}
	const (
		channel, user = 1, 2
	)
	arg := ""
	res := ""
	kind := 0
	if ctx.BoolFlags["user"] {
		m, _, err := ctx.Args[0].Member(ctx)
		if err != nil {
			return err
		}
		arg = m.User.ID
		res = m.User.String()
		kind = user
	} else if ctx.BoolFlags["channel"] {
		c, _, err := ctx.Args[0].Channel(ctx)
		if err != nil {
			return err
		}
		arg = c.ID
		res = c.Name
		kind = channel
	} else {
		m, _, err := ctx.Args[0].Member(ctx)
		if err == nil {
			arg = m.User.ID
			res = m.User.String()
			kind = user
		} else {
			c, _, err := ctx.Args[0].Channel(ctx)
			if err == nil {
				arg = c.ID
				res = c.Name
				kind = channel
			} else {
				return drc.NewParseError(nil, "404 Not found", "Could not find channel or member")
			}
		}
	}

	// check if a user is already blocking something to prevent duplicates
	{
		list, err := userBlockedChannels(ctx.Mes.Author.ID, ctx.Mes.GuildID)
		if err != nil {
			return err
		}
		for _, x := range list {
			if x == arg {
				return drc.NewFailure(nil, "you are already blocking that")
			}
		}
		list, err = userBlockedMembers(ctx.Mes.Author.ID, ctx.Mes.GuildID)
		if err != nil {
			return err
		}
		for _, x := range list {
			if x == arg {
				return drc.NewFailure(nil, "you are already blocking that")
			}
		}
	}

	var err error

	switch kind {
	default:
		return drc.NewFailure(nil, "uh oh something did not work")
	case user:
		err = userBlockMember(ctx.Mes.Author.ID, ctx.Mes.GuildID, arg)
	case channel:
		err = userBlockChannel(ctx.Mes.Author.ID, ctx.Mes.GuildID, arg)
	}

	if err != nil {
		return err
	}

	_, err = ctx.Ses.ChannelMessageSendEmbed(ctx.Mes.ChannelID, &discordgo.MessageEmbed{
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

// hlUnblock a channel or user from showing up in your highlights
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
			"index":   false,
		},
	},
	Exec: cfHlUnblock,
}

func cfHlUnblock(ctx *drc.Context) error {
	const (
		channel, user = 1, 2
	)
	arg := ""
	res := ""
	kind := 0
	if ctx.BoolFlags["user"] {
		m, _, err := ctx.Args[0].Member(ctx)
		if err != nil {
			return err
		}
		arg = m.User.ID
		res = "`" + m.User.String() + "`"
		kind = user
	} else if ctx.BoolFlags["channel"] {
		c, _, err := ctx.Args[0].Channel(ctx)
		if err != nil {
			return err
		}
		arg = c.ID
		res = "`" + c.Name + "`"
		kind = channel
	} else if ctx.BoolFlags["index"] {
		ublls, err := userBlockedMembers(ctx.Mes.Author.ID, ctx.Mes.GuildID)
		if err != nil {
			return err
		}
		cblls, err := userBlockedChannels(ctx.Mes.Author.ID, ctx.Mes.GuildID)
		if err != nil {
			return err
		}
		sort.Strings(ublls)
		sort.Strings(cblls)
		blls := append(ublls, cblls...)

		i, err := ctx.Args[0].Int()
		if err != nil {
			return err
		}
		if i >= len(blls) {
			return drc.NewParseError(nil, "invalid index (use `hl blocking` to view)")
		}

		arg = blls[i]
		if i >= len(ublls) && len(cblls) != 0 {
			kind = channel
			res = "<#" + arg + "> / " + ctx.RawArgs[0]
		} else {
			kind = user
			res = "<@" + arg + "> / " + ctx.RawArgs[0]
		}
	} else {
		m, _, err := ctx.Args[0].Member(ctx)
		if err == nil {
			arg = m.User.ID
			res = "`" + m.User.String() + "`"
			kind = user
		} else {
			c, _, err := ctx.Args[0].Channel(ctx)
			if err == nil {
				arg = c.ID
				res = "`" + c.Name + "`"
				kind = channel
			} else {
				return drc.NewFailure(nil, "404 Not found", "Could not find channel or member")
			}
		}
	}

	var err error

	switch kind {
	default:
		return drc.NewFailure(nil, "uh oh something did not a work")
	case user:
		err = userUnblockMember(ctx.Mes.Author.ID, ctx.Mes.GuildID, arg)
	case channel:
		err = userUnblockChannel(ctx.Mes.Author.ID, ctx.Mes.GuildID, arg)
	}

	if err != nil {
		return err
	}

	_, err = ctx.Ses.ChannelMessageSendEmbed(ctx.Mes.ChannelID, &discordgo.MessageEmbed{
		//Author: &discordgo.MessageEmbedAuthor{
		//	Name:    ctx.Mes.Author.Username + "'s Highlights",
		//	IconURL: ctx.Mes.Author.AvatarURL("128"),
		//},
		Description: "Removed " + res + " from your blocklist",
		Color:       ctx.Ses.State.UserColor(ctx.Mes.Author.ID, ctx.Mes.ChannelID),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text:    ctx.Mes.Author.Username + "'s Highlights",
			IconURL: ctx.Mes.Author.AvatarURL("128"),
		},
	})
	return err
}

// hlBlocking lists who you are blocking
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
	Exec: cfHlBlocking,
}

func cfHlBlocking(ctx *drc.Context) error {
	ublls, err := userBlockedMembers(ctx.Mes.Author.ID, ctx.Mes.GuildID)
	if err != nil {
		return err
	}
	cblls, err := userBlockedChannels(ctx.Mes.Author.ID, ctx.Mes.GuildID)
	if err != nil {
		return err
	}
	sort.Strings(ublls)
	sort.Strings(cblls)
	usls := ""
	chls := ""

	for i, id := range ublls {
		usls += fmt.Sprintf("%2.1v. <@!%v>\n", i, id)
	}
	for i, id := range cblls {
		chls += fmt.Sprintf("%2.1v. <#%v>\n", i+len(ublls), id)
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

// hlTest if a string matches any of your highlights
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
	Exec: cfHlTest,
}

func cfHlTest(ctx *drc.Context) error {
	if !guildIsEnabled(ctx.Mes.GuildID) {
		return ctx.Reply(notEnabledMessage)
	}

	tem := strings.Join(ctx.RawArgs, " ")

	highlights, err := userListHighlights(ctx.Mes.Author.ID, ctx.Mes.GuildID)
	if err != nil {
		return err
	}
	if len(highlights) == 0 {
		return ctx.Reply("You don't seem to have any highlights")
	}

	doots := make(map[string]bool, len(highlights))

	for _, word := range highlights {
		if start, end := checkHighlight(NormaliseString(tem), word, ctx.Mes.Author.ID, ctx.Mes); start >= 0 && end >= 0 {
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
