package main

import (
	"fmt"
	"strings"
	"text/template"
)

var helpTemplate = template.Must(template.New("help").Parse(`
{{.Name}} Command Reference

  Summary: {{.Summary}}
{{- if .Usage}}
  Usage:   {{.Usage}}
{{end}}
{{- if .Description}}
Details:

  {{.Description}}
{{end}}
`))

func printHelp(conn *Connection, cmd *Command) {
	desc := strings.ReplaceAll(strings.TrimSpace(cmd.help), "\n", "\n  ")
	helpTemplate.Execute(conn, struct {
		Name        string
		Summary     string
		Usage       string
		Description string
	}{
		Name:        cmd.name,
		Summary:     cmd.summary,
		Usage:       cmd.usage,
		Description: desc,
	})
}

var commandRegistry map[string]*Command

type Command struct {
	name     string
	summary  string
	usage    string
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
	switch name {
	case "help":
		return &helpCommand
	case "commands":
		return &commandsCommand
	case "status":
		return &statusCommand
	}
	for _, cmd := range c {
		if cmd.name == name {
			return &cmd
		}
	}
	return nil
}

func (c CommandSet) Commands() []Command {
	return append([]Command(c), statusCommand, helpCommand, commandsCommand)
}

var helpCommand = Command{
	name:    "help",
	summary: "explains how to play the game",
	usage:   "help [command-name]",
	help: `
help explains the usage of various commands in Exocolonus. On its own, the help
command displays some basic info about how the game is played. If given an
argument of a command name, the help command displays the detailed usage of the
specified command.
`,
	handler: func(conn *Connection, args ...string) {
		msg := `
Exocolonus is a game of cunning text-based, real-time strategy.  You play as
some kind of space-faring entity, faring space in your inspecific space-faring
vessel.  If you want a big one, it's big; if you want a small one, it's small.
If you want a pink one, it's pink, if you want a black one, it's black.  And so
on, and so forth.  It is the space craft of your dreams.  Or perhaps you are
one of those insect-like alien races and you play as the queen.  Yeah, that's
the ticket!  You're the biggest baddest queen bug in space.

In Exocolonus, you issue your spacecraft textual commands to control it.  The
objective of the game is to be the first person or alien or bug or magical
space ponycorn to eradicate three enemy species.  Right now that is the only
win condition.

All of the systems present in Exocolonus are named and positioned after known
exoplanet systems.  Each star system in Exocolonus is a real star system that
has been researched by astronomers, and the number of planets in each system
corresponds to the number of  known exoplanets in those systems. When
attempting to communicate from one star system to another, it takes time for
the light of your message to reach the other star systems.  Star systems that
are farther away take longer to communicate with.
        `
		if len(args) == 0 {
			msg = strings.TrimSpace(msg)
			fmt.Fprintln(conn, msg)
			fmt.Fprint(conn, "\n")
			conn.Line()
			fmt.Fprint(conn, "\n")
			fmt.Fprintln(conn, `use the "commands" command for a list of commands.`)
			fmt.Fprintln(conn, `use "help [command-name]" to get info for a specific command.`)
			return
		}
		for _, cmdName := range args {
			cmd := conn.GetCommand(cmdName)
			if cmd == nil {
				conn.Printf("no such command: %v\n", cmdName)
				continue
			}
			printHelp(conn, cmd)
		}
	},
}

type status struct {
	State       string
	GameCode    string
	Balance     int
	Bombs       int
	Kills       int
	Location    string
	Description string
}

var statusTemplate = template.Must(template.New("status").Parse(`
Current State: {{.State}}
--------------------------------------------------------------------------------
{{- if .GameCode}}
Current Game:  {{.GameCode}}
Balance:       {{.Balance}}
Bombs:         {{.Bombs}}
Kills:         {{.Kills}}
Location:      {{.Location}}
{{end}}

{{.Description}}

`))

var statusCommand = Command{
	name:    "status",
	summary: "display your current status",
	handler: func(conn *Connection, args ...string) {
		s := status{
			State: conn.ConnectionState.String(),
		}
		conn.ConnectionState.FillStatus(conn, &s)
		if conn.game != nil {
			s.GameCode = conn.game.id
			s.Balance = conn.money
			s.Bombs = conn.bombs
			s.Kills = conn.kills
		}
		statusTemplate.Execute(conn, s)
	},
}

// this isn't a real command it just puts command in the list of commands, this
// is weird and circular, this is a special case.
var commandsCommand = Command{
	name:    "commands",
	summary: "lists currently available commands",
}

func BroadcastCommand(sys *System) Command {
	return Command{
		name:    "broadcast",
		summary: "broadcast a message for all systems to hear",
		handler: func(c *Connection, args ...string) {
			msg := strings.Join(args, " ")
			b := NewBroadcast(sys, msg)
			log_info("player %s send broadcast from system %v: %v\n", c.Name(), sys, msg)
			c.game.Register(b)
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
		c.Printf("%-4s %-20s %-12s %s\n", "id", "name", "distance", "trip time")
		c.Printf("--------------------------------------------------------------------------------\n")
		for _, neighbor := range neighbors {
			other := index[neighbor.id]
			dur := NewTravel(c, sys, other).(*TravelState).tripTime()
			c.Printf("%-4d %-20s %-12.6v %v\n", other.id, other.name, neighbor.distance, dur)
		}
		c.Printf("--------------------------------------------------------------------------------\n")
	}
	return Command{
		name:    "nearby",
		summary: "list nearby star systems",
		arity:   0,
		handler: handler,
	}
}

var winCommand = Command{
	name:    "win",
	summary: "win the game.",
	debug:   true,
	handler: func(conn *Connection, args ...string) {
		conn.Win("win-command")
	},
}

var playersCommand = Command{
	name:    "players",
	summary: "lists the connected players",
	handler: func(conn *Connection, args ...string) {
		for other, _ := range conn.game.connections {
			conn.Printf("%v\n", other.Name())
		}
	},
}

var balCommand = Command{
	name:    "bal",
	summary: "displays your current balance in space duckets",
	handler: func(conn *Connection, args ...string) {
		fmt.Fprintln(conn, conn.money)
	},
}
