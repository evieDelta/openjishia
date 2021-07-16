package module

import "codeberg.org/eviedelta/drc"

// Module contains stuff about modules (obviously)
type Module struct {
	Name string

	Commands      []*drc.Command
	AdminCommands []*drc.Command
	DebugCommands []*drc.Command

	DgoHandlers []interface{}

	InitFunc  func(*Module) error
	OpenFunc  func(*Module) error
	CloseFunc func(*Module)

	Config      interface{}
	ConfigFound bool
}
