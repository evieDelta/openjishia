package highlights

import (
	"fmt"
	"runtime/debug"
	"strings"
	"time"
	"unicode"

	"github.com/bwmarrin/discordgo"
	"github.com/eviedelta/openjishia/wlog"
	"github.com/pkg/errors"
)

func in(s string, l []string) bool {
	for _, x := range l {
		if x == s {
			return true
		}
	}
	return false
}

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
	globallock.RLock()
	defer globallock.RUnlock()

	if globalruntime.Guildsettings[m.GuildID] == nil {
		return
	}
	if !globalruntime.Guildsettings[m.GuildID].Enable {
		return
	}
	if len(m.Content) <= 1 {
		return
	}

	addLimit(getrlimkey(m.Author.ID, m.GuildID, ""), delaySelf)

	//	fmt.Println("processing message: ", m.ID)

	for u, d := range globalruntime.Guildsettings[m.GuildID].MemberSettings {
		//		fmt.Println(m.ID, " | checking highlights for user ", u)
		if isLimited(getrlimkey(u, m.GuildID, "")) || d.Disabled {
			//			fmt.Println(m.ID, " | user ", u, " is currently on cooldown")
			continue
		}
		if p, err := s.State.UserChannelPermissions(u, m.ChannelID); err != nil {
			globalruntime.Guildsettings[m.GuildID].MemberSettings[u].Disabled = true
			wlog.Err.Print(errors.Wrapf(err, "@s.State.UserChannelPermissions (Check Permission for user %v)", u))

			continue
		} else if p&discordgo.PermissionViewChannel == 0 {
			//			fmt.Println(m.ID, " | user ", u, " does not have permissions to view channel")
			continue
		}
		if d.Blocks[m.Author.ID].State || d.Blocks[m.ChannelID].State {
			//			fmt.Println(m.ID, " | user ", u, " highlight blocked by user block config")
			continue
		}

		//		fmt.Println(m.ID, " | user ", u, " going forward with highlight checking")

		err := doUserHighlight(s, m, u, d)
		if err != nil {
			wlog.Err.Print(errors.Wrapf(err, "@doUserHighlight (For user %v)", u))
		}
	}
	return nil
}

func indexCheck(r rune) bool { return unicode.IsSpace(r) || unicode.IsPunct(r) }

func nextIndexAfter(l int, s string) int {
	return strings.IndexFunc(s[l:],
		func(r rune) bool { return unicode.IsSpace(r) || unicode.IsPunct(r) },
	)
}

func checkHighlight(mst string, y string, x *highlight, user string, m *discordgo.MessageCreate) (start, end int) {
	inc := 0
	incTp := func() bool {
		//	fmt.Println(m.ID, "| user", user, "| word", y, "| bumping area")
		if idx := nextIndexAfter(0, mst); idx >= 0 {
			mst = mst[idx+1:]
			inc += idx + 1
			//	fmt.Println(m.ID, "| user", user, "| word", y, "| inc", inc)
			return false
		}
		return true
	}

	for i := 0; i < 0o10000; i++ { // infinite loop protection
		if len([]rune(y)) > len([]rune(mst)) || len([]rune(mst)) <= 1 {
			//	fmt.Println(m.ID, "| user", user, "| word", y, "| lenght of y is greater than remainder, ending. ")
			break
		}

		//	fmt.Println(m.ID, "| user", user, "| word", y, "| checking", string([]rune(mst)[:len(y)]))

		if unicode.IsSpace([]rune(y)[0]) {
			//	fmt.Println(m.ID, "| user", user, "| word", y, "| next is spare space, bumping")
			if incTp() {
				break
			}
			continue
		}

		if !strings.EqualFold(y, string([]rune(mst)[:len([]rune(y))])) {
			//	fmt.Println(m.ID, "| user", user, "| word", y, "| word not equal, bumping")
			if incTp() {
				break
			}
			continue
		}

		bump := 0
		if len(mst) > len(y) && strings.EqualFold("s", string([]rune(mst)[len([]rune(y))])) {
			//	fmt.Println(m.ID, "| user", user, "| word", y, "| word is plural, adding bump")
			bump = 1
		}

		if len(mst) > len(y)+bump && !indexCheck([]rune(mst)[len([]rune(y))+bump]) {
			//	fmt.Println(m.ID, "| user", user, "| word", y, "| next border check = ,", string([]rune(mst)[len([]rune(y))+bump]))
			//	fmt.Println(m.ID, "| user", user, "| word", y, "| word doesn't seem to be whole word, bumping")
			if incTp() {
				break
			}
			continue
		}
		// fmt.Println(m.ID, "| user", user, "| word", y, "| results seem nominal", inc, inc+len(y))
		return inc, inc + len(y)
	}
	return -1, -1
}

func doUserHighlight(s *discordgo.Session, m *discordgo.MessageCreate, user string, settings *membersettings) (err error) {
	if settings == nil || len(settings.Highlightwords) < 1 {
		// fmt.Println(m.ID, "| user", user, "| user has no settings or no highlights")
		return
	}

	for y, x := range settings.Highlightwords {
		// fmt.Println(m.ID, "| user", user, "| proccessing word", y)
		if isLimited(getrlimkey(user, m.GuildID, y)) {
			// fmt.Println(m.ID, "| user", user, "| word", y, "is currently on cooldown")
			continue
		}

		if start, end := checkHighlight(m.Content, y, x, user, m); start >= 0 && end >= 0 {
			// fmt.Println(m.ID, "| user", user, "| word", y, "everything is clear, marking new cooldowns and moving to send alert")
			addLimit(getrlimkey(user, m.GuildID, ""), delayAny)
			addLimit(getrlimkey(user, m.GuildID, y), delaySpecific)

			err = sendHighlight(s, m, user, m.Content[start:end], y)
			return errors.Wrap(err, "@sendHighlight")
		}
	}
	return
}

func sendHighlight(s *discordgo.Session, m *discordgo.MessageCreate, targetuser, rec, term string) error {
	uch, err := s.UserChannelCreate(targetuser)
	if err != nil {
		return errors.Wrap(err, "@UserChannelCreate (Get DM channel)")
	}
	msgs, err := s.ChannelMessages(m.ChannelID, 5, "", "", "")
	if err != nil {
		return errors.Wrap(err, "@s.ChannelMessages (Get msgs)")
	}

	fields := make([]*discordgo.MessageEmbedField, 0, 5)

	for z := len(msgs) - 1; z >= 0; z-- {
		ct := msgs[z].Content
		if len(msgs[z].Content) > 110 {
			ct = msgs[z].Content[:100] + " ..."
		}

		author := msgs[z].Author.Username
		mem, err := s.State.Member(m.GuildID, msgs[z].Author.ID)
		if err == nil && mem.Nick != "" {
			author = mem.Nick
		}

		t, err := discordgo.SnowflakeTimestamp(msgs[z].ID)
		if err != nil {
			return errors.Wrap(err, "@dgo.SnowflakeTimestamp (Parse msg time)")
		}

		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("``[%v]`` %v", t.UTC().Format("15:04:05"), author),
			Value: "\u200B" + ct,
		})
	}

	//	fmt.Println(len(fields), fields)

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

			Timestamp: string(m.Timestamp),

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

	// user nolock because its already locked
	addDeletionQueueNoLock(msg.ID, msg.ChannelID, messageKeepTime)

	return err
}

const messageKeepTime = time.Hour * 24
