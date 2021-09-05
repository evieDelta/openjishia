package sid

import (
	"strconv"
	"strings"
	"sync"
	"time"
)

// Parse parses a Sid-Snowflake from a String ID
// in either the typical base 10 form, or the base 16 form prefixed by x
func Parse(id string) (ID, error) {
	if strings.HasPrefix(id, "x") {
		i, err := strconv.ParseUint(id[1:], 16, 64)
		return ID(i), err
	}
	i, err := strconv.ParseUint(id, 10, 64)
	return ID(i), err
}

// ID is a Sid-Snowflake ID
type ID uint64

// IsZero simply returns whether or not an ID is present or not
func (i ID) IsZero() bool { return i == 0 }

// Milliseconds returns a time.Duration of the time in an ID,
// note that since ID does not contain the epoch it can only give the duration since epoch and not the absolute time,
// you can get the absolute time by adding the duration to the epoch time or via Generator.IDTime
func (i ID) Milliseconds() time.Duration {
	t := i >> 22
	return time.Duration(t) * time.Millisecond
}

func (i ID) String() string {
	return i.Base10()
}

// Base10 returns a base10 encoded representation of i
func (i ID) Base10() string {
	return strconv.FormatUint(uint64(i), 10)
}

// Base16 returns a base16 encoded representation of i
func (i ID) Base16() string {
	return "x" + strconv.FormatUint(uint64(i), 10)
}

func (i ID) MarshalJSON() ([]byte, error) {
	return []byte("\"" + strconv.FormatUint(uint64(i), 10) + "\""), nil
}

func (i *ID) UnmarshalJSON(b []byte) error {
	s := string(b)
	s = strings.Trim(s, "\"")
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	*i = ID(n)
	return nil
}

// Generator contains some settings for id generation
type Generator struct {
	Epoch  time.Time
	Worker int

	increment uint64
	m         sync.Mutex
}

// Get returns a new fresh ID
func (g *Generator) Get() ID {
	var id uint64

	t := time.Now().Sub(g.Epoch).Milliseconds()
	id |= (uint64(t) << 22)
	id |= (uint64(g.Worker) & 0b1111111111 << 12)
	id |= (g.newIncrement())

	return ID(id)
}

// IDTime gets the time relative to an ID
func (g *Generator) IDTime(i ID) time.Time {
	return g.Epoch.Add(i.Milliseconds())
}

func (g *Generator) newIncrement() uint64 {
	g.m.Lock()
	defer g.m.Unlock()
	g.increment++
	if g.increment >= 0xFFF {
		g.increment = 0
	}
	return uint64(g.increment)
}
