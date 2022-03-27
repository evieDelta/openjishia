/*Package enpsql .

  Currently there is no automated schema management
  so schema updates will have to be applied manually for each module

  when making a module that makes use of this
  please make a dedicated sql schema for each module as to avoid conflicts between modules
  avoid using the public schema for anything in any modules

  THIS PACKAGE DOES NOT ENFORCE SQL SCHEMA BOUNDRIES
  you will need to ensure you manually specify the schema by namespacing table names

  the public schema should only be used internally for this psql package

*/
package enpsql

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"codeberg.org/eviedelta/drc"
	"github.com/bwmarrin/discordgo"
	"github.com/eviedelta/openjishia/module"
	"github.com/eviedelta/openjishia/wlog"
	"github.com/gocraft/dbr/v2"
	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"

	// Import postgresql for database stuff
	_ "github.com/lib/pq"
)

var table = struct {
	Schemas string
	//	BotConfig string // possibile future feature, managed config
}{
	Schemas: "public.schemas",
	//	BotConfig: "public.bot_config",
}

var dbc *dbr.Connection

var ses *dbr.Session

func GetSession() *dbr.Session {
	return dbc.NewSession(nil)
}

func Ping() (time.Duration, error) {
	if dbc == nil {
		return 0, errors.New("database isn't connected yet")
	}

	t := time.Now()
	if err := dbc.Ping(); err != nil {
		return 0, err
	}

	return time.Now().Sub(t), nil
}

// Config is the config that contains global config stuff, i mean it is named config so what else would it be
var Config = &localConfig{}

type localConfig struct {
	Auth struct {
		Database, Username, Password, Hostname string
	}
	External struct {
		BackupCommand string // it will call this command (exactly as given) before doing a schema update
	}
}

// Module contains the module, i mean what else would it contain
// being this is the database its probably a good idea to put this first in the module list
var Module = &module.Module{
	Name:        "enpsql",
	DgoHandlers: []interface{}{},

	Commands: []*drc.Command{},

	Config: Config,

	InitFunc: func(mod *module.Module) error {
		if !mod.ConfigFound {
			b := bytes.NewBuffer([]byte{})
			e := toml.NewEncoder(b)
			err := e.Encode(mod.Config)
			if err != nil {
				return err
			}

			fmt.Println(b.String())
			return fmt.Errorf("Config not found, database config data mandatory")
		}
		return nil
	},

	OpenFunc:  openFunc,
	CloseFunc: closeFunc,
}

func onReady(s *discordgo.Session, e *discordgo.Ready) {
	err := dbc.Ping()
	if err != nil {
		wlog.Err.Print(err)
	}
}

// OpenFunc contains some the initialisation functions for this module per the terms of the pre-alpha attempt at modularisation this is built off of
func openFunc(mod *module.Module) error {
	dsn := fmt.Sprintf("dbname=%v user=%v password='%v' host='%v' sslmode=disable",
		Config.Auth.Database, Config.Auth.Username, Config.Auth.Password, Config.Auth.Hostname)
	conn, err := dbr.Open("postgres", dsn, nil)
	if err != nil {
		return errors.Wrap(err, "DB Connect")
	}
	if err := conn.Ping(); err != nil {
		return errors.Wrap(err, "DB Ping")
	}

	dbc = conn

	ses = GetSession()

	err = initialCheck()
	if err != nil {
		return errors.Wrap(err, "DB schema core")
	}
	err = updateAll()
	if err != nil {
		return errors.Wrap(err, "DB update schemas")
	}

	return nil
}

// CloseFunc contains the closing functions for this module per the terms of the pre-alpha attempt at modularisation this is built off of
func closeFunc(mod *module.Module) {
	log.Println(dbc.Close())
}
