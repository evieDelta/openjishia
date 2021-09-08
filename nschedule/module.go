package nschedule

import (
	"embed"

	"codeberg.org/eviedelta/drc"
	"github.com/bwmarrin/discordgo"
	"github.com/eviedelta/openjishia/enpsql"
	"github.com/eviedelta/openjishia/module"
	"github.com/eviedelta/openjishia/nschedule/insched"
	"github.com/eviedelta/openjishia/wlog"
)

// RegisterHandler adds a handler to the global scheduler
func RegisterHandler(name string, h insched.Handler) {
	err := Scheduler.AddHandler(name, h)
	if err != nil {
		panic(err)
	}
}

var db = &Database{}
var Scheduler = insched.New(db)

// Config is the config that contains global config stuff, i mean it is named config so what else would it be
var Config = &localConfig{}

type localConfig struct {
}

// Module contains the module, i mean what else would it contain
var Module = &module.Module{
	Name: "nschedule",
	DgoHandlers: []interface{}{
		onReady,
	},

	DebugCommands: []*drc.Command{
		cSendlater,
		cDeletelater,
	},

	Config: Config,

	InitFunc: func(mod *module.Module) error {
		Scheduler.Log = new(logger)

		return nil
	},
	OpenFunc: func(m *module.Module) error {
		db.ss = enpsql.GetSession()

		return nil
	},
	CloseFunc: func(_ *module.Module) {
		Scheduler.Stop()
	},
}

func onReady(s *discordgo.Session, e *discordgo.Ready) {
	_sendLater.Dg = s
	_msgDeleter.Dg = s
	err := Scheduler.Run()
	if err != nil {
		wlog.Err.Print(err)
	}
}

//go:embed schemas/*
var schemas embed.FS

func init() {
	enpsql.RegisterSchemaFS("nschedule", schemas, "schemas")
}

type logger struct{}

func (l *logger) Error(err error, handler string, stop bool, evt *insched.LoggerEntry) {
	s := ""
	if stop {
		s = "(halting scheduler) "
	}
	if evt != nil {
		wlog.Err.Printf(s+"error from scheduler handler %v: %v\n\t| Deferred? %v for %v Defer Count: %v\n\t| ID: %v, Originally Scheduled for: %v\n\t| Details: %v",
			handler, err, evt.Deferred, evt.DeferLen, evt.DeferCount, evt.ID, evt.Time, evt.Details)
	} else {
		wlog.Err.Printf(s+"error from scheduler handler %v: %v\n", handler, err)
	}
}
