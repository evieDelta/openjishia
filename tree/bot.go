package tree

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"codeberg.org/eviedelta/drc"
	"codeberg.org/eviedelta/drc/detc"
	"codeberg.org/eviedelta/drc/subpresets"
	"codeberg.org/eviedelta/trit"
	"github.com/bwmarrin/discordgo"
	"github.com/eviedelta/openjishia/config"
	"github.com/eviedelta/openjishia/module"
	"github.com/eviedelta/openjishia/wlog"
)

// Conf is the config
var Conf config.Config

// Dg is the main discord session
var Dg *discordgo.Session

// Hn is the global handler
var Hn *drc.Handler

func onReady(s *discordgo.Session, r *discordgo.Ready) {
	err := wlog.Info.Printf("onReady Event Received, Bot running and connected to Discord.\nAccount: %v (%v)", Dg.State.User.String(), Dg.State.User.ID)
	if err != nil {
		log.Println(err)
		err = wlog.Err.Printf("Error sending log message: %v", err)
		if err != nil {
			log.Println(Dg.UpdateListeningStatus("LOGGING ERROR, PLEASE REPORT TO ADMIN"))
		}
		log.Println(err)
	}
}

var adminCommands *drc.Command
var debugCommands *drc.Command

// Setup this is crappy, i'll rewrite this to be prettier eventually ~~or never apparently~~
func Setup(modules []*module.Module) {
	var err error

	// make a new bot session
	Dg, err = discordgo.New("Bot " + Conf.Auth.Token)
	if err != nil {
		log.Fatalln("Error creating discord session", err)
	}

	Dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)

	Dg.Client.Timeout = time.Minute

	// initialise handler session
	Hn = drc.NewHandler(drc.CfgHandler{
		Prefixes: Conf.Bot.Prefix,
		Admins:   Conf.Bot.Admins,

		ReplyOn: drc.ActOn{
			Error:   trit.True,
			Denied:  trit.True,
			Failure: trit.True,
		},
		ReactOn: drc.ActOn{
			Error:   trit.True,
			Denied:  trit.True,
			Failure: trit.True,
		},

		LogDebug: Conf.Bot.Debugm,

		DefaultSubcommands: []drc.Command{
			subpresets.SubcommandSimpleHelp,
		},
	}, Dg)

	// Initialise debug and admin commands
	adminCommands = Hn.Commands.Add(&drc.Command{
		Name: "admin",
		Permissions: drc.Permissions{
			BotAdmin: trit.True,
		},
		Exec: func(ctx *drc.Context) error {
			// this is crappy but its admin only anyways
			return ctx.Reply("```\n", detc.FormatterThing(fmt.Sprint(ctx.Com.Subcommands.ListRecursive(true, 2, false))), "\n```")
		},
	})
	debugCommands = Hn.Commands.Add(&drc.Command{
		Name: "debug",
		Permissions: drc.Permissions{
			BotAdmin: trit.True,
		},
		Exec: func(ctx *drc.Context) error {
			return ctx.Reply("```\n", detc.FormatterThing(fmt.Sprint(ctx.Com.Subcommands.ListRecursive(true, 2, false))), "\n```")
		},
	})
	Hn.Commands.Add(adminCommands)
	if Conf.Bot.DebugCommands {
		Hn.Commands.Add(debugCommands)
	}

	//	fmt.Println("debug commands", Conf.Bot.DebugCommands)

	// add the core handlers (guild create mostly just exists for debug at this point)
	Dg.AddHandler(onMessage)
	Dg.AddHandler(guildCreate)
	Dg.AddHandler(onReady)

	var modlist string

	// Load up all the provided modules
	for _, m := range modules {
		modlist += m.Name + "\n"
		if m.Config != nil {
			err := config.AnyConf(ConfigDir, m.Name+".toml", m.Config)
			if err == nil {
				m.ConfigFound = true
			} else if !errors.Is(err, os.ErrNotExist) {
				panic(err)
			}
		}

		if m.InitFunc != nil {
			err := m.InitFunc(m)
			if err != nil {
				panic(err)
			}
		}

		// Add all commands
		Hn.Commands.AddBulk(m.Commands)
		adminCommands.Subcommands.AddBulk(m.AdminCommands)

		// Add debug commands only if debug commands are enabled
		if Conf.Bot.DebugCommands {
			debugCommands.Subcommands.AddBulk(m.DebugCommands)
		}

		// if no extra discordgo handlers were provided skip and contine
		if len(m.DgoHandlers) < 1 {
			continue
		}
		// else add those to the discordgo session
		for _, h := range m.DgoHandlers {
			Dg.AddHandler(h)
		}
	}

	// Initialise the handler
	err = Hn.Ready()
	if err != nil {
		log.Fatalln("Error initialising handler", err)
	}

	err = wlog.Info.Printf("Setup bot with modules:```\v%v```\nRunning DRC: %v\nBot Software: %v v%v", modlist, drc.Version, BotSoftware, BotVersion)

	if err != nil {
		err2 := wlog.Err.Print("Error sending init log", err)
		if err != nil {
			log.Println(err2)
		}
		log.Fatalln(err)
	}
}

// StartUntilStop starts the bot and runs it until something requests it to close
func StartUntilStop(modules []*module.Module) {
	for _, m := range modules {
		if m.OpenFunc != nil {
			err := m.OpenFunc(m)
			if err != nil {
				panic(err)
			}
		}
		if m.CloseFunc != nil {
			defer m.CloseFunc(m)
		}
	}

	// Connect to discord
	err := Dg.Open()
	if err != nil {
		log.Fatalln("Error opening connection", err)
	}
	// Disconnect once everything is done
	defer Dg.Close()

	defer func() {
		log.Println(wlog.Info.Printf("Bot Shutting Down..."))
	}()

	// Wait until a shutdown signal is recived
	fmt.Println("Bot Running, Press CTRL-C to Shutdown")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

// guildCreate just logs all guilds added for debug purposes
func guildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	if Conf.Bot.Debugm {
		log.Println("guild create: ", g.Name, " (", g.ID, ")")
	}
}
