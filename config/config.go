package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml"
)

// Config contains the core configuration for the bot, i mean what else would it contain
type Config struct {
	Auth    Auth
	Bot     Bot
	Logging Logging
}

// Auth contains authentication configuration
type Auth struct {
	Token string
}

// Bot contains some general purpose configuration
type Bot struct {
	Debugm bool
	Prefix []string
	Admins []string

	DebugCommands bool `toml:"EnableDebugCommands"`
}

// Logging contains some logging configs
type Logging struct {
	Webhooks struct {
		Info string
		Errs string
		Spam string
	}
	InfoPfp    string
	ErrsPfp    string
	SpamPfp    string
	NamePrefix string
}

// AConf is a function dedicated to checking various locations for a config, and loading that config
func AConf(loc string) (Config, error) {
	var search []string
	if loc == "" {
		search = []string{
			"./config.toml", "./data/config.toml", "./config/config.toml"}
	} else {
		search = []string{
			filepath.Join(loc, "config.toml"),
			"./config.toml", "./data/config.toml", "./config/config.toml"}
	}

	for _, f := range search {
		if !CheckExist(f) {
			continue
		}
		cfg, err := Load(f)
		return cfg, err
	}

	fmt.Println("No configuration found. locations checked: ", search)

	os.Exit(0)
	return Config{}, nil
}

// Load loads a toml based config file
func Load(loc string) (cfg Config, err error) {
	dat, err := os.ReadFile(loc)
	if err != nil {
		return cfg, err
	}

	err = toml.Unmarshal(dat, &cfg)
	return cfg, err
}

// CheckExist is a lazy function to determine if a file exists or not
func CheckExist(loc string) bool {
	i, err := os.Stat(loc)
	if err != nil || i.IsDir() {
		return false
	}
	return true
}

// CheckDirExist is a lazy function to determine if a directory exists or not
func CheckDirExist(loc string) bool {
	i, err := os.Stat(loc)
	if err != nil || !i.IsDir() {
		return false
	}
	return true
}

// AnyConf is like AConf but it accepts any struct and will attempt to unmarshal the config file in the location defined by prefix into it
func AnyConf(loc, filename string, cfg interface{}) error {
	var search []string
	if loc == "" {
		search = []string{"./", "./data/", "./config/"}
	} else {
		if !strings.HasSuffix(loc, "/") {
			loc += "/"
		}
		search = []string{loc, "./", "./data/", "./conifg/"}
	}

	for _, f := range search {
		if !CheckExist(f + filename) {
			continue
		}

		dat, err := os.ReadFile(f + filename)
		if err != nil {
			return err
		}

		err = toml.Unmarshal(dat, cfg)
		return err
	}
	return fmt.Errorf("%w, locations checked for `%v`: %v", os.ErrNotExist, filename, search)
}
