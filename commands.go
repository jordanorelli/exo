package main

import (
	"fmt"
	"sort"
)

var commandRegistry map[string]*Command

type Command struct {
	name    string
	help    string
	handler func(*Connection, ...string)
}

// var scanCommand = &Command{
// 	name: "scan",
// 	help: "scans for resources",
// 	handler: func(conn *Connection, args ...string) {
// 	},
// }

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

// var superscanCommand = &Command{
// 	name: "super-scan",
// 	help: "super duper scan",
// 	handler: func(conn *Connection, args ...string) {
// 		for id, _ := range index {
//             if id == conn.System().id {
//                 continue
//             }
//
// 		}
// 	},
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
}
