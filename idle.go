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
		helpCommand,
		playersCommand,
		BroadcastCommand(sys),
		NearbyCommand(sys),
		Command{
			name:    "goto",
			help:    "travel between star systems",
			arity:   1,
			handler: i.travelTo,
		},
		Command{
			name:    "bomb",
			help:    "bomb another star system",
			arity:   1,
			handler: i.bomb,
		},
		Command{
			name:    "mine",
			help:    "mine the current system for resources",
			arity:   0,
			handler: i.mine,
		},
		Command{
			name:    "info",
			help:    "gives you information about the current star system",
			arity:   0,
			handler: i.info,
		},
		Command{
			name:    "scan",
			help:    "scans the galaxy for signs of life",
			arity:   0,
			handler: i.scan,
		},
		Command{
			name:    "make",
			help:    "makes things",
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
	dest, err := GetSystem(args[0])
	if err != nil {
		c.Printf("%v\n", err)
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

	target, err := GetSystem(args[0])
	if err != nil {
		c.Printf("Cannot send bomb: %v\n", err)
		return
	}

	c.bombs -= 1
	c.lastBomb = time.Now()
	bomb := NewBomb(c, i.System, target)
	currentGame.Register(bomb)
}

func (i *IdleState) mine(c *Connection, args ...string) {
	c.SetState(Mine(i.System))
}

func (i *IdleState) info(c *Connection, args ...string) {
	c.Printf("Currently idle on system %v\n", i.System)
	c.Printf("Space duckets available: %v\n", i.money)
}

func (i *IdleState) scan(c *Connection, args ...string) {
	if time.Since(c.lastScan) < 1*time.Minute {
		return
	}
	c.Printf("Scanning the galaxy for signs of life...\n")
	currentGame.Register(NewScan(i.System))
}

// "make" is already a keyword
func (i *IdleState) maek(c *Connection, args ...string) {
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
	default:
		c.Printf("I don't know how to make a %v.\n", args[0])
	}
}
