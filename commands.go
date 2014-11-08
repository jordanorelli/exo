package main

import (
	"fmt"
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
	registerCommand(infoCommand)
}
