package main

import (
	"fmt"
	"time"
)

type IdleState struct {
	CommandSuite
	NopEnter
	NopExit
	*System
}

func Idle(sys *System) ConnectionState {
	i := &IdleState{System: sys}
	i.CommandSuite = CommandSet{
		balCommand,
		commandsCommand,
		helpCommand,
		playersCommand,
		Command{
			name:    "goto",
			help:    "travel between star systems",
			arity:   1,
			handler: i.travelTo,
		},
		Command{
			name:    "nearby",
			help:    "list nearby star systems",
			arity:   0,
			handler: i.nearby,
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
	}
	return i
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

func (i *IdleState) nearby(c *Connection, args ...string) {
	neighbors, err := i.Nearby(25)
	if err != nil {
		log_error("unable to get neighbors: %v", err)
		return
	}
	c.Printf("--------------------------------------------------------------------------------\n")
	c.Printf("%-4s %-20s %s\n", "id", "name", "distance")
	c.Printf("--------------------------------------------------------------------------------\n")
	for _, neighbor := range neighbors {
		other := index[neighbor.id]
		c.Printf("%-4d %-20s %v\n", other.id, other.name, neighbor.distance)
	}
	c.Printf("--------------------------------------------------------------------------------\n")
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
}
