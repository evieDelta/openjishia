package nschedule

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/eviedelta/openjishia/nschedule/insched"
)

func MsgDelete(at time.Time, meta MetaMsgDelete) (uint64, error) {
	e, err := Scheduler.Schedule(at, keyMsgDeleter, meta)
	if err != nil {
		return 0, err
	}
	return e.ID, err
}

type MetaMsgDelete struct {
	ChannelID string
	MessageID string
}

var _msgDeleter = &msgDeleter{
	Conf: insched.HandlerConfig{
		Precise:      false,
		ScanPeriod:   time.Hour / 2, // we don't expect much yet so we scan every half hour
		DeferOnPanic: -1,
	},
}

type msgDeleter struct {
	Dg   *discordgo.Session
	Conf insched.HandlerConfig
}

func (h *msgDeleter) Call(e insched.Entry) (Defer time.Duration, err error) {
	meta := new(MetaMsgDelete)
	err = e.UnmarshalDetails(&meta)
	if err != nil {
		return 0, err
	}

	err = h.Dg.ChannelMessageDelete(meta.ChannelID, meta.MessageID)

	return 0, err
}

func (h *msgDeleter) Config() insched.HandlerConfig {
	return h.Conf
}

const keyMsgDeleter = "nschedule:message_deleter"

func init() {
	err := Scheduler.AddHandler(keyMsgDeleter, _msgDeleter)
	if err != nil {
		panic(err)
	}
}
