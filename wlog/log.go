package wlog

import (
	"codeberg.org/eviedelta/dwhook"
	"github.com/eviedelta/openjishia/wlog/wlogger"
)

// Info is a standard logger for info
var Info = wlogger.Logger{
	Status: "Info",
	Color:  5353325,

	Webhook: dwhook.NewFromID("", ""),
}

// Err is a standard logger for errors
var Err = wlogger.Logger{
	Status: "Error",
	Color:  13971547,

	Webhook: dwhook.NewFromID("", ""),
}

// Spam is a standard logger for spammy debug logging
var Spam = wlogger.Logger{
	Status: "Spam",
	Color:  0x643A2B,

	Webhook: dwhook.NewFromID("", ""),
}
