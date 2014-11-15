package main

import (
	"fmt"
)

var idleCommands = CommandSet{
	balCommand,
	commandsCommand,
	helpCommand,
	playersCommand,
}

type IdleState struct {
	CommandSuite
	*System
}

func Idle(sys *System) ConnectionState {
	return &IdleState{idleCommands, sys}
}

func (i *IdleState) String() string {
	return fmt.Sprintf("idle on %v", i.System)
}

func (i *IdleState) Enter(c *Connection) {
	c.Printf("You have landed on %v.\n", i.System)
}

func (i *IdleState) Tick(c *Connection, frame int64) ConnectionState {
	return i
}

func (i *IdleState) Exit(c *Connection) {
	c.Printf("Now leaving %v.\n", i.System)
}

func (i *IdleState) travelTo(c *Connection, args ...string) {
	dest, err := GetSystem(args[0])
	if err != nil {
		c.Printf("%v\n", err)
		return
	}
	c.SetState(NewTravel(c, i.System, dest))
}

func (i *IdleState) GetCommand(name string) *Command {
	return idleCommands.GetCommand(name)
}

// func (i *IdleState) RunCommand(c *Connection, name string, args ...string) ConnectionState {
// 	switch name {
// 	case "goto":
// 		dest, err := GetSystem(args[0])
// 		if err != nil {
// 			c.Printf("%v\n", err)
// 			break
// 		}
// 		return NewTravel(c, i.System, dest)
// 	case "nearby":
// 		neighbors, err := i.Nearby(25)
// 		if err != nil {
// 			log_error("unable to get neighbors: %v", err)
// 			break
// 		}
// 		c.Printf("--------------------------------------------------------------------------------\n")
// 		c.Printf("%-4s %-20s %s\n", "id", "name", "distance")
// 		c.Printf("--------------------------------------------------------------------------------\n")
// 		for _, neighbor := range neighbors {
// 			other := index[neighbor.id]
// 			c.Printf("%-4d %-20s %v\n", other.id, other.name, neighbor.distance)
// 		}
// 		c.Printf("--------------------------------------------------------------------------------\n")
// 	default:
// 		c.Printf("No such command: %v\n", name)
// 	}
// 	return i
// }
