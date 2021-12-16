package highlights

import (
	"fmt"
	"time"

	"codeberg.org/eviedelta/drc"
	"codeberg.org/eviedelta/trit"
	"github.com/eviedelta/openjishia/nschedule"
)

// dellogtest help
var dellogtest = &drc.Command{
	Name:   "dellogtest",
	Manual: []string{"help"},
	Permissions: drc.Permissions{
		BotAdmin: trit.True,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		Listable:    false,
		MinimumArgs: 0,
	},
	Exec: cfDellogtest,
}

func cfDellogtest(ctx *drc.Context) error {
	m, err := ctx.XReply("uwu to delete in 1 minute")
	if err != nil {
		return err
	}
	addDeletionQueue(m.ID, m.ChannelID, time.Minute)
	return err
}

func addDeletionQueue(m, c string, dur time.Duration) {
	_, err := nschedule.MsgDelete(time.Now().Add(dur), nschedule.MetaMsgDelete{
		ChannelID: c,
		MessageID: m,
	})
	if err != nil {
		fmt.Printf("Error scheduling message delete for %v/%v: %v\n", c, m, err)
	}
}
