/*Package enpsql .

  Currently there is no automated schema management
  so schema updates will have to be applied manually for each module

  when making a module that makes use of this
  please make a dedicated sql schema for each module as to avoid conflicts between modules
  avoid using the public schema for anything in any modules

  the public schema should only be used internally for the psql package

*/
package enpsql

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"time"

	"codeberg.org/eviedelta/drc"
	"github.com/burntsushi/toml"
	"github.com/eviedelta/openjishia/module"
	"github.com/gocraft/dbr/v2"

	// Import postgresql for database stuff
	_ "github.com/lib/pq"
)

var dbc *dbr.Connection

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
	Database, Username, Password, Hostname string
}

// Module contains the module, i mean what else would it contain
var Module = &module.Module{
	Name:        "psql",
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

	OpenFunc:  OpenFunc,
	CloseFunc: CloseFunc,
}

// OpenFunc contains some the initialisation functions for this module per the terms of the pre-alpha attempt at modularisation this is built off of
func OpenFunc(mod *module.Module) error {
	dsn := fmt.Sprintf("dbname=%v user=%v password='%v' host='%v' sslmode=disable",
		Config.Database, Config.Username, Config.Password, Config.Hostname)
	conn, err := dbr.Open("postgres", dsn, nil)
	if err != nil {
		return err
	}
	if err := conn.Ping(); err != nil {
		return err
	}

	dbc = conn

	return nil
}

// CloseFunc contains the closing functions for this module per the terms of the pre-alpha attempt at modularisation this is built off of
func CloseFunc(mod *module.Module) {
	log.Println(dbc.Close())
}
