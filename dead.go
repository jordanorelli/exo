package main

import ()

type DeadState struct {
	CommandSuite
	start int64
}

func NewDeadState(died int64) ConnectionState {
	return &DeadState{start: died}
}

func (d *DeadState) Enter(c *Connection) {
	c.Printf("You are dead.\n")
}

func (d *DeadState) Tick(c *Connection, frame int64) ConnectionState {
	if frame-d.start > options.respawnFrames {
		return SpawnRandomly()
	}
	return d
}

func (d *DeadState) Exit(c *Connection) {
	c.Printf("You're alive again.\n")
}

func (d *DeadState) String() string {
	return "dead"
}

func (d *DeadState) PrintStatus(c *Connection) {
	panic("not done")
}
