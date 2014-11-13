package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

type Connection struct {
	net.Conn
	*bufio.Reader
	player          *Player
	location        *System
	dest            *System
	travelRemaining int64
	lastScan        time.Time
	lastBomb        time.Time
	kills           int
	dead            bool
	money           int
	colonies        []*System
	bombs           int
	state           PlayerState // this is wrong...
}

type PlayerState int

const (
	idle PlayerState = iota
	dead
	inTransit
	mining
)

func NewConnection(conn net.Conn) *Connection {
	c := &Connection{
		Conn:   conn,
		Reader: bufio.NewReader(conn),
		bombs:  1,
	}
	currentGame.Join(c)
	return c
}

func (conn *Connection) Reset() {
	*conn = Connection{
		Conn:   conn.Conn,
		Reader: bufio.NewReader(conn.Conn),
		bombs:  1,
		player: conn.player,
	}
	currentGame.Join(conn)
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
				log_error("unable to create player record: %v", err)
			}
			fmt.Fprintf(c, "you look new around these parts, %s.\n", player.name)
			fmt.Fprintf(c, `if you'd like a description of how to play, type the "help" command`)
			c.player = player
		} else {
			c.player = player
			fmt.Fprintf(c, "welcome back, %s.\n", player.name)
		}
		break
	}
	currentGame.Register(c)
}

func (c *Connection) Dead() bool {
	return false
}

func (c *Connection) Tick(frame int64) {
	// fuck

	switch c.state {
	case idle:
	case dead:
	case inTransit:
		c.travelRemaining -= 1
		log_info("player %s has remaining travel: %v", c.PlayerName(), c.travelRemaining)
		if c.travelRemaining == 0 {
			c.land()
		}
	case mining:
		c.Deposit(options.miningRate)
		log_info("%v", c.money)
	default:
		log_error("connection %v has invalid state wtf", c)
	}
}

func (c *Connection) TravelTo(dest *System) {
	fmt.Fprintf(c, "traveling to: %s\n", dest.Label())
	dist := c.System().DistanceTo(dest)
	c.travelRemaining = int64(dist / (options.lightSpeed * options.playerSpeed))
	c.location = nil
	c.dest = dest
	c.state = inTransit // fuck everything about this
}

func (c *Connection) land() {
	fmt.Fprintf(c, "you have arrived at %v\n", c.dest.Label())
	c.location = c.dest
	c.dest = nil
	c.state = idle
}

func (c *Connection) SetSystem(s *System) {
	c.location = s
}

func (c *Connection) System() *System {
	return c.location
}

func (c *Connection) Close() error {
	log_info("player disconnecting: %s", c.PlayerName())
	currentGame.Quit(c)
	if c.Conn != nil {
		return c.Conn.Close()
	}
	return nil
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
	fmt.Fprintln(c, "scanning known systems for signs of life")
	c.lastScan = time.Now()
	time.AfterFunc(1*time.Minute, func() {
		fmt.Fprintln(c, "scanner ready")
	})
}

func (c *Connection) RecordBomb() {
	c.lastBomb = time.Now()
	time.AfterFunc(15*time.Second, func() {
		fmt.Fprintln(c, "bomb arsenal reloaded")
	})
}

func (c *Connection) CanScan() bool {
	return time.Since(c.lastScan) > 1*time.Minute
}

func (c *Connection) CanBomb() bool {
	return time.Since(c.lastBomb) > 15*time.Second
}

func (c *Connection) NextScan() time.Duration {
	return -time.Since(c.lastScan.Add(time.Minute))
}

func (c *Connection) NextBomb() time.Duration {
	return -time.Since(c.lastBomb.Add(15 * time.Second))
}

func (c *Connection) MadeKill(victim *Connection) {
	c.kills += 1
	if c.kills == 3 {
		c.Win("military")
	}
}

func (c *Connection) Mine() {
	switch c.state {
	case idle:
		fmt.Fprintf(c, "now mining %s with a payout rate of %v\n", c.System().name, c.System().miningRate)
		fmt.Fprintln(c, "(press enter to stop mining)")
		c.state = mining
	default:
		fmt.Fprintf(c, "no\n")
	}
}

func (c *Connection) StopMining() {
	fmt.Fprintf(c, "done mining\n")
	c.state = idle
}

func (c *Connection) IsMining() bool {
	return c.state == mining
}

func (c *Connection) Withdraw(n int) {
	c.money -= n
}

func (c *Connection) Deposit(n int) {
	c.money += n
	if c.money >= options.economic {
		c.Win("economic")
	}
}

func (c *Connection) Win(method string) {
	currentGame.Win(c, method)
}

func (c *Connection) Die() {
	fmt.Fprintf(c, "you were bombed.  You will respawn in 1 minutes.\n")
	c.dead = true
	c.System().Leave(c)
	time.AfterFunc(30*time.Second, func() {
		fmt.Fprintf(c, "respawn in 30 seconds.\n")
	})
	time.AfterFunc(time.Minute, c.Respawn)
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
