package highlights

import (
	"strconv"
	"time"
)

const messageKeepTime = time.Hour * 24
const previewSize = 150

const maxHighlights = 40
const maxHighlightLength = 128

var maxHighlightsString = strconv.Itoa(maxHighlights)
var maxHighlightLengthString = strconv.Itoa(maxHighlightLength)

const notEnabledMessage = "Highlights are currently not enabled on this server\nUse ``hlconf enable`` (req: Manage Server) to enable"

var truefalseemote = map[bool]string{
	true:  "✅",
	false: "❌",
}
