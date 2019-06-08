package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"runtime"
	"sort"
	"strings"
	"time"
)

type Connection struct {
	*bufio.Reader
	game *Game
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
		bombs:  options.startBombs,
		money:  options.startMoney,
	}
	c.SetState(EnterLobby())
	return c
}

func (c *Connection) Dead() bool {
	return false
}

func (c *Connection) Tick(game *Game) {
	if c.ConnectionState == nil {
		log_error("connected client has nil state.")
		c.Printf("somehow you have a nil state.  I don't know what to do so I'm going to kick you off.")
		c.Close()
		return
	}
	c.SetState(c.ConnectionState.Tick(c, game.frame))
}

func (c *Connection) RunCommand(name string, args ...string) {
	defer func() {
		if r := recover(); r != nil {
			c.Printf("(something is broken)")
			c.Printf("ERROR: %v\n", r)
			callers := make([]uintptr, 40)
			n := runtime.Callers(5, callers)
			callers = callers[:n]
			frames := runtime.CallersFrames(callers)
			log_error("recovered: %v", r)
			for {
				frame, more := frames.Next()

				if !more {
					break
				}

				log_error("  %s +%d (%s)\n", frame.File, frame.Line, frame.Function)
			}
		}
	}()
	switch name {
	case "commands":
		c.ListCommands()
		return
	}

	cmd := c.GetCommand(name)
	if cmd == nil {
		c.Printf("No such command: %v\n", name)
		return
	}
	cmd.handler(c, args...)
}

func (c *Connection) ListCommands() {
	c.Printf("\n")
	c.Line()
	c.Printf("- Available Commands in state: %s\n", c.ConnectionState.String())
	c.Line()
	commands := c.Commands()
	names := make([]string, len(commands))
	for i := range commands {
		names[i] = commands[i].name
	}
	sort.Strings(names)
	for _, name := range names {
		cmd := c.GetCommand(name)
		c.Printf("%-20s%s\n", name, cmd.summary)
	}
	c.Printf("\n")
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
	c.ConnectionState = s
	s.Enter(c)
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

func (c *Connection) Line() {
	c.Printf("--------------------------------------------------------------------------------\n")
}

func (c *Connection) Printf(template string, args ...interface{}) (int, error) {
	return fmt.Fprintf(c, template, args...)
}

func (c *Connection) Close() error {
	log_info("player disconnecting: %s", c.Name())
	if c.game != nil {
		c.game.Quit(c)
	}
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
	c.Printf("Scanning known systems for signs of life\n")
	c.lastScan = time.Now()
	time.AfterFunc(options.scanTime, func() {
		c.Printf("Scanner ready\n")
	})
}

func (c *Connection) RecordBomb() {
	c.lastBomb = time.Now()
	time.AfterFunc(15*time.Second, func() {
		fmt.Fprintln(c, "Bomb arsenal reloaded")
	})
}

func (c *Connection) CanScan() bool {
	return time.Since(c.lastScan) > options.scanTime
}

func (c *Connection) NextScan() time.Duration {
	return -time.Since(c.lastScan.Add(options.scanTime))
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
	c.game.Win(c, method)
}

func (c *Connection) Die(frame int64) {
	c.SetState(NewDeadState(frame))
}
