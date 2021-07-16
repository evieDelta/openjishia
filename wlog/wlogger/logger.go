package wlogger

import (
	"fmt"
	"log"
	"path"
	"runtime"

	"codeberg.org/eviedelta/dwhook"
)

// Logger is a logger
type Logger struct {
	Color  int
	Avatar string
	Name   string
	Status string

	Webhook *dwhook.Webhook
}

// Send sends a single string to the webhook
func (l *Logger) Send(s string, skip int) error {
	log.Println(s)
	if l.Webhook.ID == "" {
		return nil
	}

	_, f, i, ok := runtime.Caller(1 + skip)

	_, err := l.Webhook.Send(dwhook.Message{
		Username:  l.Name,
		AvatarURL: l.Avatar,

		Embeds: []dwhook.Embed{{
			Title:       l.Status,
			Description: s,
			Color:       l.Color,

			Footer: dwhook.EmbedFooter{
				Text: func() string {
					if ok {
						return fmt.Sprintf("%v :%v", path.Base(f), i)
					}
					return "no source info"
				}(),
			},
		}},
	})
	if err != nil {
		log.Println(err)
	}
	return err
}

// Print prints a message to the webhook using the logic from fmt.Sprint()
func (l *Logger) Print(in ...interface{}) error {
	return l.Send(fmt.Sprint(in...), 1)
}

// Printf prints a message to the webhook using the logic from fmt.Sprintf()
func (l *Logger) Printf(s string, in ...interface{}) error {
	return l.Send(fmt.Sprintf(s, in...), 1)
}
