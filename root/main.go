package root

import (
	"codeberg.org/eviedelta/openjishia/metacmd"
	"codeberg.org/eviedelta/openjishia/module"
	"codeberg.org/eviedelta/openjishia/tree"
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
	}
)

func Main() {
	tree.BotSoftware = BotSoftware
	tree.BotVersion = BotVersion

	tree.InitHandleFlags()
	tree.Setup(Modules)
	tree.StartUntilStop(Modules)
}
