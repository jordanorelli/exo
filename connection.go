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
	*bufio.Reader
	net.Conn
	ConnectionState
	bombs    int
	colonies []*System
	kills    int
	lastBomb time.Time
	lastScan time.Time
	money    int
	profile  *Profile
}

func NewConnection(conn net.Conn) *Connection {
	c := &Connection{
		Conn:   conn,
		Reader: bufio.NewReader(conn),
		bombs:  1,
	}
	c.SetState(SpawnRandomly())
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
	if c.ConnectionState == nil {
		log_error("connected client has nil state.")
		c.Printf("somehow you have a nil state.  I don't know what to do so I'm going to kick you off.")
		c.Close()
		return
	}
	c.SetState(c.ConnectionState.Tick(c, frame))
}

func (c *Connection) RunCommand(name string, args ...string) {
	defer func() {
		if r := recover(); r != nil {
			c.Printf("shit is *really* fucked up.\n")
			log_error("recovered: %v", r)
		}
	}()
	switch name {
	case "commands":
		commands := c.Commands()
		for _, command := range commands {
			c.Printf("%s\n", command.name)
		}
		return
	}

	cmd := c.GetCommand(name)
	if cmd == nil {
		c.Printf("No such command: %v\n", name)
		return
	}
	cmd.handler(c, args...)
}

func (c *Connection) SetState(s ConnectionState) {
	if c.ConnectionState == s {
		return
	}
	log_info("set state: %v", s)
	if c.ConnectionState != nil {
		log_info("exit state: %v", c.ConnectionState)
		c.ConnectionState.Exit(c)
	}
	log_info("enter state: %v", s)
	s.Enter(c)
	c.ConnectionState = s
}

// func (c *Connection) SendBomb(target *System) {
// 	if c.bombs <= 0 {
// 		fmt.Fprintln(c, "cannot send bomb: no bombs left")
// 		return
// 	}
// 	if time.Since(c.lastBomb) < 5*time.Second {
// 		fmt.Fprintln(c, "cannod send bomb: bombs are reloading")
// 		return
// 	}
// 	c.bombs -= 1
// 	c.lastBomb = time.Now()
// 	bomb := NewBomb(c, target)
// 	currentGame.Register(bomb)
// 	c.Printf("sending bomb to system %v\n", target)
// }

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

type ConnectionState interface {
	CommandSuite
	String() string
	Enter(c *Connection)
	Tick(c *Connection, frame int64) ConnectionState
	Exit(c *Connection)
}

// No-op enter struct, for composing connection states that have no interesitng
// Enter mechanic.
type NopEnter struct{}

func (n NopEnter) Enter(c *Connection) {}

// No-op exit struct, for composing connection states that have no interesting
// Exit mechanic.
type NopExit struct{}

func (n NopExit) Exit(c *Connection) {}

func SpawnRandomly() ConnectionState {
	sys, err := randomSystem()
	if err != nil {
		return NewErrorState(fmt.Errorf("unable to create idle state: %v", err))
	}
	return Idle(sys)
}
