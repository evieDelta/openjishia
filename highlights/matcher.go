package highlights

import (
	"strings"
	"unicode"

	"github.com/bwmarrin/discordgo"
)

func isBorder(r rune) bool { return unicode.IsSpace(r) || unicode.IsPunct(r) }

func nextIndexAfter(l int, s string) int {
	return strings.IndexFunc(s[l:], isBorder)
}

func checkHighlight(mst string, y string, user string, m *discordgo.MessageCreate) (start, end int) {
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

		if len(mst) > len(y)+bump && !isBorder([]rune(mst)[len([]rune(y))+bump]) {
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
