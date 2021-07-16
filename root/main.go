package root

import (
	"github.com/eviedelta/openjishia/highlights"
	"github.com/eviedelta/openjishia/metacmd"
	"github.com/eviedelta/openjishia/module"
	"github.com/eviedelta/openjishia/tree"
)

// see the root dir for the actual main package

//
const (
	BotVersion  = "0.1.12.0-whyDoIHaveThisIliterallyNeverUpdateIt-Edition" // fun fact, this has been 0.1.12 since January 2021
	BotSoftware = "openSpaghettishia"
)

//
var (
	Modules = []*module.Module{
		metacmd.Module,
		highlights.Module,
	}
)

func Main() {
	tree.BotSoftware = BotSoftware
	tree.BotVersion = BotVersion

	tree.InitHandleFlags()
	tree.Setup(Modules)
	tree.StartUntilStop(Modules)
}
