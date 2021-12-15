package highlights

import (
	"embed"
	"fmt"

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

func toggleForUser(userID, guildID string, toggleTo bool) {
	_, err := db.s.Exec("update highlights.highlights set enabled = $3 where user_id = $1 and guild_id = $2", userID, guildID, toggleTo)
	if err != nil {
		fmt.Printf("Error toggling highlights for u:%v/g:%v: %v\n", userID, guildID, err)
	}
}

type Highlight struct {
	UserID     string
	Enabled    bool
	Highlights []string
	Blocks     []string
}

func guildHighlightsForHighlighter(guildID string) (hls []Highlight) {
	rows, err := db.s.Query("select user_id, enabled, words, blocks from highlights.highlights where guild_id = $1", guildID)
	if err != nil {
		wlog.Err.Printf("Getting highlights for g:%v: %v", guildID, err)

		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var hl Highlight
		if err := rows.Scan(&hl.UserID, &hl.Enabled, pq.Array(&hl.Highlights), pq.Array(&hl.Blocks)); err != nil {
			wlog.Err.Printf("Getting highlights for g:%v: %v", guildID, err)
		}
		hls = append(hls, hl)
	}

	if err := rows.Err(); err != nil {
		wlog.Err.Printf("Getting highlights for g:%v: %v", guildID, err)
	}
	return hls
}
