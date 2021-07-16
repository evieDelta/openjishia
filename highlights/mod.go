package highlights

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"codeberg.org/eviedelta/drc"
	"github.com/bwmarrin/discordgo"
	"github.com/eviedelta/openjishia/module"
	"github.com/eviedelta/openjishia/wlog"
)

var Config = struct {
	Debug bool
}{}

// Module はエンコのmodule
var Module = &module.Module{
	Name: "highlights",

	Config: &Config,

	DgoHandlers: []interface{}{
		guildCreate,
		messageCreate,
		onReady,
		onMemberCreate,
		onMemberLeave,
	},

	Commands: []*drc.Command{
		hlconf,
		hladmin,
		hlmod,
		cmdhighlight,
		{Name: "hl", AliasText: []string{"highlight"}},
	},

	InitFunc: func(*module.Module) error {
		globallock.Lock()
		defer globallock.Unlock()

		err := globalruntime.Load(filename)
		if err != nil {
			return err
		}

		if globalruntime.Guildsettings == nil {
			globalruntime.Guildsettings = make(map[string]*guildsettings)
		}
		if globalruntime.DeletionQueue == nil {
			globalruntime.DeletionQueue = &deletionqueue{}
		}
		if globalruntime.DeletionQueue.Entries == nil {
			globalruntime.DeletionQueue.Entries = make(map[string]*deletionentry)
		}

		//		for _, x := range globalruntime.Guildsettings {
		//		}

		return nil
	},
	CloseFunc: func(*module.Module) {
		err := globalruntime.Save(filename)
		if err != nil {
			wlog.Err.Print(err, "\n\n>>> ```\n", string(debug.Stack()), "\n```")
		}
	},
}

func guildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	globallock.Lock()
	defer globallock.Unlock()

	if globalruntime.Guildsettings[g.ID] == nil {
		globalruntime.Guildsettings[g.ID] = &guildsettings{}
	}
	if globalruntime.Guildsettings[g.ID].MemberSettings == nil {
		globalruntime.Guildsettings[g.ID].MemberSettings = make(map[string]*membersettings, g.MemberCount)
	}
}

func onReady(s *discordgo.Session, r *discordgo.Ready) {
	globallock.Lock()
	defer globallock.Unlock()

	time.Sleep(time.Second * 5)

	globalruntime.DeletionQueue.close = true

	globalruntime.DeletionQueue.otherlock.Lock()

	go globalruntime.DeletionQueue.DeletionScheduler(s)

	globalruntime.DeletionQueue.otherlock.Unlock()
}

func onMemberCreate(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	globallock.Lock()
	defer globallock.Unlock()

	if globalruntime.Guildsettings[m.GuildID].MemberSettings[m.User.ID] == nil {
		return
	}

	globalruntime.Guildsettings[m.GuildID].MemberSettings[m.User.ID].Disabled = false
}

func onMemberLeave(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	globallock.Lock()
	defer globallock.Unlock()

	if globalruntime.Guildsettings[m.GuildID].MemberSettings[m.User.ID] != nil {
		globalruntime.Guildsettings[m.GuildID].MemberSettings[m.User.ID].Disabled = true
	}
}

const filename = "highlights.json"

var globallock = sync.RWMutex{}
var globalruntime runtimecrap

type highlight struct{}

type membersettings struct {
	Highlightwords map[string]*highlight

	Blocks map[string]blockerstuff

	Disabled bool
}

const (
	blockTypeChannel = iota
	blockTypeUser
)

type blockerstuff struct {
	State bool
	Thing int
}

type guildsettings struct {
	Enable bool

	MemberSettings map[string]*membersettings
}

type runtimecrap struct {
	Guildsettings map[string]*guildsettings
	DeletionQueue *deletionqueue
}

func (r *runtimecrap) Load(file string) error {
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	if r.Guildsettings == nil {
		r.Guildsettings = make(map[string]*guildsettings)
	}

	err = json.Unmarshal(dat, &r)
	if err != nil {
		return err
	}

	return nil
}

func (r *runtimecrap) Save(file string) error {
	dat, err := json.MarshalIndent(*r, "", "	")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, dat, 0640)
}
