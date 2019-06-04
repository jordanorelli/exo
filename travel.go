package main

import (
	"fmt"
	"text/template"
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
				c.Printf("%v\n", t.remaining())
			},
		},
	}
	return t
}

var enterTravelTemplate = template.Must(template.New("enter-travel").Parse(`
Departing:       {{.Departing}}
Destination:     {{.Destination}}
Total Trip Time: {{.Duration}}
Current Time:    {{.Time}}
ETA:             {{.ETA}}
`))

func (t *TravelState) Enter(c *Connection) {
	enterTravelTemplate.Execute(c, struct {
		Departing   *System
		Destination *System
		Duration    time.Duration
		Time        time.Time
		ETA         time.Time
	}{
		t.start,
		t.dest,
		t.tripTime(),
		time.Now(),
		t.eta(),
	})
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

func (t *TravelState) PrintStatus(c *Connection) {
	desc := fmt.Sprintf("Traveling from %v to %v", t.start, t.dest)
	statusTemplate.Execute(c, status{
		State:       "In Transit",
		Balance:     c.money,
		Bombs:       c.bombs,
		Kills:       c.kills,
		Location:    fmt.Sprintf("%s -> %s", t.start, t.dest),
		Description: desc,
	})
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
