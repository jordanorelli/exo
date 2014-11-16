package main

import (
	"fmt"
	"sort"
	// "strconv"
	"strings"
)

var commandRegistry map[string]*Command

type Command struct {
	name     string
	help     string
	arity    int
	variadic bool
	handler  func(*Connection, ...string)
	debug    bool // marks command as a debug mode command
}

type CommandSuite interface {
	GetCommand(name string) *Command
	Commands() []Command
}

func (c Command) GetCommand(name string) *Command {
	if name == c.name {
		return &c
	}
	return nil
}

func (c Command) Commands() []Command {
	return []Command{c}
}

type CommandSet []Command

func (c CommandSet) GetCommand(name string) *Command {
	for _, cmd := range c {
		if cmd.name == name {
			return &cmd
		}
	}
	return nil
}

func (c CommandSet) Commands() []Command {
	return []Command(c)
}

var helpCommand = Command{
	name: "help",
	help: "helpful things to help you",
	handler: func(conn *Connection, args ...string) {
		msg := `
Star Dragons is a stupid name, but it's the name that Brian suggested.  It has
nothing to do with Dragons.

Anyway, Star Dragons is a game of cunning text-based, real-time strategy.  You
play as some kind of space-faring entity, faring space in your inspecific
space-faring vessel.  If you want a big one, it's big; if you want a small one,
it's small.  If you want a pink one, it's pink, if you want a black one, it's
black.  And so on, and so forth.  It is the space craft of your dreams.  Or
perhaps you are one of those insect-like alien races and you play as the queen.
Yeah, that's the ticket!  You're the biggest baddest queen bug in space.

In Star Dragons, you issue your spacecraft (which is *not* called a Dragon)
textual commands to control it.  The objective of the game is to be the first
person or alien or bug or magical space ponycorn to eradicate three enemy
species.  Right now that is the only win condition.

All of the systems present in Star Dragons are named and positioned after known
exoplanet systems.  When attempting to communicate from one star system to
another, it takes time for the light of your message to reach the other star
systems.  Star systems that are farther away take longer to communicate with.
        `
		msg = strings.TrimSpace(msg)
		fmt.Fprintln(conn, msg)

		if len(args) == 0 {
			fmt.Fprintln(conn, `use the "commands" command for a list of commands.`)
			fmt.Fprintln(conn, `use "help [command-name]" to get info for a specific command.`)
			return
		}
		for _, cmdName := range args {
			cmd, ok := commandRegistry[cmdName]
			if !ok {
				conn.Printf("no such command: %v\n", cmdName)
				continue
			}
			conn.Printf("%v: %v\n", cmdName, cmd.help)
		}
	},
}

var commandsCommand = Command{
	name: "commands",
	help: "gives you a handy list of commands",
	handler: func(conn *Connection, args ...string) {
		names := make([]string, 0, len(commandRegistry))
		for name, _ := range commandRegistry {
			names = append(names, name)
		}
		sort.Strings(names)
		fmt.Fprintln(conn, "--------------------------------------------------------------------------------")
		for _, name := range names {
			cmd := commandRegistry[name]
			conn.Printf("%-16s %s\n", name, cmd.help)
		}
		fmt.Fprintln(conn, "--------------------------------------------------------------------------------")
	},
}

func BroadcastCommand(sys *System) Command {
	return Command{
		name: "broadcast",
		help: "broadcast a message for all systems to hear",
		handler: func(c *Connection, args ...string) {
			msg := strings.Join(args, " ")
			b := NewBroadcast(sys, msg)
			log_info("player %s send broadcast from system %v: %v\n", c.Name(), sys, msg)
			currentGame.Register(b)
		},
	}
}

func NearbyCommand(sys *System) Command {
	handler := func(c *Connection, args ...string) {
		neighbors, err := sys.Nearby(25)
		if err != nil {
			log_error("unable to get neighbors: %v", err)
			return
		}
		c.Printf("--------------------------------------------------------------------------------\n")
		c.Printf("%-4s %-20s %s\n", "id", "name", "distance")
		c.Printf("--------------------------------------------------------------------------------\n")
		for _, neighbor := range neighbors {
			other := index[neighbor.id]
			c.Printf("%-4d %-20s %-5.6v\n", other.id, other.name, neighbor.distance)
		}
		c.Printf("--------------------------------------------------------------------------------\n")
	}
	return Command{
		name:    "nearby",
		help:    "list nearby star systems",
		arity:   0,
		handler: handler,
	}
}

var winCommand = Command{
	name:  "win",
	help:  "win the game.",
	debug: true,
	handler: func(conn *Connection, args ...string) {
		conn.Win("win-command")
	},
}

var playersCommand = Command{
	name: "players",
	help: "lists the connected players",
	handler: func(conn *Connection, args ...string) {
		for other, _ := range currentGame.connections {
			conn.Printf("%v\n", other.Name())
		}
	},
}

var balCommand = Command{
	name: "bal",
	help: "displays your current balance in space duckets",
	handler: func(conn *Connection, args ...string) {
		fmt.Fprintln(conn, conn.money)
	},
}
