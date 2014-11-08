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
	player   *Player
	location *System
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
		player, err := loadPlayer(name)
		if err != nil {
			log_error("could not read player: %v", err)
			player = &Player{name: name}
			if err := player.Create(); err != nil {

			}
			fmt.Fprintf(c, "godspeed, %s.\n", player.name)
		} else {
			c.player = player
			fmt.Fprintf(c, "welcome back, %s.\n", player.name)
		}
		break
	}

}

func (c *Connection) SetSystem(s *System) {
	c.location = s
}

func (c *Connection) System() *System {
	return c.location
}

func (c *Connection) Close() error {
	log_info("player disconnecting: %s", c.PlayerName())
	return c.Conn.Close()
}

func (c *Connection) PlayerName() string {
	if c.player == nil {
		return ""
	}
	return c.player.name
}
