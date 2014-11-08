package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type Connection struct {
	net.Conn
	*bufio.Reader
	player Player
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		Conn:   conn,
		Reader: bufio.NewReader(conn),
	}
}

func (c *Connection) Login() {
	for {
		fmt.Fprintf(c, "what is your name, adventurer?\n")
		name, err := c.ReadString('\n')
		if err == nil {
			name = strings.TrimSpace(name)
		} else {
			log_error("player failed to connect: %v", err)
			return
		}
		if !ValidName(name) {
			fmt.Fprintf(c, "that name is illegal.\n")
			continue
		}
		log_info("player connected: %v", name)
		c.player = Player{name: name}
		break
	}
}

func (c *Connection) Close() error {
	log_info("player disconnecting: %s", c.player.name)
	return c.Conn.Close()
}

func (c *Connection) PlayerName() string {
	return c.player.name
}
