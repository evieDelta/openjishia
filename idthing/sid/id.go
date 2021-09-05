package sid

/*
import (
	"math/rand"
	"time"
)

// Epoch is the default Epoch, not recommended to modify this
var Epoch = time.Unix(1609529715, 0)

// Default is the default Generator
var Default = Generator{
	Epoch: Epoch, // the time i wrote this lmao

	// Reduce the risk of conflicting IDs by starting things off at random points
	Worker:    rand.New(rand.NewSource(time.Now().UnixNano())).Intn(0b1111111111),
	increment: uint64(rand.New(rand.NewSource(time.Now().Unix())).Intn(0xFFF)),
}

// Get gives you a new ID from the default generator
func Get() ID {
	return Default.Get()
}

// IDTime returns the time of an ID relative to the default Epoch
func IDTime(i ID) time.Time {
	return Epoch.Add(i.Milliseconds())
}
*/
