package ratelimiter

import (
	"sync"
	"time"
)

const cleanPeriod = time.Minute * 5

const cleanKey = "~.:Internal.CleanKey"

// New returns a new limiter
func New(c Config) *Limiter {
	return &Limiter{
		Config: c,
		store:  make(map[string]*rkey),
	}
}

// Config for the ratelimiter
type Config struct {
	// If false it will only set a new limit if the action is allowed to happen
	// If true then they will set the limit period regardless
	SetWhenLimited bool

	// the timer period before limits are clear
	Period time.Duration
}

// Limiter is a ratelimit management thingy
type Limiter struct {
	Config

	store map[string]*rkey
	sync  sync.RWMutex
}

type rkey struct {
	t time.Time
}

func nrkey() *rkey { return &rkey{} }

// HowLong returns how long key has to wait before it can try again
func (l *Limiter) HowLong(key string) time.Duration {
	l.sync.RLock()
	defer l.sync.RUnlock()

	k := l.store[key]
	if k == nil {
		return 0
	}
	if time.Now().After(k.t) {
		return 0
	}
	return time.Now().Sub(k.t)
}

// DoWait waits until the next event can happen
func (l *Limiter) DoWait(key string) {
	l.XDoWait(key, l.Period)
}

// XDoWait is DoWait but with custom time
func (l *Limiter) XDoWait(key string, period time.Duration) {
	if l.Allowed(key) {
		l.XLimit(key, period)
		return
	}
	time.Sleep(l.store[key].t.Sub(time.Now()))
	l.XLimit(key, period)
}

// Do returns if an action is allowed to happen at this moment, and sets a new limit
func (l *Limiter) Do(key string) bool {
	return l.XDo(key, l.Period)
}

// XDo is like do except you give it a custom time
func (l *Limiter) XDo(key string, dur time.Duration) bool {
	l.clean()

	st := l.Allowed(key)
	if st || l.SetWhenLimited {
		l.XLimit(key, dur)
	}
	return st
}

// Dont is the inverse of Do
func (l *Limiter) Dont(key string) bool {
	return l.XDont(key, l.Period)
}

// XDont is the inverse of XDo
func (l *Limiter) XDont(key string, dur time.Duration) bool {
	return !l.XDo(key, dur)
}

// Limit sets a new limit on key
func (l *Limiter) Limit(key string) {
	l.XLimit(key, l.Period)
}

// XLimit sets a new limit on key with a custom duration
func (l *Limiter) XLimit(key string, dur time.Duration) {
	l.clean()

	l.addLimit(key, dur)
}

func (l *Limiter) addLimit(key string, dur time.Duration) {
	l.sync.Lock()
	defer l.sync.Unlock()

	if l.store[key] == nil {
		l.store[key] = nrkey()
	}
	l.store[key].t = time.Now().Add(dur)
}

// Allowed returns if an action is allowed to happen but does not set a new limit
func (l *Limiter) Allowed(key string) bool {
	l.sync.RLock()
	defer l.sync.RUnlock()

	k := l.store[key]
	if k == nil {
		return true
	}
	st := k.t.After(time.Now())

	//	fmt.Println(key, " | ", st)
	return !st
}

// Disallowed is the inverse of Allowed
func (l *Limiter) Disallowed(key string) bool {
	return !l.Allowed(key)
}

func (l *Limiter) clean() {
	if l.Disallowed(cleanKey) {
		return
	}
	l.addLimit(cleanKey, cleanPeriod)
	l.sync.Lock()
	defer l.sync.Unlock()

	for i, x := range l.store {
		if x.t.Before(time.Now()) {
			delete(l.store, i)
		}
	}
}
