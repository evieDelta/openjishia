package highlights

import (
	"strings"
	"time"

	"codeberg.org/eviedelta/drc"
	"codeberg.org/eviedelta/trit"
	"github.com/bwmarrin/discordgo"
	"github.com/eviedelta/openjishia/wlog"
)

// cAdmin some commands for highlight admin and debug
var cAdmin = &drc.Command{
	Name:         "hladmin",
	Manual:       []string{"some commands for highlight admin and debug"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.True,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		Listable:    false,
		MinimumArgs: 0,
	},
	Exec: cfHladmin,
}

func cfHladmin(ctx *drc.Context) error {
	list := ctx.Com.Subcommands.ListRecursiveStructured(false, 1)
	ls := "SubCommands:\n"
	for _, x := range list {
		ls += " - " + x.Name + "\n"
	}
	return ctx.DumpReply("hladmin", list)
}

func init() {
	cAdmin.Subcommands.Add(debugRatelimitstatus)
	cAdmin.Subcommands.Add(debugRateclean)
	cAdmin.Subcommands.Add(debugForcerateclear)

	cAdmin.Subcommands.Add(debugDellogtest)

	cAdmin.Subcommands.Add(debugView)
	cAdmin.Subcommands.Add(debugRemove)
}

// debugRatelimitstatus gets the current status of the ratelimits
var debugRatelimitstatus = &drc.Command{
	Name:         "ratelimitstatus",
	Manual:       []string{"gets the current status of the ratelimits"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.True,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		Listable:    false,
		MinimumArgs: 0,
	},
	Exec: cfRatelimitstatus,
}

func cfRatelimitstatus(ctx *drc.Context) error {
	ratelimitLock.RLock()
	defer ratelimitLock.RUnlock()

	ls := "> Active\n"
	for i, x := range ratelimit {
		if isLimited(i) {
			ls += i + " | " + time.Until(x).Truncate(time.Millisecond).String() + "\n"
		}
	}
	ls += "> Inactive\n"
	for i, x := range ratelimit {
		if !isLimited(i) {
			ls += i + " | " + time.Until(x).Truncate(time.Millisecond).String() + "\n"
		}
	}

	return ctx.DumpReply("highlight ratelimit status", ls)
}

// debugRateclean force runs the routine to clean the highlight ratelimit cache
var debugRateclean = &drc.Command{
	Name:         "rateclean",
	Manual:       []string{"force runs the routine to clean the highlight ratelimit cache"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.True,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		Listable:    false,
		MinimumArgs: 0,
	},
	Exec: cfRateclean,
}

func cfRateclean(ctx *drc.Context) error {
	wlog.Info.Print("Running ratelimit cache cleaner...")

	ratelimitLock.Lock()
	defer ratelimitLock.Unlock()

	for i, x := range ratelimit {
		if x.Before(time.Now()) {
			delete(ratelimit, i)
		}
	}
	return ctx.Reply("done")
}

// debugForcerateclear forcefully cleans the ratelimit cache
var debugForcerateclear = &drc.Command{
	Name:         "forcerateclear",
	Manual:       []string{"forcefully clears the entire ratelimit cache"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.True,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		Listable:    false,
		MinimumArgs: 0,
	},
	Exec: cfForcerateclear,
}

func cfForcerateclear(ctx *drc.Context) error {
	wlog.Info.Print("Running forced ratelimit cache cleaner...")
	time.Sleep(time.Second * 5)

	ratelimitLock.Lock()
	defer ratelimitLock.Unlock()

	for i := range ratelimit {
		delete(ratelimit, i)
	}
	return ctx.Reply("done")
}

// debugView any other users highlights
var debugView = &drc.Command{
	Name:         "view",
	Manual:       []string{"views another users highlights"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.Unset,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		Listable:    false,
		MinimumArgs: 2,
	},
	Exec: cfDebugView,
}

func cfDebugView(ctx *drc.Context) error {
	user, err := ctx.Ses.User(ctx.RawArgs[1])
	if err != nil {
		return err
	}

	var ls string
	list, err := userListHighlights(ctx.RawArgs[1], ctx.RawArgs[0])
	if err != nil {
		return err
	}
	for _, x := range list {
		ls += x + "\n"
	}

	return ctx.ReplyEmbed(&discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    user.String() + "'s Highlights",
			IconURL: user.AvatarURL("128"),
		},
		Description: "```\n" + ls + "```",
		Color:       ctx.Ses.State.UserColor(ctx.Mes.Author.ID, ctx.Mes.ChannelID),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	})
}

// remove removes a highlight from a user
var debugRemove = &drc.Command{
	Name:         "remove",
	Manual:       []string{"removes a highlight from a user"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.Unset,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		Listable:    false,
		MinimumArgs: 3,
	},
	Exec: cfDebugRemove,
}

func cfDebugRemove(ctx *drc.Context) error {
	user, err := ctx.Ses.User(ctx.RawArgs[1])
	if err != nil {
		return err
	}

	tem := strings.Join(ctx.RawArgs[2:], " ")
	userRemoveHighlight(user.ID, ctx.RawArgs[0], tem)

	_, err = ctx.Ses.ChannelMessageSendEmbed(ctx.Mes.ChannelID, &discordgo.MessageEmbed{
		//		Author: &discordgo.MessageEmbedAuthor{
		//			Name:    ctx.Mes.Author.Username + "'s Highlights",
		//			IconURL: ctx.Mes.Author.AvatarURL("128"),
		//		},
		Description: "Removed ``" + tem + "`` from their highlights,",
		Color:       ctx.Ses.State.UserColor(ctx.Mes.Author.ID, ctx.Mes.ChannelID),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text:    user.Username + "'s Highlights",
			IconURL: user.AvatarURL("128"),
		},
	})
	return err
}
