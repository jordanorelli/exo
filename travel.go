package main

import (
	"fmt"
)

type TravelState struct {
	CommandSuite
	start     *System
	dest      *System
	travelled float64
	dist      float64
}

func NewTravel(c *Connection, start, dest *System) ConnectionState {
	return &TravelState{
		start: start,
		dest:  dest,
		dist:  start.DistanceTo(dest),
	}
}

func (t *TravelState) Enter(c *Connection) {
	c.Printf("Leaving %v, bound for %v.\n", t.start, t.dest)
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
}

func (t *TravelState) String() string {
	return fmt.Sprintf("Traveling from %v to %v", t.start, t.dest)
}
