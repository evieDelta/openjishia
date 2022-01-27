package highlights

import (
	"embed"

	"github.com/eviedelta/openjishia/enpsql"
	"github.com/eviedelta/openjishia/wlog"
	"github.com/gocraft/dbr/v2"
	"github.com/lib/pq"
)

//go:embed schemas/*
var schemas embed.FS

func init() {
	_ = enpsql.RegisterSchemaFS("highlights", schemas, "schemas")
}

var db = &Database{}

type Database struct {
	s *dbr.Session
}

type Highlight struct {
	UserID        string
	Enabled       bool
	Highlights    []string
	ChannelBlocks []string
	UserBlocks    []string
}

func guildGetAllHighlights(guildID string) (hls []Highlight) {
	rows, err := db.s.Query("select user_id, enabled, words, blocked_channels, blocked_users from highlights.highlights where guild_id = $1", guildID)
	if err != nil {
		wlog.Err.Printf("Getting highlights for g:%v: %v", guildID, err)

		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var hl Highlight
		if err := rows.Scan(&hl.UserID, &hl.Enabled, pq.Array(&hl.Highlights), pq.Array(&hl.ChannelBlocks), pq.Array(&hl.UserBlocks)); err != nil {
			wlog.Err.Printf("Getting highlights for g:%v: %v", guildID, err)
		}
		hls = append(hls, hl)
	}

	if err := rows.Err(); err != nil {
		wlog.Err.Printf("Getting highlights for g:%v: %v", guildID, err)
	}
	return hls
}
