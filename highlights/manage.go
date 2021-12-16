package highlights

import (
	"fmt"

	"github.com/eviedelta/openjishia/wlog"
	"github.com/lib/pq"
)

func setguildenable(guildID string, state bool) error {
	_, err := db.s.Exec("insert into highlights.guilds (guild_id, enabled) values ($1, $2) on conflict (guild_id) do update set enabled = $2", guildID, state)
	if err != nil {
		wlog.Err.Printf("Toggling highlights for g:%v: %v", guildID, err)
	}
	return err
}

func isguildenabled(guildID string) (enabled bool) {
	err := db.s.QueryRow("select enabled from highlights.guilds where guild_id = $1", guildID).Scan(&enabled)
	if err != nil {
		wlog.Err.Printf("Checking highlights enabled for g:%v: %v", guildID, err)
	}

	return enabled
}

func useraddhl(userID, guildID string, term string) error {
	_, err := db.s.Exec("insert into highlights.highlights (user_id, guild_id, words) values ($1, $2, $3) on conflict (user_id, guild_id) do update set words = array_cat(highlights.words, $3)", userID, guildID, pq.Array(&[]string{term}))
	if err != nil {
		wlog.Err.Printf("Updating highlights for u:%v/g:%v: %v", userID, guildID, err)
	}
	return err
}

func userremhl(userID, guildID string, term string) error {
	_, err := db.s.Exec("insert into highlights.highlights (user_id, guild_id, words) values ($1, $2, array[]::text[]) on conflict (user_id, guild_id) do update set words = array_remove(highlights.words, $3)", userID, guildID, term)
	if err != nil {
		wlog.Err.Printf("Updating highlights for u:%v/g:%v: %v", userID, guildID, err)
	}
	return err
}

func userHlClear(userID, guildID string) error {
	_, err := db.s.Exec("insert into highlights.highlights (user_id, guild_id, words) values ($1, $2, array[]::text[]) on conflict (user_id, guild_id) do update set words = array[]::text[]", userID, guildID)
	if err != nil {
		wlog.Err.Printf("Updating highlights for u:%v/g:%v: %v", userID, guildID, err)
	}
	return err
}

func userlshl(userID, guildID string) (highlights []string, err error) {
	err = db.s.QueryRow("insert into highlights.highlights (user_id, guild_id) values ($1, $2) on conflict (user_id, guild_id) do update set user_id = $1 returning words", userID, guildID).Scan(pq.Array(&highlights))
	if err != nil {
		wlog.Err.Printf("Getting highlights for u:%v/g:%v: %v", userID, guildID, err)
	}
	return highlights, err
}

func userAddUserBlock(userID, guildID string, target string) error {
	_, err := db.s.Exec("insert into highlights.highlights (user_id, guild_id, user_blocks) values ($1, $2, $3) on conflict (user_id, guild_id) do update set user_blocks = array_cat(highlights.user_blocks, $3)", userID, guildID, pq.Array(&[]string{target}))
	if err != nil {
		wlog.Err.Printf("Updating highlight blocks for u:%v/g:%v: %v", userID, guildID, err)
	}
	return err
}

func userRemUserBlock(userID, guildID string, target string) error {
	fmt.Println("remove USER block", target, "for", userID, guildID)
	_, err := db.s.Exec("insert into highlights.highlights (user_id, guild_id, user_blocks) values ($1, $2, array[]::text[]) on conflict (user_id, guild_id) do update set user_blocks = array_remove(highlights.user_blocks, $3)", userID, guildID, target)
	if err != nil {
		wlog.Err.Printf("Updating highlight blocks for u:%v/g:%v: %v", userID, guildID, err)
	}
	return err
}

func userListUserBlocks(userID, guildID string) (blocks []string, err error) {
	err = db.s.QueryRow("insert into highlights.highlights (user_id, guild_id) values ($1, $2) on conflict (user_id, guild_id) do update set user_id = $1 returning user_blocks", userID, guildID).Scan(pq.Array(&blocks))
	if err != nil {
		wlog.Err.Printf("Getting highlight blocks for u:%v/g:%v: %v", userID, guildID, err)
	}
	return blocks, err
}

func userAddChannelBlock(userID, guildID string, target string) error {
	_, err := db.s.Exec("insert into highlights.highlights (user_id, guild_id, channel_blocks) values ($1, $2, $3) on conflict (user_id, guild_id) do update set channel_blocks = array_cat(highlights.channel_blocks, $3)", userID, guildID, pq.Array(&[]string{target}))
	if err != nil {
		wlog.Err.Printf("Updating highlight blocks for u:%v/g:%v: %v", userID, guildID, err)
	}
	return err
}

func userRemChannelBlock(userID, guildID string, target string) error {
	fmt.Println("remove CHANNEL block", target, "for", userID, guildID)
	_, err := db.s.Exec("insert into highlights.highlights (user_id, guild_id, channel_blocks) values ($1, $2, array[]::text[]) on conflict (user_id, guild_id) do update set channel_blocks = array_remove(highlights.channel_blocks, $3)", userID, guildID, target)
	if err != nil {
		wlog.Err.Printf("Updating highlight blocks for u:%v/g:%v: %v", userID, guildID, err)
	}
	return err
}

func userListChannelBlocks(userID, guildID string) (blocks []string, err error) {
	err = db.s.QueryRow("insert into highlights.highlights (user_id, guild_id) values ($1, $2) on conflict (user_id, guild_id) do update set user_id = $1 returning channel_blocks", userID, guildID).Scan(pq.Array(&blocks))
	if err != nil {
		wlog.Err.Printf("Getting highlight blocks for u:%v/g:%v: %v", userID, guildID, err)
	}
	return blocks, err
}
