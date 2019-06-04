package main

import (
	"fmt"
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
			help:    "stops mining",
			arity:   0,
			handler: m.stop,
		},
		Command{
			name:    "info",
			help:    "gives you information about the current mining operation",
			arity:   0,
			handler: m.info,
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

func (m *MiningState) info(c *Connection, args ...string) {
	c.Printf("Currently mining system %v\n", m.System)
	c.Printf("Mined so far: %v\n", m.mined)
	c.Printf("Remaining space duckets on %v: %v\n", m.System, m.money)
}

func (m *MiningState) PrintStatus(c *Connection) {
	panic("not done")
}
