package highlights

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	delaySelf     = time.Minute * 5
	delayAny      = time.Minute * 5
	delaySpecific = time.Minute * 15
)

var hlratelimit = map[string]time.Time{}
var hlratelimitlock = sync.RWMutex{}
var hlratelimitclean = rate.NewLimiter(rate.Every(time.Minute*30), 1)

func getrlimkey(userid, guildid, tag string) string {
	return guildid + ":" + userid + ":" + tag
}

func addLimit(key string, dur time.Duration) {
	//	fmt.Println("adding ratelimit key ", key, " for ", dur.String())

	cleanratelimit()

	hlratelimitlock.Lock()
	defer hlratelimitlock.Unlock()

	hlratelimit[key] = time.Now().Add(dur)

	//	ls := "> Active\n"
	//	for i, x := range hlratelimit {
	//		if hlratelimit[i].After(time.Now()) {
	//			ls += i + " | " + x.Sub(time.Now()).Truncate(time.Millisecond).String() + "\n"
	//		}
	//	}
	//	ls += "> Inactive\n"
	//	for i, x := range hlratelimit {
	//		if !hlratelimit[i].After(time.Now()) {
	//			ls += i + " | " + x.Sub(time.Now()).Truncate(time.Millisecond).String() + "\n"
	//		}
	//	}
	//	fmt.Println(ls)
}

func isAllow(key string) bool {
	hlratelimitlock.RLock()
	defer hlratelimitlock.RUnlock()

	st := !hlratelimit[key].After(time.Now())
	//	fmt.Println(key, " | ", st)
	return st
}

func isLimited(key string) bool {
	return !isAllow(key)
}

func init() {
	hlratelimitclean.Allow()
}

func cleanratelimit() {
	if !hlratelimitclean.Allow() {
		return
	}
	hlratelimitlock.Lock()
	defer hlratelimitlock.Unlock()

	//	wlog.Info.Print("Running ratelimit cache cleaner...")
	for i, x := range hlratelimit {
		if x.Before(time.Now()) {
			//			fmt.Println("removing ", i, x.String(), "behind")
			delete(hlratelimit, i)
		}
	}
}
