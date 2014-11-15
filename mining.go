package main

import (
	"fmt"
)

type MiningState struct {
	CommandSuite
	sys   *System
	mined int
}

func Mine(sys *System) ConnectionState {
	return &MiningState{sys: sys}
}

func (m *MiningState) Enter(c *Connection) {
	c.Printf("Mining %v. %v space duckets remaining.\n", m.sys, m.sys.money)
}

func (m *MiningState) Tick(c *Connection, frame int64) ConnectionState {
	if m.sys.money <= 0 {
		c.Printf("system %s is all out of space duckets.\n", m.sys)
		return Idle(m.sys)
	} else {
		c.Deposit(1)
		m.mined += 1
		m.sys.money -= 1
		return m
	}
}

func (m *MiningState) Exit(c *Connection) {
	if m.sys.money == 0 {
		c.Printf("Done mining %v.  Mined %v space duckets total.  %v space duckets remain on %v, and it can be mined again.", m.sys, m.mined, m.sys.money, m.sys)
	} else {
		c.Printf("Done mining %v.  Mined %v space duckets total.  No space duckets remain on %v, and it can't be mined again.", m.sys, m.mined, m.sys)
	}
}

func (m *MiningState) String() string {
	return fmt.Sprintf("mining %v", m.sys)
}
