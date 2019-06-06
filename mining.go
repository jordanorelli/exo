package main

import (
	"fmt"
	"strings"
)

type MiningState struct {
	CommandSuite
	*System
	mined int
}

func Mine(sys *System) ConnectionState {
	m := &MiningState{System: sys}
	m.CommandSuite = CommandSet{
		balCommand,
		playersCommand,
		BroadcastCommand(sys),
		NearbyCommand(sys),
		Command{
			name:    "stop",
			summary: "stops mining",
			arity:   0,
			handler: m.stop,
		},
	}
	return m
}

func (m *MiningState) Enter(c *Connection) {
	c.Printf("Mining %v. %v space duckets remaining.\n", m.System, m.money)
}

func (m *MiningState) Tick(c *Connection, frame int64) ConnectionState {
	if m.money <= 0 {
		c.Printf("system %s is all out of space duckets.\n", m.System)
		return Idle(m.System)
	} else {
		c.Deposit(1)
		m.mined += 1
		m.money -= 1
		return m
	}
}

func (m *MiningState) Exit(c *Connection) {
	if m.money == 0 {
		c.Printf("Done mining %v.\nMined %v space duckets total.\nNo space duckets remain on %v, and it can't be mined again.\n", m.System, m.mined, m.System)
	} else {
		c.Printf("Done mining %v.\nMined %v space duckets total.\n%v space duckets remain on %v, and it can be mined again.\n", m.System, m.mined, m.money, m.System)
	}
}

func (m *MiningState) String() string {
	return fmt.Sprintf("mining %v", m.System)
}

func (m *MiningState) stop(c *Connection, args ...string) {
	c.SetState(Idle(m.System))
}

func (m *MiningState) FillStatus(c *Connection, s *status) {
	s.Location = m.System.String()
	s.Description = strings.TrimSpace(fmt.Sprintf(`
Currently mining on system: %s
Mined so far:               %d
Available space duckets:    %d
`, m.System.String(), m.mined, m.money))
}
