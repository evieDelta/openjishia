package highlights

import (
	"github.com/eviedelta/openjishia/wlog"
	"github.com/lib/pq"
)

func setguildenable(guildID string, state bool) {
	_, err := db.s.Exec("insert into highlights.guilds (guild_id, enabled) values ($1, $2) on conflict (guild_id) do update set enabled = $2", guildID, state)
	if err != nil {
		wlog.Err.Printf("Toggling highlights for g:%v: %v", guildID, err)
	}
}

func isguildenabled(guildID string) (enabled bool) {
	err := db.s.QueryRow("select enabled from highlights.guilds where guild_id = $1", guildID).Scan(&enabled)
	if err != nil {
		wlog.Err.Printf("Checking highlights enabled for g:%v: %v", guildID, err)
	}

	return enabled
}

func useraddhl(userID, guildID string, term string) {
	_, err := db.s.Exec("insert into highlights.highlights (user_id, guild_id, words) values ($1, $2, $3) on conflict (user_id, guild_id) do update set words = array_cat(highlights.words, $3)", userID, guildID, pq.Array(&[]string{term}))
	if err != nil {
		wlog.Err.Printf("Updating highlights for u:%v/g:%v: %v", userID, guildID, err)
	}
}

func userremhl(userID, guildID string, term string) {
	_, err := db.s.Exec("insert into highlights.highlights (user_id, guild_id, words) values ($1, $2, array[]::text[]) on conflict (user_id, guild_id) do update set words = array_remove(highlights.words, $3)", userID, guildID, term)
	if err != nil {
		wlog.Err.Printf("Updating highlights for u:%v/g:%v: %v", userID, guildID, err)
	}
}

func userlshl(userID, guildID string) (highlights []string) {
	err := db.s.QueryRow("insert into highlights.highlights (user_id, guild_id) values ($1, $2) on conflict (user_id, guild_id) do update set user_id = $1 returning words", userID, guildID).Scan(pq.Array(&highlights))
	if err != nil {
		wlog.Err.Printf("Getting highlights for u:%v/g:%v: %v", userID, guildID, err)
	}
	return highlights
}

func useraddhlblock(userID, guildID string, term string) {
	_, err := db.s.Exec("insert into highlights.highlights (user_id, guild_id, blocks) values ($1, $2, $3) on conflict (user_id, guild_id) do update set blocks = array_cat(highlights.blocks, $3)", userID, guildID, pq.Array(&[]string{term}))
	if err != nil {
		wlog.Err.Printf("Updating highlight blocks for u:%v/g:%v: %v", userID, guildID, err)
	}
}

func userremhlblock(userID, guildID string, term string) {
	_, err := db.s.Exec("insert into highlights.highlights (user_id, guild_id, blocks) values ($1, $2, array[]::text[]) on conflict (user_id, guild_id) do update set blocks = array_remove(highlights.blocks, $3)", userID, guildID, term)
	if err != nil {
		wlog.Err.Printf("Updating highlight blocks for u:%v/g:%v: %v", userID, guildID, err)
	}
}

func userlshlblock(userID, guildID string) (blocks []string) {
	err := db.s.QueryRow("insert into highlights.highlights (user_id, guild_id) values ($1, $2) on conflict (user_id, guild_id) do update set user_id = $1 returning blocks", userID, guildID).Scan(pq.Array(&blocks))
	if err != nil {
		wlog.Err.Printf("Getting highlight blocks for u:%v/g:%v: %v", userID, guildID, err)
	}
	return blocks
}
