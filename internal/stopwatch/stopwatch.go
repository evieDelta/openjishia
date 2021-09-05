package stopwatch

import "time"

// New returns a fresh new stopwatch
func New() *Stopwatch {
	return &Stopwatch{}
}

// NewAndStart creates a new stopwatch and starts it
// mainly exists just to reduce a line of boilerplate
func NewAndStart() *Stopwatch {
	s := New()
	defer s.Start()
	return s
}

// Stopwatch calculates
type Stopwatch struct {
	durTotal  time.Duration
	active    bool
	startTime time.Time
}

// Start starts the counting
func (s *Stopwatch) Start() {
	s.startTime = time.Now()
	s.active = true
}

// Resume unpauses the stopwatch
func (s *Stopwatch) Resume() {
	s.startTime = time.Now()
	s.active = true
}

// Current returns the current time
func (s *Stopwatch) Current() time.Duration {
	if !s.active {
		return s.durTotal
	}
	return s.durTotal + time.Now().Sub(s.startTime)
}

// Pause the stopwatch and return the current total
func (s *Stopwatch) Pause() time.Duration {
	if !s.active {
		return s.Current()
	}
	s.active = false
	s.durTotal += time.Now().Sub(s.startTime)
	return s.durTotal
}

// Stop and reset the stopwatch for future use, returns the final time
func (s *Stopwatch) Stop() time.Duration {
	dr := s.Pause()
	s.Reset()
	return dr
}

// Reset resets the stopwatch for future use
func (s *Stopwatch) Reset() time.Duration {
	d := s.Current()
	s.active = false
	s.durTotal = 0
	s.startTime = time.Time{}
	return d
}
