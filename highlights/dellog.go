package highlights

import (
	"context"
	"runtime/debug"
	"sync"
	"time"

	"codeberg.org/eviedelta/drc"
	"codeberg.org/eviedelta/trit"
	"github.com/bwmarrin/discordgo"
	"github.com/eviedelta/openjishia/wlog"
	"golang.org/x/time/rate"
)

// dellogtest help
var dellogtest = &drc.Command{
	Name:   "dellogtest",
	Manual: []string{"help"},
	Permissions: drc.Permissions{
		BotAdmin: trit.True,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		Listable:    false,
		MinimumArgs: 0,
	},
	Exec: cfDellogtest,
}

func cfDellogtest(ctx *drc.Context) error {
	m, err := ctx.XReply("uwu to delete in 1 minute")
	if err != nil {
		return err
	}
	addDeletionQueue(m.ID, m.ChannelID, time.Minute)
	return err
}

// delstatus gets the status of the auto deletion system
var delstatus = &drc.Command{
	Name:         "delstatus",
	Manual:       []string{"gets the status of the auto deletion system"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.True,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		Listable:    false,
		MinimumArgs: 0,
	},
	Exec: cfDelstatus,
}

func cfDelstatus(ctx *drc.Context) error {
	globallock.RLock()
	defer globallock.RUnlock()

	now := time.Now()

	list := ""
	for _, x := range globalruntime.DeletionQueue.Entries {
		list += x.ChannelID + ":" + x.MessageID + "\n" +
			"	| at    " + x.TimeMade.Format("2006-01-02 15:04:05") + " | for       " + x.Duration.Truncate(time.Millisecond).String() + "\n" +
			"	| until " + x.TimeMade.Add(x.Duration).Format("2006-01-02 15:04:05") + " | remaining " + x.TimeMade.Add(x.Duration).Sub(now).Truncate(time.Second).String() + "\n"
	}
	return ctx.DumpReply("delstatus", list)
}

// delnow runs the deletion scheduler now
var delnow = &drc.Command{
	Name:         "delnow",
	Manual:       []string{"runs the deletion scheduler now"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.Unset,
		Discord:  0,
	},
	Config: drc.CfgCommand{
		Listable:    false,
		MinimumArgs: 0,
	},
	Exec: cfDelnow,
}

func cfDelnow(ctx *drc.Context) error {
	globalruntime.DeletionQueue.DeleteHandler(ctx.Ses)
	return nil
}

type deletionqueue struct {
	Entries map[string]*deletionentry

	close     bool
	lock      sync.Mutex
	otherlock sync.Mutex
}

func addDeletionQueue(m, c string, dur time.Duration) {
	globallock.Lock()
	defer globallock.Unlock()

	addDeletionQueueNoLock(m, c, dur)
}

func addDeletionQueueNoLock(m, c string, dur time.Duration) {
	d := deletionentry{
		ChannelID: c,
		MessageID: m,
		Duration:  dur,
		TimeMade:  time.Now(),
	}

	globalruntime.DeletionQueue.Entries[m] = &d
}

func (d *deletionqueue) DeletionScheduler(s *discordgo.Session) {
	defer func() {
		err := recover()
		debug.PrintStack()
		wlog.Err.Print(err)
	}()
	d.close = false

	d.lock.Lock()
	defer d.lock.Unlock()

	clock := time.NewTicker(time.Hour * 4)
	defer clock.Stop()

	for {
		if d.close {
			break
		}
		t := <-clock.C
		if Config.Debug {
			wlog.Spam.Print("Running deletion schedualer on ", t.Format(time.Stamp))
		}

		d.otherlock.Lock()

		d.DeleteHandler(s)

		d.otherlock.Unlock()

		err := globalruntime.Save(filename)
		if err != nil {
			wlog.Err.Print("Error in Dellog ", err)
		}
	}
}

func (d *deletionqueue) DeleteHandler(s *discordgo.Session) {
	now := time.Now()
	for _, x := range d.Entries {
		if now.After(x.TimeMade.Add(x.Duration)) {
			go x.Delete(s, d)
		}
	}
}

var delratelimit = rate.NewLimiter(rate.Every(time.Second*10), 3)

func (d *deletionentry) Delete(s *discordgo.Session, dq *deletionqueue) {
	globallock.Lock()
	defer globallock.Unlock()

	if Config.Debug {
		wlog.Spam.Print("deleting ", d.ChannelID, " ", d.MessageID)
	}

	delratelimit.Wait(context.Background())

	delete(dq.Entries, d.MessageID)
	if err := s.ChannelMessageDelete(d.ChannelID, d.MessageID); err != nil {
		wlog.Err.Print(err)
	}
}

type deletionentry struct {
	MessageID string
	ChannelID string

	TimeMade time.Time
	Duration time.Duration
}
