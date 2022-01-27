package insched

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/eviedelta/openjishia/wlog"
	"github.com/pkg/errors"
)

type blocking struct {
	list map[uint64]bool
	mu   sync.RWMutex
}

func (b *blocking) add(i uint64)      { b.mu.Lock(); defer b.mu.Unlock(); b.list[i] = true }
func (b *blocking) rem(i uint64)      { b.mu.Lock(); defer b.mu.Unlock(); delete(b.list, i) }
func (b *blocking) get(i uint64) bool { b.mu.RLock(); defer b.mu.RUnlock(); return b.list[i] }

type action struct {
	han Handler
	sch *Scheduler

	once sync.Once
	name string

	lastrun time.Time
	block   blocking
}

type failure struct {
	Error   error
	Handler string
	Break   bool
	Entry   *LoggerEntry

	Requested bool
}

func (a *action) scan(halt *bool, poke <-chan struct{}, wentWrong chan failure) {
	// ticker to scan at a set interval
	t := time.NewTicker(a.han.Config().ScanPeriod)
	defer t.Stop()
	ticker := t.C

	// wg to make sure everything has exited before exiting out
	wg := &sync.WaitGroup{}

	// had to add this jankiness to the once reset
	// because it panicked from unlocking an unlocked mutex
	// if i reset it from within the once function
	resetOnce := func() {}
	defer resetOnce()

	// make sure only one runs at a time
	a.once.Do(func() {
		defer func() {
			resetOnce = func() { a.once = sync.Once{} } // but allow another one to start when it exits
		}()

		wg.Add(1)
		defer wg.Done()

		for {
			// exit if required before doing any more fancy stuff
			if !*halt {
				break
			}

			//			wlog.Spam.Printf("running scheduler for %v", a.name)

			// get pending requests
			queue, err := a.sch.db.GetPending(a.name,
				time.Now().UTC().Add(a.han.Config().ScanPeriod+time.Second), // but also include the next second just so we get any on the edge now
			)
			if err != nil {
				// uh oh spaghetti oes
				wentWrong <- failure{
					Error:   err,
					Handler: a.name,
					Break:   true, // break here as its most likely a database problem
				}
				return
			}

			a.lastrun = time.Now()

			if Debug {
				wlog.Spam.Printf("Putting %v entries into live queue for %v", len(queue), a.name)
			}

			// if the handler is precise we call the precise queuer, if not we call the dumb queuer
			if a.han.Config().Precise {
				a.queuePrecise(queue, wentWrong, halt, poke, wg)
			} else {
				a.queue(queue, wentWrong, halt, poke, wg) // queueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueueue
			}

			// check for a halt request before waiting
			if !*halt {
				break
			}

			// typically we wait on the timer
			// but in the case we want to close out it'll also jump past when we do that
			select {
			case <-ticker:
			case <-poke:
				break
			}
		}
	})

	wg.Wait()
}

// dumb queue, this one executes them whenever it does
func (a *action) queue(queue []Entry, wentWrong chan failure, halt *bool, poke <-chan struct{}, wg *sync.WaitGroup) {
	// lets at least put them in the right order though
	sort.Slice(queue, func(i, j int) bool {
		return queue[i].Time.Before(queue[j].Time)
	})

	// we might not be precise but lets evenly spread them across the scan period at least instead of going rapidfire
	timer := time.NewTicker(a.han.Config().ScanPeriod / time.Duration(len(queue)+1))
	defer timer.Stop()

	for _, entry := range queue {
		if a.block.get(entry.ID) {
			continue
		}
		//		wlog.Spam.Printf("queuing id:%v in %v", x.ID, a.han.Config().ScanPeriod/time.Duration(len(que)+1))

		go a.doEntry(entry, wentWrong, wg)

		// wait for either the next tick, or a cancelation
		select {
		case <-poke:
			break
		case <-timer.C:
		}

		if !*halt { // exit if requested
			break
		}
	}
}

// smarter queue, this one executes them about as on the dot as it'll get by time.Sleeping them to the moment
func (a *action) queuePrecise(queue []Entry, wentWrong chan failure, halt *bool, poke <-chan struct{}, wg *sync.WaitGroup) {
	// we put them in order, so we only need to wait on the next one up at each moment
	// and we can have less go routines active waiting at a time as a result
	sort.Slice(queue, func(i, j int) bool {
		return queue[i].Time.Before(queue[j].Time)
	})

	timer := time.NewTimer(time.Millisecond)
	<-timer.C

	for _, entry := range queue {
		if a.block.get(entry.ID) {
			continue
		}

		timer.Reset(entry.Time.Sub(time.Now().UTC()))

		//		wlog.Spam.Printf("precice queuing id:%v in %v", x.ID, x.Time.Sub(time.Now().UTC()))

		// wait for the next one in the queue
		select {
		case <-timer.C:
		case <-poke:
			timer.Stop()
			break
		}

		if !*halt { // exit if requested
			break
		}

		// execute the next entry, using a go routine so a slow task won't hold up the next ones in queue
		go a.doEntry(entry, wentWrong, wg)
	}
}

var Debug bool

// entry executers
func (a *action) doEntry(entry Entry, wentWrong chan failure, wg *sync.WaitGroup) (no bool) {
	// oops protection
	defer func() {
		er := recover()
		if er == nil {
			return
		}

		a.block.add(entry.ID) // as we've panicked prevent this event from running again until a restart
		no = true             // if this is run directly by the schedule function (and not run by the scanner) we use this to tell it to not remove it from the blocker

		// we do the finish entry system, but we use the deferOnPanic setting instead
		a.finishEntry(a.han.Config().DeferOnPanic, entry, wentWrong)

		// get where the panic happened so we can include it in the error
		_, f, l, ok := runtime.Caller(2)
		append := ""
		if ok {
			append = f + ":" + strconv.Itoa(l)
		}

		// build the error
		var err error
		if e, ok := er.(error); ok {
			err = errors.Wrap(e, append)
		} else {
			err = errors.New(append + ": " + fmt.Sprint(er))
		}

		// and send it
		wentWrong <- failure{
			Error:   err,
			Handler: a.name,
			Break:   false, // don't need to break here, its probably just an error in the handler

			Entry: &LoggerEntry{
				Entry: entry,
			},
		}
	}()

	// for the waitgroup given by the scan function, so it won't just quit on us
	wg.Add(1)
	defer wg.Done()

	if Debug {
		wlog.Spam.Printf("Action: %v, ID: %v, Defer? %v\n`%v`\n\n```\n%v\n```", entry.Action, entry.ID, entry.DeferCount, entry.Time.Format("2006-01-02 15:04:05"), entry.Details)
	}

	// call the handler with the entry data
	def, err := a.han.Call(entry)
	if err != nil {
		wentWrong <- failure{
			Error:   err,
			Handler: a.name,
			Break:   false, // don't need to break here, its probably just an error in the handler

			Entry: &LoggerEntry{
				Entry:    entry,
				Deferred: def > 0,
				DeferLen: def,
			},
		}
	}

	// take care of removal or defering
	a.finishEntry(def, entry, wentWrong)

	return
}

func (a *action) finishEntry(d time.Duration, entry Entry, wentWrong chan failure) {
	var err error

	switch {
	case d > 0:
		// if a handler has requested we defer, we do that
		err = a.sch.db.Defer(entry.ID, d)
	case d == 0:
		// else we just remove the event as its done now
		err = a.sch.db.Remove(entry.ID)
	case d < 0:
		// unless its requested we do nothing at all
		return
	}

	if err != nil {
		wentWrong <- failure{
			Error:   errors.Wrap(err, "error removing or deferring entry"),
			Handler: a.name,
			Break:   true, // break here because its likely a database error and signifies something may be wrong with the database connection

			Entry: &LoggerEntry{
				Entry:    entry,
				Deferred: d > 0,
				DeferLen: d,
			},
		}
	}
}

// New scheduler
func New(db Database) *Scheduler {
	return &Scheduler{
		actions: make(map[string]*action),
		db:      db,
		closing: true, // start off closed
		wg:      &sync.WaitGroup{},

		Log: new(ConsoleLogger),
	}
}

// Scheduler manages schedule stuff or something
type Scheduler struct {
	actions   map[string]*action
	db        Database
	mu        sync.Mutex
	closing   bool
	wentWrong chan failure
	wg        *sync.WaitGroup

	Log Logger
}

// Run starts up the scheduler and scanners
// only one can run at a time, it will exit out if there is one running that isn't closing
// if there is one running that is currently closing it will wait on that one to exit, and then will start again from the new call
func (S *Scheduler) Run() error {
	var err error

	// if its not closing we don't really need to sit here waiting do we
	if !S.closing {
		return nil
	}

	// ensure we only have 1 up at a time
	S.mu.Lock()
	defer S.mu.Unlock()

	S.closing = false
	defer func() { S.closing = true }() // ensure another one can start if this one exits

	// jiggery to handle the many concurrent thingies
	wg := S.wg
	run := new(bool)
	*run = true
	poke := make(chan struct{}, len(S.actions))

	// refresh this
	S.wentWrong = make(chan failure)
	defer func() { close(S.wentWrong); S.wentWrong = nil }() // eliminate it on close though

	// start up all the handlers
	for _, x := range S.actions {
		action := x
		go func() {
			wg.Add(1)
			action.scan(run, poke, S.wentWrong)

			wg.Done()
		}()
	}

	// error management
	for {
		f := <-S.wentWrong

		if !f.Requested { // don't need to log requested closures
			S.Log.Error(f.Error, f.Handler, f.Break, f.Entry)
		}

		if f.Break {
			S.closing = true // mark us as closing first thing
			*run = false

			// poke all the handler scanners to make sure they close
			// so we aren't left waiting potentially ages for any
			for i := 0; i < len(S.actions); i++ {
				poke <- struct{}{}
			}

			// the error we will exit with
			err = errors.Wrap(f.Error, "handler: "+f.Handler)

			break
		}
	}

	// wait for all handler scanners to exit
	wg.Wait()

	// close out
	close(poke)

	return err
}

// Stop starts the process of stopping any currently running scheduler
// and waits until Run exits
func (S *Scheduler) Stop() {
	if S.wentWrong == nil {
		return
	}
	S.closing = true
	S.wentWrong <- failure{
		Break:     true,
		Requested: true,
	}
	// cheap way of ensuring run exits before we allow us to exit
	S.mu.Lock()
	S.mu.Unlock()
}

// AddHandler adds an action to the handler list
// action handlers should be setup before runtime and not touched after.
//
// the name should be something constant and unlikely to collide with another package
// it doesn't matter what the name is, the only thing that matters is that its unique
// and will not be changed during operation
// (changing the name without setting up a migration will cause any existing events of the old name to be stuck in limbo)
func (S *Scheduler) AddHandler(name string, a Handler) error {
	if a == nil {
		return errors.New("cannot create a nil handler")
	}

	S.actions[name] = &action{
		han: a,
		sch: S,

		once: sync.Once{},
		name: name,

		block: blocking{list: make(map[uint64]bool)},
	}

	return nil
}

// Schedule an event to be called at a specified time with the specified handler and metadata
// the meta data is stored in JSON in the database, and given to the handler as JSON
// this function can be used directly
// though its recommended to wrap this with handler specific functions for type safety in the meta data input
func (S *Scheduler) Schedule(at time.Time, handler string, data interface{}) (Entry, error) {
	if x, ok := S.actions[handler]; !ok || x.han == nil {
		return Entry{}, errors.New("cannot schedule an event with an unregistered handler")
	}

	at = at.UTC() // screw timezones

	e := Entry{
		Action: handler,
		Time:   at,
	}

	err := e.marshalDetails(data)
	if err != nil {
		return e, err
	}

	e, err = S.db.Add(e)
	if err != nil {
		return e, err
	}

	// if its schedule time is before the next scan
	// we can just queue it here so it isn't stuck waiting for the next scan
	act := S.actions[handler]
	if at.Before(act.lastrun.Add(act.han.Config().ScanPeriod)) {
		go func() {
			act.block.add(e.ID) // thingy so the main scanner won't also try execute it if we're doing it here
			time.Sleep(at.Sub(time.Now()))
			no := act.doEntry(e, S.wentWrong, S.wg)
			if !no {
				act.block.rem(e.ID) // if not no we allow it again
			}
		}()
	}

	return e, nil
}

// Cancel a scheduled event
func (S *Scheduler) Cancel(id uint64) error {
	return S.db.Remove(id)
}

// Reschedule an event to a different time
func (S *Scheduler) Reschedule(id uint64, to time.Time) error {
	return S.db.Reschedule(id, to)
}

// ShiftSchedule of an event forward or backwards by a specified amount of time
func (S *Scheduler) ShiftSchedule(id uint64, by time.Duration) error {
	return S.db.Defer(id, by)
}

// Database is the core storage functions needed to operate
type Database interface {
	// schedule an action into the database
	// ID is not set by the scheduler, the database implementation is expected to handle IDs itself
	// it should return the entry with the ID filled out
	Add(Entry) (Entry, error)

	// Fetch an entry
	Get(id uint64) (Entry, error)

	// look for scheduled items of an action where the scheduled time is before t
	// (t being the current time + the handlers scan period, typically 5 to 240 minutes into the future but can be anything in the future)
	GetPending(action string, t time.Time) (entry []Entry, err error)

	// Remove an entry from the database
	Remove(ID uint64) error

	// Update the time an entry is to be called
	Defer(ID uint64, by time.Duration) error
	Reschedule(id uint64, to time.Time) error
}

// Handler is the interface for implementing an action to be called once the timer runs out
type Handler interface {
	// some configuration for the handler
	Config() HandlerConfig

	// Call handles an entry
	//
	// # Deferring
	// optionally if the handler determines that the action should be attempted again at a later time a Defer length can be given
	// if given the scheduler will keep the entry in the database and add that time to the deadline and call it again at that time
	// instead of removing it from the database as it normally would
	//
	// the defer time also works if an error is given, normally if Call errors out the scheduler
	// will still remove the entry from the database as if it succeeded
	// but this behavior is not always ideal as some errors can be resolved by trying again later,
	// so, if the handler runs into an error and determines that it should try again later instead
	// of stopping, it can give a duration to tell the scheduler to try again after the alloted time
	// instead of marking the entry as completed
	//
	// As an aside, be aware that a defer time less than the handlers ScanPeriod may lead to it being called later than scheduled
	//
	// The DeferCount field in entry marks how many times it has been deferred,
	// so a handler can choose to stop trying again after 3 failed attempts
	Call(Entry) (Defer time.Duration, err error)
}

// HandlerConfig defines some metadata and configuration for handlers
type HandlerConfig struct {
	// if this is enabled it'll attempt to call the function on the dot via time.Sleep
	// (limited by the precision of go routines and time.Sleep)
	// if disabled it'll do the action sometime within the closest scan period
	// (up to the duration of ScanPeriod after an items deadline)
	Precise bool

	// How often it will scan for items that are approching their deadline
	// if Precise is false this also determines the relative expected accuracy
	// (eg if the period is set to 10 minutes it'll be called sometime within 10 minutes after the deadline)
	// not recommended to have this too long, a period of 1-30 minutes is usually fine
	ScanPeriod time.Duration

	// if a handler panics should we
	// - ( > 0) defer it to try again later (after a restart)
	// - ( < 0) try again immedietly after the next restart
	// - (== 0) just delete it
	// note that even if its set to defer it will not run again until a full restart
	// as its assumed panics are most likely an issue in the code and won't be fixed by just trying again
	// also note that -1 currently does not increment the defer count
	DeferOnPanic time.Duration
}

// Entry is a scheduled event
type Entry struct {
	ID      uint64 `db:"id"`
	Action  string `db:"action"`
	Details string `db:"details"`

	DeferCount int `db:"defer_count"`

	// The time it'll attempt to call this entry around
	Time time.Time `db:"time"`
}

// marshalDetails takes a struct and marshals it into json for the Details field
func (e *Entry) marshalDetails(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	e.Details = string(b)
	return nil
}

// UnmarshalDetails takes the json meta data stored within details and unmarshals it into a struct
// this should be used by handlers for getting the needed metadata about a scheduled event
func (e *Entry) UnmarshalDetails(v interface{}) error {
	return json.Unmarshal([]byte(e.Details), v)
	//          ParseInformation
	//          LoadMetaToStruct
	// i'd like to rename this to ReadMetadata or ParseMetadata or something
	// but the current name is perfectly aligned and i don't wanna mess with it
}

// LoggerEntry wraps some info about a scheduled event for the error log
type LoggerEntry struct {
	Entry
	Deferred bool
	DeferLen time.Duration
}

// Logger is used for reporting errors that happen from scheduled events
// as there is otherwise no way to know about errors that happen from scheduled events
// by default if not logger is given to a scheduler it will use the ConsoleLogger
type Logger interface {
	// stop is whether or not it stopped the scheduler
	Error(err error, handler string, stop bool, evt *LoggerEntry)
}

// NilLogger if you don't want the scheduler to log to anything (by default it logs errors to the console if you don't give it a logger)
type NilLogger struct{}

func (l *NilLogger) Error(err error, handler string, stop bool, evt *LoggerEntry) {}

// ConsoleLogger logs errors to the console via stdlib "log"
// be aware it may dump some sensitive information to the console as it drops info as is
// so it is recommended to create your own logging type that understands your Event handlers metadata
type ConsoleLogger struct{}

func (l *ConsoleLogger) Error(err error, handler string, stop bool, evt *LoggerEntry) {
	if stop {
		log.Printf("scheduler error stopping scheduler...")
	}
	if evt != nil {
		log.Printf("error from scheduler handler %v: %v\n\t| Deferred? %v for %v Defer Count: %v\n\t| ID: %v, Originally Scheduled for: %v\n\t| Details: %v",
			handler, err, evt.Deferred, evt.DeferLen, evt.DeferCount, evt.ID, evt.Time, evt.Details)
	} else {
		log.Printf("error from scheduler handler %v: %v\n", handler, err)
	}
}
