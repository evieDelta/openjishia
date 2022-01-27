package highlights

import (
	"embed"
	"fmt"

	"codeberg.org/eviedelta/drc"
	"github.com/bwmarrin/discordgo"
	"github.com/eviedelta/openjishia/enpsql"
	"github.com/eviedelta/openjishia/helputil"
	"github.com/eviedelta/openjishia/module"
)

var config = struct {
	Debug bool
}{}

// Module はエンコのmodule
var Module = &module.Module{
	Name: "highlights",

	Config: &config,

	DgoHandlers: []interface{}{
		onGuildCreate,
		messageCreate,
		onMemberCreate,
		onMemberLeave,
	},

	Commands: []*drc.Command{
		cConf,
		cAdmin,
		cMod,
		cHighlight,
		{Name: "hl", AliasText: []string{"highlight"}},
	},

	InitFunc: func(*module.Module) error {
		return nil
	},
	OpenFunc: func(m *module.Module) error {
		db.s = enpsql.GetSession()
		return nil
	},
	CloseFunc: func(*module.Module) {
	},
}

func onGuildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	_, err := db.s.Exec("insert into highlights.guilds (guild_id) values ($1) on conflict (guild_id) do nothing", g.ID)
	if err != nil {
		fmt.Printf("Error adding guild %v to highlight config: %v\n", g.ID, err)
	}
}

func onMemberCreate(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	userSetEnabled(m.User.ID, m.GuildID, true)
}

func onMemberLeave(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	userSetEnabled(m.User.ID, m.GuildID, false)
}

//go:embed topics/*
var docs embed.FS

func init() {
	helputil.MustAddFS(docs)
}
