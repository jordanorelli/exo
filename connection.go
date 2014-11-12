package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"
)

type Connection struct {
	net.Conn
	*bufio.Reader
	player   *Player
	location *System
	lastScan time.Time
	lastBomb time.Time
	kills    int
	dead     bool
	money    int64
	mining   bool
	colonies []*System
	bombs    int
}

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
	After(1*time.Minute, func() {
		fmt.Fprintln(c, "scanner ready")
	})
}

func (c *Connection) RecordBomb() {
	c.lastBomb = time.Now()
	After(15*time.Second, func() {
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

func (c *Connection) StartMining() {
	fmt.Fprintf(c, "now mining %s with a payout rate of %v\n", c.System().name, c.System().miningRate)
	fmt.Fprintln(c, "(press enter to stop mining)")
	c.mining = true
}

func (c *Connection) StopMining() {
	fmt.Fprintf(c, "done mining\n")
	c.mining = false
}

func (c *Connection) IsMining() bool {
	return c.mining
}

func (c *Connection) Payout() {
	if c.dead {
		return
	}
	reward := int64(rand.NormFloat64()*5.0 + 100.0*c.System().miningRate)
	c.Deposit(reward)
	fmt.Fprintf(c, "mined: %d space duckets. total: %d\n", reward, c.money)
}

func (c *Connection) Withdraw(n int64) {
	c.money -= n
}

func (c *Connection) Deposit(n int64) {
	c.money += n
	if c.money >= 25000 {
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
	After(30*time.Second, func() {
		fmt.Fprintf(c, "respawn in 30 seconds.\n")
	})
	After(time.Minute, c.Respawn)
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
