package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

var commandRegistry map[string]*Command

type Command struct {
	name    string
	help    string
	handler func(*Connection, ...string)
	mobile  bool
}

var infoCommand = &Command{
	name: "info",
	help: "gives you some info about your current position",
	handler: func(conn *Connection, args ...string) {
		fmt.Fprintf(conn, "current planet: %s\n", conn.System().name)
	},
}

var nearbyCommand = &Command{
	name: "nearby",
	help: "list objects nearby",
	handler: func(conn *Connection, args ...string) {
		system := conn.System()
		neighbors, err := system.Nearby(25)
		if err != nil {
			log_error("unable to get neighbors: %v", err)
			return
		}
		for _, neighbor := range neighbors {
			other := index[neighbor.id]
			fmt.Fprintf(conn, "%s: %v\n", other.name, neighbor.distance)
		}
	},
}

var helpCommand = &Command{
	name: "help",
	help: "helpful things to help you",
	handler: func(conn *Connection, args ...string) {
		if len(args) == 0 {
			fmt.Fprintln(conn, `use the "commands" command for a list of commands.`)
			fmt.Fprintln(conn, `use "help [command-name]" to get info for a specific command.`)
			return
		}
		for _, cmdName := range args {
			cmd, ok := commandRegistry[cmdName]
			if !ok {
				fmt.Fprintf(conn, "no such command: %v\n", cmdName)
				continue
			}
			fmt.Fprintf(conn, "%v: %v\n", cmdName, cmd.help)
		}
	},
}

var commandsCommand = &Command{
	name: "commands",
	help: "gives you a handy list of commands",
	handler: func(conn *Connection, args ...string) {
		names := make([]string, 0, len(commandRegistry))
		for name, _ := range commandRegistry {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			fmt.Fprintln(conn, name)
		}
	},
}

var scanCommand = &Command{
	name: "scan",
	help: "super duper scan",
	handler: func(conn *Connection, args ...string) {
		if !conn.CanScan() {
			fmt.Fprintf(conn, "scanners are still recharging.  Can scan again in %v\n", conn.NextScan())
			return
		}
		conn.RecordScan()
		system := conn.System()
		log_info("scan sent from %s", system.name)
		for id, _ := range index {
			if id == system.id {
				continue
			}
			delay := system.TimeTo(index[id])
			id2 := id
			After(delay, func() {
				scanSystem(id2, system.id)
			})
		}
	},
}

var broadcastCommand = &Command{
	name: "broadcast",
	help: "broadcast a message for all systems to hear",
	handler: func(conn *Connection, args ...string) {
		msg := strings.Join(args, " ")
		system := conn.System()
		log_info("broadcast sent from %s: %v\n", system.name, msg)
		for id, _ := range index {
			if id == system.id {
				continue
			}
			delay := system.TimeTo(index[id])
			id2 := id
			After(delay, func() {
				deliverMessage(id2, system.id, msg)
			})
		}
	},
}

var gotoCommand = &Command{
	name: "goto",
	help: "moves to a different system, specified by either name or ID",
	handler: func(conn *Connection, args ...string) {
		dest_name := strings.Join(args, " ")
		to, ok := nameIndex[dest_name]
		if ok {
			move(conn, to)
			return
		}

		id_n, err := strconv.Atoi(dest_name)
		if err != nil {
			fmt.Fprintf(conn, `hmm, I don't know a system by the name "%s", try something else`, dest_name)
			return
		}

		to, ok = index[id_n]
		if !ok {
			fmt.Fprintf(conn, `oh dear, there doesn't seem to be a system with id %d`, id_n)
			return
		}
		move(conn, to)
	},
}

func move(conn *Connection, to *System) {
	start := conn.System()
	start.Leave(conn)

	delay := start.TimeTo(to)
	delay = time.Duration(int64(float64(delay/time.Nanosecond) * 1.25))
	After(delay, func() {
		to.Arrive(conn)
		fmt.Fprintf(conn, "You have arrived at the %s system.\n", to.name)
	})
}

// var bombCommand = &Command{
//     name: "bomb",
//     help: "bombs a system, with a big space bomb",
//     handler: func(conn *Connection, args ...string) {
//
//     },
// }

func isCommand(name string) bool {
	_, ok := commandRegistry[name]
	return ok
}

func runCommand(conn *Connection, name string, args ...string) {
	cmd, ok := commandRegistry[name]
	if !ok {
		fmt.Fprintf(conn, "no such command: %s\n", name)
		return
	}

	if conn.InTransit() && !cmd.mobile {
		fmt.Fprintf(conn, "command %s can not be used while in transit", name)
		return
	}
	cmd.handler(conn, args...)
}

func registerCommand(c *Command) {
	commandRegistry[c.name] = c
}

func init() {
	commandRegistry = make(map[string]*Command, 16)
	registerCommand(commandsCommand)
	registerCommand(helpCommand)
	registerCommand(infoCommand)
	registerCommand(nearbyCommand)
	registerCommand(scanCommand)
	registerCommand(gotoCommand)
	registerCommand(broadcastCommand)
	// registerCommand(bombCommand)
}
