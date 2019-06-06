package main

import (
	"fmt"
)

func MakeShield(c *Connection, s *System) {
	m := &MakeShieldState{
		System: s,
		CommandSuite: CommandSet{
			balCommand,
			BroadcastCommand(s),
			NearbyCommand(s),
			playersCommand,
		},
	}
	c.SetState(m)
}

type MakeShieldState struct {
	CommandSuite
	*System
	start int64
}

func (m *MakeShieldState) Enter(c *Connection) {
	c.Printf("Making shield on %v...\n", m.System)
}

func (m *MakeShieldState) Tick(c *Connection, frame int64) ConnectionState {
	if m.start == 0 {
		m.start = frame
	}
	if framesToDur(frame-m.start) >= options.makeShieldTime {
		return Idle(m.System)
	}
	return m
}

func (m *MakeShieldState) Exit(c *Connection) {
	c.Printf("Done!  System %v is now shielded.\n", m.System)
	m.System.Shield = new(Shield)
}

func (m *MakeShieldState) String() string {
	return fmt.Sprintf("Making shield on %v", m.System)
}

func (m *MakeShieldState) FillStatus(c *Connection, s *status) {
	s.Location = m.System.String()
}

type Shield struct {
	energy float64
}

func (s *Shield) Tick(frame int64) {
	if s.energy < 1000 {
		s.energy += (1000 - s.energy) * 0.0005
	}
}

func (s *Shield) Hit() bool {
	if s.energy > 750 {
		s.energy -= 750
		return true
	}
	return false
}

func (s *Shield) Dead() bool {
	return false
}
