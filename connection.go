package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

type Connection struct {
	net.Conn
	*bufio.Reader
	profile         *Profile
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

func (c *Connection) Login() {
	for {
		c.Printf("what is your name, adventurer?\n")
		name, err := c.ReadString('\n')
		if err == nil {
			name = strings.TrimSpace(name)
		} else {
			log_error("player failed to connect: %v", err)
			return
		}
		if !ValidName(name) {
			c.Printf("that name is illegal.\n")
			continue
		}
		log_info("player connected: %v", name)
		profile, err := loadProfile(name)
		if err != nil {
			log_error("could not read profile: %v", err)
			profile = &Profile{name: name}
			if err := profile.Create(); err != nil {
				log_error("unable to create profile record: %v", err)
			}
			c.Printf("you look new around these parts, %s.\n", profile.name)
			c.Printf(`if you'd like a description of how to play, type the "help" command\n`)
			c.profile = profile
		} else {
			c.profile = profile
			c.Printf("welcome back, %s.\n", profile.name)
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
		if c.travelRemaining == 0 {
			c.land()
		}
	case mining:
		sys := c.System()
		if sys == nil {
			log_error("a player is in the mining state with no system. what?")
			break
		}
		if sys.money <= 0 {
			c.Printf("system %s is all out of space duckets.\n", sys.Label())
			c.StopMining()
		} else {
			c.Deposit(1)
			sys.money -= 1
		}
	default:
		log_error("connection %v has invalid state wtf", c)
	}
}

func (c *Connection) TravelTo(dest *System) {
	dist := c.System().DistanceTo(dest)
	c.travelRemaining = int64(dist / (options.lightSpeed * options.playerSpeed))
	t := time.Duration(c.travelRemaining) * (time.Second / time.Duration(options.frameRate))
	c.Printf("traveling to: %s. ETA: %v\n", dest.Label(), t)
	c.location.Leave(c)
	c.location = nil
	c.dest = dest
	c.state = inTransit // fuck everything about this
}

func (c *Connection) SendBomb(target *System) {
	if c.bombs <= 0 {
		fmt.Fprintln(c, "cannot send bomb: no bombs left")
		return
	}
	if time.Since(c.lastBomb) < 5*time.Second {
		fmt.Fprintln(c, "cannod send bomb: bombs are reloading")
		return
	}
	c.bombs -= 1
	c.lastBomb = time.Now()
	bomb := NewBomb(c, target)
	currentGame.Register(bomb)
	c.Printf("sending bomb to system %v\n", target.Label())
}

func (c *Connection) ReadLines(out chan []string) {
	defer close(out)

	for {
		line, err := c.ReadString('\n')
		switch err {
		case io.EOF:
			return
		case nil:
			break
		default:
			log_error("unable to read line on connection: %v", err)
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out <- strings.Split(line, " ")
	}
}

func (c *Connection) Printf(template string, args ...interface{}) (int, error) {
	return fmt.Fprintf(c, template, args...)
}

func (c *Connection) land() {
	c.Printf("you have arrived at %v\n", c.dest.Label())
	c.location = c.dest
	c.location.Arrive(c)
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
	log_info("player disconnecting: %s", c.Name())
	currentGame.Quit(c)
	if c.Conn != nil {
		return c.Conn.Close()
	}
	return nil
}

func (c *Connection) Name() string {
	if c.profile == nil {
		return ""
	}
	return c.profile.name
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

func (c *Connection) NextScan() time.Duration {
	return -time.Since(c.lastScan.Add(time.Minute))
}

func (c *Connection) NextBomb() time.Duration {
	return -time.Since(c.lastBomb.Add(15 * time.Second))
}

func (c *Connection) MadeKill(victim *Connection) {
	if c == victim {
		log_info("player %s commited suicide.", c.Name())
		return
	}
	c.kills += 1
	if c.kills == 3 {
		c.Win("military")
	}
}

func (c *Connection) Mine() {
	switch c.state {
	case idle:
		c.Printf("now mining %s. %v space duckets remaining.\n", c.System().name, c.System().money)
		fmt.Fprintln(c, "(press enter to stop mining)")
		c.state = mining
	default:
		c.Printf("no\n")
	}
}

func (c *Connection) StopMining() {
	c.Printf("done mining\n")
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
	c.Printf("you were bombed.  You will respawn in 1 minutes.\n")
	c.dead = true
	c.System().Leave(c)
	time.AfterFunc(30*time.Second, func() {
		c.Printf("respawn in 30 seconds.\n")
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
