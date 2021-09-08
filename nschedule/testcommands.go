package nschedule

import (
	"fmt"
	"strings"
	"time"

	"codeberg.org/eviedelta/detctime/durationparser"
	"codeberg.org/eviedelta/drc"
	"codeberg.org/eviedelta/trit"
	"github.com/bwmarrin/discordgo"
)

// deletelater deletes a message later
var cDeletelater = &drc.Command{
	Name:         "deletelater",
	Manual:       []string{"deletes a message later"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.True,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		Listable:    false,
		MinimumArgs: 2,
		DataFlags: map[string]string{
			"c": "",
		},
	},
	Exec: cfDeletelater,
}

func cfDeletelater(ctx *drc.Context) error {
	ch := ctx.Mes.ChannelID
	c, ok, err := ctx.Flags["c"].Channel(ctx)
	if ok && err != nil {
		return err
	}
	if ok {
		ch = c.ID
	}

	d, err := durationparser.Parse(ctx.RawArgs[0])
	if err != nil {
		return err
	}
	t := time.Now().Add(d)

	msg, err := ctx.Ses.ChannelMessage(ch, ctx.RawArgs[1])
	if err != nil {
		return err
	}

	id, err := MsgDelete(t, MetaMsgDelete{
		ChannelID: msg.ChannelID,
		MessageID: msg.ID,
	})
	if err != nil {
		return err
	}

	return ctx.ReplyEmbed(&discordgo.MessageEmbed{
		Description: fmt.Sprintf("okay, i will delete msg:%v from <#%v>, at <t:%v:f> <t:%v:R>", msg.ID, msg.ChannelID, t.Unix(), t.Unix()),
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("ID: %v", id),
		},
		Timestamp: t.Format(time.RFC3339),
	})
}

// sendlater sends a message later
var cSendlater = &drc.Command{
	Name:         "sendlater",
	Manual:       []string{"sends a message later"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.True,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		Listable:    false,
		MinimumArgs: 2,

		DataFlags: map[string]string{
			"c": "",
		},
	},
	Exec: cfSendlater,
}

func cfSendlater(ctx *drc.Context) error {
	ch := ctx.Mes.ChannelID
	c, ok, err := ctx.Flags["c"].Channel(ctx)
	if ok && err != nil {
		return err
	}
	if ok {
		ch = c.ID
	}

	d, err := durationparser.Parse(ctx.RawArgs[0])
	if err != nil {
		return err
	}
	t := time.Now().Add(d)

	id, err := DeferSendMessage(t, SendLaterMeta{
		Channel:  ch,
		Contents: strings.Join(ctx.RawArgs[1:], " "),
	})
	if err != nil {
		return err
	}

	return ctx.Replyf("okay, scheduled message for %v, ID: %v", t.Format(time.RFC3339), id)
}
