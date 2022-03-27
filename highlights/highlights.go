package highlights

import (
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/eviedelta/openjishia/wlog"
	"github.com/pkg/errors"
)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	defer func() {
		err := recover()
		if err != nil {
			wlog.Err.Print(err, "\n\n>>> ```\n", string(debug.Stack()), "\n```")
		}
	}()

	err := highlighter(s, m)
	if err != nil {
		wlog.Err.Print(errors.Wrap(err, "@highlighter"))
	}
}

func highlighter(s *discordgo.Session, m *discordgo.MessageCreate) (err error) {
	if m.GuildID == "" || m.WebhookID != "" {
		return
	}

	// there are cases where you may want to quietly ensure a message doesn't highlight
	// so we allow people to do that with a _ _ suffix which is entirely transparent
	if strings.HasSuffix(m.Content, "_ _") {
		return
	}

	if !guildIsEnabled(m.GuildID) {
		return
	}

	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		return err
	}

	blocks, err := guildBlockedChannels(m.GuildID)
	if err != nil {
		return err
	}

	for _, id := range blocks {
		if id == m.ChannelID || id == channel.ParentID {
			//fmt.Println(m.ID, " | user ", hls.UserID, " channel blocked this message on a guild level")
			return
		}
	}

	addLimit(getrlimkey(m.Author.ID, m.ChannelID, ""), delaySelf)

	//	fmt.Println("processing message: ", m.ID)

	content := NormaliseString(m.Content)

mainloop:
	for _, hls := range guildGetAllHighlights(m.GuildID) {
		//fmt.Println(m.ID, " | checking highlights for user ", hls.UserID)
		if isLimited(getrlimkey(hls.UserID, m.ChannelID, "")) || !hls.Enabled {
			//fmt.Println(m.ID, " | user ", hls.UserID, " is currently on cooldown")
			continue
		}
		if p, err := s.State.UserChannelPermissions(hls.UserID, m.ChannelID); err != nil {
			userSetEnabled(hls.UserID, m.GuildID, false)
			wlog.Err.Print(errors.Wrapf(err, "@s.State.UserChannelPermissions (Check Permission for user %v)", hls.UserID))

			continue
		} else if p&discordgo.PermissionViewChannel == 0 {
			//fmt.Println(m.ID, " | user ", hls.UserID, " does not have permissions to view channel")
			continue
		}

		for _, id := range hls.ChannelBlocks {
			if id == m.ChannelID || id == channel.ParentID {
				//fmt.Println(m.ID, " | user ", hls.UserID, " channel blocked this message")
				continue mainloop
			}
		}
		for _, id := range hls.UserBlocks {
			if id == m.Author.ID {
				//fmt.Println(m.ID, " | user ", hls.UserID, " channel blocked this message")
				continue mainloop
			}
		}

		err := doUserHighlight(s, m, content, hls.UserID, hls.Highlights)
		if err != nil {
			wlog.Err.Print(errors.Wrapf(err, "@doUserHighlight (For user %v)", hls.UserID))
		}
	}
	return nil
}

func doUserHighlight(s *discordgo.Session, m *discordgo.MessageCreate, message, user string, highlights []string) (err error) {
	if len(highlights) < 1 {
		// fmt.Println(m.ID, "| user", user, "| user has no settings or no highlights")
		return
	}

	for _, word := range highlights {
		// fmt.Println(m.ID, "| user", user, "| proccessing word", y)
		if isLimited(getrlimkey(user, m.ChannelID, word)) {
			// fmt.Println(m.ID, "| user", user, "| word", y, "is currently on cooldown")
			continue
		}

		if start, end := checkHighlight(message, word, user, m); start >= 0 && end >= 0 {
			// fmt.Println(m.ID, "| user", user, "| word", y, "everything is clear, marking new cooldowns and moving to send alert")
			addLimit(getrlimkey(user, m.ChannelID, ""), delayAny)
			addLimit(getrlimkey(user, m.ChannelID, word), delaySpecific)

			err = sendHighlight(s, m, user, word, start, end)
			return errors.Wrap(err, "@sendHighlight")
		}
	}
	return
}

func sendHighlight(s *discordgo.Session, m *discordgo.MessageCreate, targetuser, term string, start, end int) error {
	uch, err := s.UserChannelCreate(targetuser)
	if err != nil {
		return errors.Wrap(err, "@UserChannelCreate (Get DM channel)")
	}

	msgs, err := s.ChannelMessages(m.ChannelID, 4, m.ID, "", "")
	if err != nil {
		return errors.Wrap(err, "@s.ChannelMessages (Get msgs)")
	}

	fields := make([]*discordgo.MessageEmbedField, 0, 5)

	for z := len(msgs) - 1; z >= 0; z-- {
		me, err := formatMessage(s, m.GuildID, msgs[z], 0, 0)
		if err != nil {
			return err
		}

		fields = append(fields, me)
	}

	{
		me, err := formatMessage(s, m.GuildID, m.Message, start, end)
		if err != nil {
			return err
		}

		fields = append(fields, me)
	}

	//	fmt.Println(len(fields), fields)

	t, err := discordgo.SnowflakeTimestamp(m.ID)
	if err != nil {
		return errors.Wrap(err, "@discordgo.SnowflakeTimestamp")
	}

	g, err := s.State.Guild(m.GuildID)
	if err != nil {
		return errors.Wrap(err, "@s.State.Guild (Get guild)")
	}

	msg, err := s.ChannelMessageSendComplex(uch.ID, &discordgo.MessageSend{
		// yes the \u200B: is necessary, don't remove it
		Content: fmt.Sprintf("You were mentioned in **%v**; <#%v>, with highlight word \"__%v__\"\u200B:", g.Name, m.ChannelID, term),
		//		Content: "You have been highlighted in **" + g.Name + "** <#" + m.ChannelID + "> with highlight word __" + term + "__\u200B:",

		Embed: &discordgo.MessageEmbed{
			Title:       strings.Title(term),
			Description: "**[Click here to Jump!](https://discordapp.com/channels/" + m.GuildID + "/" + m.ChannelID + "/" + m.ID + ")**",
			Fields:      fields,
			Color:       s.State.UserColor(m.Author.ID, m.ChannelID),

			Timestamp: t.Format(time.RFC3339),

			Footer: &discordgo.MessageEmbedFooter{
				Text: func() string {
					if m.Member != nil && m.Member.Nick != "" {
						return m.Member.Nick
					}
					return m.Author.Username
				}(),
				IconURL: m.Author.AvatarURL("128"),
			},
		},
		AllowedMentions: &discordgo.MessageAllowedMentions{},
	})

	if err != nil {
		return errors.Wrap(err, "@s.ChannelMessageSendComplex (Send Highlight)")
	}

	addDeletionQueue(msg.ID, msg.ChannelID, messageKeepTime)

	return err
}

func formatMessage(s *discordgo.Session, guild string, msg *discordgo.Message, start, end int) (*discordgo.MessageEmbedField, error) {
	ct := msg.Content

	if start == 0 && end == 0 {
		if len(msg.Content) > (previewSize * 1.1) {
			ct = trimToNearestBorder(ct, 0, previewSize)
		}
	} else {
		start, end := findPreviewArea(start, end, previewSize, len(ct))
		ct = trimToNearestBorder(ct, start, end)
	}

	author := msg.Author.Username
	mem, err := s.State.Member(guild, msg.Author.ID)
	if err == nil && mem.Nick != "" {
		author = mem.Nick
	}

	t, err := discordgo.SnowflakeTimestamp(msg.ID)
	if err != nil {
		return nil, errors.Wrap(err, "@dgo.SnowflakeTimestamp (Parse msg time)")
	}

	return &discordgo.MessageEmbedField{
		Name:  fmt.Sprintf("``[%v]`` %v", t.UTC().Format("15:04:05"), author),
		Value: "\u200B" + ct,
	}, nil
}

func findPreviewArea(start, end int, size int, length int) (_start int, _end int) {
	// find the center
	anchor := start + ((end - start) / 2)
	start = anchor - (size / 2)
	end = anchor + (size / 2)

	// if start is under 0 we clamp it back and push it to the end
	if start < 0 {
		end -= start // since start is negative it will add how much under 0 it is
		start = 0
	}
	// if the end is longer than the message content, we clamp it to the end
	if end > length {
		// but if we can fit it, we move the start back
		diff := (end - length)
		if start > 0 && start-diff >= 0 {
			start -= diff
		}
		end = length
	}

	return start, end
}

func trimToNearestBorder(s string, start, end int) string {
	rs := []rune(s)

	start = nearestBorder(rs, start, 10, 0)
	end = nearestBorder(rs, end, 10, 0)

	res := strings.TrimFunc(string(rs[start:end]), isBorder)

	switch {
	case start > 0 && end < len(rs):
		return "... " + res + " ..."
	case start > 0:
		return "... " + res
	case end < len(rs):
		return res + " ..."
	default:
		return res
	}
}

func nearestBorder(rs []rune, point int, maxDerivation int, shift int) int {
	if point == len(rs) {
		return point
	}

	for ticker := 1; ticker < maxDerivation; ticker++ {
		ahead := point + ticker
		behind := point - ticker

		switch {
		case ticker > maxDerivation:
			return point
		case ahead >= len(rs):
			return len(rs)
		case behind <= 0:
			return 0
		case isBorder(rs[ahead]):
			return ahead
		case isBorder(rs[behind]):
			return behind
		}
	}

	return point
}
