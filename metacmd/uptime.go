package metacmd

import (
	"strconv"
	"time"

	"codeberg.org/eviedelta/drc"
	"codeberg.org/eviedelta/trit"
	"github.com/bwmarrin/discordgo"
)

func what(then, now time.Time) string {
	return strconv.Itoa(
		int(
			(now.Sub(then).Truncate(time.Hour/24) / 24).Hours(),
		),
	) + " Days, " +
		now.Sub(
			then.Add(
				((now.Sub(then) / 24).Truncate(time.Hour))*24,
			),
		).Truncate(time.Second).String()
}

// uptime reads out the bots current uptime
var uptime = &drc.Command{
	Name:         "uptime",
	Manual:       []string{"reads out the bots current uptime"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.Unset,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		Listable:    true,
		MinimumArgs: 0,
	},
	Exec: cfUptime,
}

func cfUptime(ctx *drc.Context) error {
	return ctx.ReplyEmbed(&discordgo.MessageEmbed{
		Fields: []*discordgo.MessageEmbedField{{
			Name:   "Runtime",
			Value:  what(InitTime, time.Now().UTC()) + "\nSince: " + InitTime.Format("2006-01-02 15:04:05"),
			Inline: true,
		}, {
			Name:   "Uptime",
			Value:  what(ReadyTime, time.Now().UTC()) + "\nSince: " + ReadyTime.Format("2006-01-02 15:04:05"),
			Inline: true,
		}},
		Color:     ctx.Ses.State.UserColor(ctx.Ses.State.User.ID, ctx.Mes.ChannelID),
		Timestamp: time.Now().Format(time.RFC3339),
	})
}
