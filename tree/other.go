package tree

import (
	"flag"
	"log"
	"os"

	"codeberg.org/eviedelta/dwhook"
	"github.com/eviedelta/openjishia/config"
	"github.com/eviedelta/openjishia/wlog"
)

//
var (
	BotVersion  = "0.1.12.0-whyDoIHaveThisIliterallyNeverUpdateIt-Edition" // fun fact, this has been 0.1.12 since January 2021
	BotSoftware = "openSpaghettishia"
)

//
var (
	ConfigDir string
	DataDir   string
)

// InitHandleFlags is the same as Initialise but it loads its options from flags
func InitHandleFlags() {
	// flags variables for some global configuration
	var (
		flagConfig  string
		flagDataDir string
	)

	// set the config file location (it has some defaults to check so you don't need to specify with the flag if its in ./config.toml or ./data/config.toml)
	flag.StringVar(&flagConfig, "config", "", "the directory to look in for config files")
	flag.StringVar(&flagDataDir, "data", "", "the location to look for data in")
	flag.Parse()

	Initialise(flagDataDir, flagConfig)
}

// Initialise loads the config data from the config file and sets the current directory to datadir for loading and saving data
func Initialise(datadir, configdir string) {
	if datadir != "" {
		err := os.Chdir(datadir)
		if err != nil {
			panic(err)
		}
	}

	ConfigDir = configdir
	DataDir = datadir

	// load le config
	cfg, err := config.AConf(configdir)
	if err != nil {
		log.Fatalln("Configuration error", err)
	}
	Conf = cfg

	id, token, err := dwhook.ParseWebhookURL(Conf.Logging.Webhooks.Errs)
	if err != nil {
		log.Fatalln("err wh", err)
	}
	wlog.Err.Webhook.ID = id
	wlog.Err.Webhook.Token = token
	wlog.Err.Name = Conf.Logging.NamePrefix + ": Errors"
	wlog.Err.Avatar = Conf.Logging.ErrsPfp

	id, token, err = dwhook.ParseWebhookURL(Conf.Logging.Webhooks.Info)
	if err != nil {
		log.Fatalln("info wh", err)
	}
	wlog.Info.Webhook.ID = id
	wlog.Info.Webhook.Token = token
	wlog.Info.Name = Conf.Logging.NamePrefix + ": Info"
	wlog.Info.Avatar = Conf.Logging.InfoPfp

	id, token, err = dwhook.ParseWebhookURL(Conf.Logging.Webhooks.Spam)
	if err != nil {
		log.Println("spam wh", err)
	}
	wlog.Spam.Webhook.ID = id
	wlog.Spam.Webhook.Token = token
	wlog.Spam.Name = Conf.Logging.NamePrefix + ": Spam"
	wlog.Spam.Avatar = Conf.Logging.SpamPfp
}
