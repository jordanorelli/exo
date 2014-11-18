package main

import (
	"fmt"
	"time"
)

type TravelState struct {
	CommandSuite
	start     *System
	dest      *System
	travelled float64 // distance traveled so far in parsecs
	dist      float64 // distance between start and end in parsecs
}

func NewTravel(c *Connection, start, dest *System) ConnectionState {
	t := &TravelState{
		start: start,
		dest:  dest,
		dist:  start.DistanceTo(dest),
	}
	t.CommandSuite = CommandSet{
		helpCommand,
		playersCommand,
		balCommand,
		Command{
			name:    "progress",
			help:    "displays how far you are along your travel",
			arity:   0,
			handler: t.progress,
		},
		Command{
			name:  "eta",
			help:  "displays estimated time of arrival",
			arity: 0,
			handler: func(c *Connection, args ...string) {
				c.Printf("Remaining: %v\n", t.remaining())
				c.Printf("Current time: %v\n", time.Now())
				c.Printf("ETA: %v\n", t.eta())
			},
		},
	}
	return t
}

func (t *TravelState) Enter(c *Connection) {
	c.Printf("Leaving %v, bound for %v.\n", t.start, t.dest)
	c.Printf("Trip duration: %v\n", t.tripTime())
	c.Printf("Current time: %v\n", time.Now())
	c.Printf("ETA: %v\n", t.eta())
	t.start.Leave(c)
}

func (t *TravelState) Tick(c *Connection, frame int64) ConnectionState {
	t.travelled += options.playerSpeed * options.lightSpeed
	if t.travelled >= t.dist {
		return Idle(t.dest)
	}
	return t
}

func (t *TravelState) Exit(c *Connection) {
	c.Printf("You have arrived at %v.\n", t.dest)
	t.dest.Arrive(c)
}

func (t *TravelState) String() string {
	return fmt.Sprintf("Traveling from %v to %v", t.start, t.dest)
}

func (t *TravelState) progress(c *Connection, args ...string) {
	c.Printf("%v\n", t.travelled/t.dist)
}

func (t *TravelState) remaining() time.Duration {
	remaining := t.dist - t.travelled
	frames := remaining / (options.playerSpeed * options.lightSpeed)
	return framesToDur(int64(frames))
}

func (t *TravelState) eta() time.Time {
	// distance remaining in parsecs
	return time.Now().Add(t.remaining())
}

func (t *TravelState) tripTime() time.Duration {
	frames := t.dist / (options.playerSpeed * options.lightSpeed)
	return framesToDur(int64(frames))
}
