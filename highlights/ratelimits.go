package highlights

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	delaySelf     = time.Minute * 5
	delayAny      = time.Minute * 1
	delaySpecific = time.Minute * 8
)

var ratelimit = map[string]time.Time{}
var ratelimitLock = sync.RWMutex{}
var ratelimitClean = rate.NewLimiter(rate.Every(time.Minute*30), 1)

func getrlimkey(userID, guildID, tag string) string {
	return guildID + ":" + userID + ":" + tag
}

func addLimit(key string, dur time.Duration) {
	//	fmt.Println("adding ratelimit key ", key, " for ", dur.String())

	cleanRateLimit()

	ratelimitLock.Lock()
	defer ratelimitLock.Unlock()

	ratelimit[key] = time.Now().Add(dur)

	//	ls := "> Active\n"
	//	for i, x := range ratelimit {
	//		if ratelimit[i].After(time.Now()) {
	//			ls += i + " | " + x.Sub(time.Now()).Truncate(time.Millisecond).String() + "\n"
	//		}
	//	}
	//	ls += "> Inactive\n"
	//	for i, x := range ratelimit {
	//		if !ratelimit[i].After(time.Now()) {
	//			ls += i + " | " + x.Sub(time.Now()).Truncate(time.Millisecond).String() + "\n"
	//		}
	//	}
	//	fmt.Println(ls)
}

func isAllow(key string) bool {
	ratelimitLock.RLock()
	defer ratelimitLock.RUnlock()

	st := !ratelimit[key].After(time.Now())
	//	fmt.Println(key, " | ", st)
	return st
}

func isLimited(key string) bool {
	return !isAllow(key)
}

func init() {
	ratelimitClean.Allow()
}

func cleanRateLimit() {
	if !ratelimitClean.Allow() {
		return
	}
	ratelimitLock.Lock()
	defer ratelimitLock.Unlock()

	//	wlog.Info.Print("Running ratelimit cache cleaner...")
	for i, x := range ratelimit {
		if x.Before(time.Now()) {
			//			fmt.Println("removing ", i, x.String(), "behind")
			delete(ratelimit, i)
		}
	}
}
