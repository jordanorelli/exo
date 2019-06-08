package main

import (
	"fmt"
	"time"
)

type IdleState struct {
	CommandSuite
	NopExit
	*System
}

func Idle(sys *System) ConnectionState {
	i := &IdleState{System: sys}
	i.CommandSuite = CommandSet{
		balCommand,
		playersCommand,
		BroadcastCommand(sys),
		NearbyCommand(sys),
		Command{
			name:    "goto",
			summary: "travel between star systems",
			arity:   1,
			handler: i.travelTo,
		},
		Command{
			name:    "bomb",
			summary: "bomb another star system",
			arity:   1,
			usage:   "bomb [system-name or system-id]",
			handler: i.bomb,
		},
		Command{
			name:    "mine",
			summary: "mine the current system for resources",
			arity:   0,
			handler: i.mine,
		},
		Command{
			name:    "scan",
			summary: "scans the galaxy for signs of life",
			arity:   0,
			handler: i.scan,
		},
		Command{
			name:    "make",
			summary: "makes things",
			handler: i.maek,
		},
	}
	return i
}

func (i *IdleState) Enter(c *Connection) {
	i.System.Arrive(c)
}

func (i *IdleState) String() string {
	return fmt.Sprintf("idle on %v", i.System)
}

func (i *IdleState) Tick(c *Connection, frame int64) ConnectionState {
	return i
}

func (i *IdleState) travelTo(c *Connection, args ...string) {
	dest := c.game.galaxy.GetSystem(args[0])
	if dest == nil {
		c.Printf("no such system: %s", args[0])
		return
	}
	c.SetState(NewTravel(c, i.System, dest))
}

func (i *IdleState) bomb(c *Connection, args ...string) {
	if c.bombs <= 0 {
		c.Printf("Cannot send bomb: no bombs left!  Build more bombs!\n")
		return
	}
	if time.Since(c.lastBomb) < 5*time.Second {
		c.Printf("Cannot send bomb: bombs are reloading\n")
		return
	}

	target := c.game.galaxy.GetSystem(args[0])
	if target == nil {
		c.Printf("Cannot send bomb: no such system: %v\n", args[0])
		return
	}

	c.bombs -= 1
	c.lastBomb = time.Now()
	bomb := NewBomb(c, i.System, target)
	c.game.Register(bomb)
}

func (i *IdleState) mine(c *Connection, args ...string) {
	c.SetState(Mine(i.System))
}

func (i *IdleState) scan(c *Connection, args ...string) {
	if time.Since(c.lastScan) < 1*time.Minute {
		return
	}
	c.Printf("Scanning the galaxy for signs of life...\n")
	c.game.Register(NewScan(i.System))
}

// "make" is already a keyword
func (i *IdleState) maek(c *Connection, args ...string) {
	if len(args) != 1 {
		c.Printf("not sure what to do! Expecting a command like this: make [thing]\ne.g.:\nmake bomb\nmake colony")
		return
	}
	switch args[0] {
	case "bomb":
		if c.money < options.bombCost {
			c.Printf("Not enough money!  Bombs costs %v but you only have %v space duckets.  Mine more space duckets!\n", options.bombCost, c.money)
			return
		}
		c.SetState(MakeBomb(i.System))
	case "colony":
		MakeColony(c, i.System)
		return
	case "shield":
		MakeShield(c, i.System)
	default:
		c.Printf("I don't know how to make a %v.\n", args[0])
	}
}

func (i *IdleState) FillStatus(c *Connection, s *status) {
	s.Location = i.System.String()
	s.Description = "Just hanging out, enjoying outer space."
}
