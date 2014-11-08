package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

var connected = make(map[*Connection]bool, 32)

type Connection struct {
	net.Conn
	*bufio.Reader
	player   *Player
	location *System
	lastScan time.Time
	lastBomb time.Time
	kills    int
	dead     bool
}

func NewConnection(conn net.Conn) *Connection {
	c := &Connection{
		Conn:   conn,
		Reader: bufio.NewReader(conn),
	}
	connected[c] = true
	return c
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
	delete(connected, c)
	return c.Conn.Close()
}

func (c *Connection) PlayerName() string {
	if c.player == nil {
		return ""
	}
	return c.player.name
}

func (c *Connection) InTransit() bool {
	return c.location == nil
}

func (c *Connection) RecordScan() {
	c.lastScan = time.Now()
}

func (c *Connection) RecordBomb() {
	c.lastBomb = time.Now()
}

func (c *Connection) CanScan() bool {
	return time.Since(c.lastScan) > 1*time.Minute
}

func (c *Connection) CanBomb() bool {
	return time.Since(c.lastBomb) > 1500*time.Millisecond
}

func (c *Connection) NextScan() time.Duration {
	return -time.Since(c.lastScan.Add(time.Minute))
}

func (c *Connection) NextBomb() time.Duration {
	return -time.Since(c.lastBomb.Add(1500 * time.Millisecond))
}

func (c *Connection) MadeKill(victim *Connection) {
	c.kills += 1
	if c.kills == 3 {
		c.Win()
	}
}

func (c *Connection) Win() {
	for conn, _ := range connected {
		fmt.Fprintf(conn, "player %s has won.\n", c.PlayerName())
		conn.Close()
	}
}

func (c *Connection) Die() {
	fmt.Fprintf(c, "you were bombed.  You will respawn in 2 minutes.\n")
	c.dead = true
	c.System().Leave(c)
	After(30*time.Second, func() {
		fmt.Fprintf(c, "respawn in 90 seconds.\n")
	})
	After(time.Minute, func() {
		fmt.Fprintf(c, "respawn in 60 seconds.\n")
	})
	After(90*time.Second, func() {
		fmt.Fprintf(c, "respawn in 30 seconds.\n")
	})
	After(2*time.Minute, c.Respawn)
}

func (c *Connection) Respawn() {
	c.dead = false

WUT:
	s, err := randomSystem()
	if err != nil {
		log_error("error in respawn: %v", err)
		goto WUT
	}
	s.Arrive(c)

}
