package highlights

import (
	"fmt"

	"github.com/eviedelta/openjishia/wlog"
	"github.com/lib/pq"
)

func guildSetEnabled(guildID string, state bool) error {
	_, err := db.s.Exec("insert into highlights.guilds (guild_id, enabled) values ($1, $2) on conflict (guild_id) do update set enabled = $2", guildID, state)
	if err != nil {
		wlog.Err.Printf("Toggling highlights for g:%v: %v", guildID, err)
	}
	return err
}

func guildIsEnabled(guildID string) (enabled bool) {
	err := db.s.QueryRow("select enabled from highlights.guilds where guild_id = $1", guildID).Scan(&enabled)
	if err != nil {
		wlog.Err.Printf("Checking highlights enabled for g:%v: %v", guildID, err)
	}

	return enabled
}

func guildBlockChannel(guildID string, target string) error {
	_, err := db.s.Exec("insert into highlights.guilds (guild_id, blocked_channels) values ($1, $2) on conflict (guild_id) do update set blocked_channels = array_cat(guilds.blocked_channels, $2)", guildID, pq.Array(&[]string{target}))
	if err != nil {
		wlog.Err.Printf("Updating highlight blocks for g:%v: %v", guildID, err)
	}
	return err
}

func guildUnblockChannel(guildID string, target string) error {
	_, err := db.s.Exec("insert into highlights.guilds (guild_id, blocked_channels) values ($1, array[]::text[]) on conflict (guild_id) do update set blocked_channels = array_remove(guilds.blocked_channels, $2)", guildID, target)
	if err != nil {
		wlog.Err.Printf("Updating highlight blocks for g:%v: %v", guildID, err)
	}
	return err
}

func guildBlockedChannels(guildID string) (blocks []string, err error) {
	err = db.s.QueryRow("insert into highlights.guilds (guild_id) values ($1) on conflict (guild_id) do update set guild_id = $1 returning blocked_channels", guildID).Scan(pq.Array(&blocks))
	if err != nil {
		wlog.Err.Printf("Getting highlight blocks for/g:%v: %v", guildID, err)
	}
	return blocks, err
}

func userSetEnabled(userID, guildID string, toggleTo bool) {
	_, err := db.s.Exec("update highlights.highlights set enabled = $3 where user_id = $1 and guild_id = $2", userID, guildID, toggleTo)
	if err != nil {
		fmt.Printf("Error toggling highlights for u:%v/g:%v: %v\n", userID, guildID, err)
	}
}

func userAddHighlight(userID, guildID string, term string) error {
	_, err := db.s.Exec("insert into highlights.highlights (user_id, guild_id, words) values ($1, $2, $3) on conflict (user_id, guild_id) do update set words = array_cat(highlights.words, $3)", userID, guildID, pq.Array(&[]string{term}))
	if err != nil {
		wlog.Err.Printf("Updating highlights for u:%v/g:%v: %v", userID, guildID, err)
	}
	return err
}

func userRemoveHighlight(userID, guildID string, term string) error {
	_, err := db.s.Exec("insert into highlights.highlights (user_id, guild_id, words) values ($1, $2, array[]::text[]) on conflict (user_id, guild_id) do update set words = array_remove(highlights.words, $3)", userID, guildID, term)
	if err != nil {
		wlog.Err.Printf("Updating highlights for u:%v/g:%v: %v", userID, guildID, err)
	}
	return err
}

func userClearHighlights(userID, guildID string) error {
	_, err := db.s.Exec("insert into highlights.highlights (user_id, guild_id, words) values ($1, $2, array[]::text[]) on conflict (user_id, guild_id) do update set words = array[]::text[]", userID, guildID)
	if err != nil {
		wlog.Err.Printf("Updating highlights for u:%v/g:%v: %v", userID, guildID, err)
	}
	return err
}

func userListHighlights(userID, guildID string) (highlights []string, err error) {
	err = db.s.QueryRow("insert into highlights.highlights (user_id, guild_id) values ($1, $2) on conflict (user_id, guild_id) do update set user_id = $1 returning words", userID, guildID).Scan(pq.Array(&highlights))
	if err != nil {
		wlog.Err.Printf("Getting highlights for u:%v/g:%v: %v", userID, guildID, err)
	}
	return highlights, err
}

func userBlockMember(userID, guildID string, target string) error {
	_, err := db.s.Exec("insert into highlights.highlights (user_id, guild_id, blocked_users) values ($1, $2, $3) on conflict (user_id, guild_id) do update set blocked_users = array_cat(highlights.blocked_users, $3)", userID, guildID, pq.Array(&[]string{target}))
	if err != nil {
		wlog.Err.Printf("Updating highlight blocks for u:%v/g:%v: %v", userID, guildID, err)
	}
	return err
}

func userUnblockMember(userID, guildID string, target string) error {
	//	fmt.Println("remove USER block", target, "for", userID, guildID)
	_, err := db.s.Exec("insert into highlights.highlights (user_id, guild_id, blocked_users) values ($1, $2, array[]::text[]) on conflict (user_id, guild_id) do update set blocked_users = array_remove(highlights.blocked_users, $3)", userID, guildID, target)
	if err != nil {
		wlog.Err.Printf("Updating highlight blocks for u:%v/g:%v: %v", userID, guildID, err)
	}
	return err
}

func userBlockedMembers(userID, guildID string) (blocks []string, err error) {
	err = db.s.QueryRow("insert into highlights.highlights (user_id, guild_id) values ($1, $2) on conflict (user_id, guild_id) do update set user_id = $1 returning blocked_users", userID, guildID).Scan(pq.Array(&blocks))
	if err != nil {
		wlog.Err.Printf("Getting highlight blocks for u:%v/g:%v: %v", userID, guildID, err)
	}
	return blocks, err
}

func userBlockChannel(userID, guildID string, target string) error {
	_, err := db.s.Exec("insert into highlights.highlights (user_id, guild_id, blocked_channels) values ($1, $2, $3) on conflict (user_id, guild_id) do update set blocked_channels = array_cat(highlights.blocked_channels, $3)", userID, guildID, pq.Array(&[]string{target}))
	if err != nil {
		wlog.Err.Printf("Updating highlight blocks for u:%v/g:%v: %v", userID, guildID, err)
	}
	return err
}

func userUnblockChannel(userID, guildID string, target string) error {
	fmt.Println("remove CHANNEL block", target, "for", userID, guildID)
	_, err := db.s.Exec("insert into highlights.highlights (user_id, guild_id, blocked_channels) values ($1, $2, array[]::text[]) on conflict (user_id, guild_id) do update set blocked_channels = array_remove(highlights.blocked_channels, $3)", userID, guildID, target)
	if err != nil {
		wlog.Err.Printf("Updating highlight blocks for u:%v/g:%v: %v", userID, guildID, err)
	}
	return err
}

func userBlockedChannels(userID, guildID string) (blocks []string, err error) {
	err = db.s.QueryRow("insert into highlights.highlights (user_id, guild_id) values ($1, $2) on conflict (user_id, guild_id) do update set user_id = $1 returning blocked_channels", userID, guildID).Scan(pq.Array(&blocks))
	if err != nil {
		wlog.Err.Printf("Getting highlight blocks for u:%v/g:%v: %v", userID, guildID, err)
	}
	return blocks, err
}
