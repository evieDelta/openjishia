package nschedule

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/eviedelta/openjishia/nschedule/insched"
)

// DeferSendMessage schedules a message to be sent at a later time
// currently only debug, don't use
func DeferSendMessage(when time.Time, data SendLaterMeta) (uint64, error) {
	et, err := Scheduler.Schedule(when, sendLaterKey, data)
	return et.ID, err
}

type SendLaterMeta struct {
	Channel  string
	Contents string
}

var _sendLater = &sendLater{
	Conf: insched.HandlerConfig{
		Precise:      false,
		ScanPeriod:   time.Minute * 5,
		DeferOnPanic: -1,
	},
}

type sendLater struct {
	Dg   *discordgo.Session
	Conf insched.HandlerConfig
}

func (sl *sendLater) Call(e insched.Entry) (Defer time.Duration, err error) {
	meta := SendLaterMeta{}
	err = e.UnmarshalDetails(&meta)
	if err != nil {
		return 0, err
	}

	// panic("beenz")

	fmt.Println(meta.Channel)
	fmt.Println(meta.Contents)
	fmt.Println(e.Details)

	_, err = sl.Dg.ChannelMessageSend(meta.Channel,
		meta.Contents+"\n\n"+fmt.Sprintf("ID: %v, Time off from target %v\nnow: `%v`, target: `%v`",
			e.ID, e.Time.Sub(time.Now()), time.Now().UTC(), e.Time.UTC()))
	if err == nil {
		return 0, nil
	}

	if e.DeferCount > 0 {
		return 0, err
	}

	switch {
	case strings.Contains(err.Error(), "net/http"):
		fallthrough
	case strings.Contains(err.Error(), "dial tcp"):
		return time.Minute * 5, err
	}

	return 0, err
}

func (sl *sendLater) Config() insched.HandlerConfig {
	return sl.Conf
}

const sendLaterKey = "nschedule:message_send"

func init() {
	err := Scheduler.AddHandler(sendLaterKey, _sendLater)
	if err != nil {
		panic(err)
	}
}
