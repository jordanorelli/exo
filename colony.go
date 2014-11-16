package main

import ()

func MakeColony(c *Connection, sys *System) {
	if c.money < options.colonyCost {
		c.Printf("Not enough money!  Colonies cost %v but you only have %v space duckets.  Mine more space duckets!\n", options.colonyCost, c.money)
		return
	}
	if sys.colonizedBy == c {
		c.Printf("You've already colonized this system.\n")
		return
	}
	c.money -= options.colonyCost
	m := &MakeColonyState{
		System: sys,
		CommandSuite: CommandSet{
			balCommand,
			BroadcastCommand(sys),
			helpCommand,
			NearbyCommand(sys),
			playersCommand,
		},
	}
	c.SetState(m)
}

type MakeColonyState struct {
	CommandSuite
	*System
	start int64
}

func (m *MakeColonyState) Enter(c *Connection) {
	c.Printf("Making colony on %v...\n", m.System)
}

func (m *MakeColonyState) Tick(c *Connection, frame int64) ConnectionState {
	if m.start == 0 {
		m.start = frame
	}
	if framesToDur(frame-m.start) >= options.makeColonyTime {
		return Idle(m.System)
	}
	return m
}

func (m *MakeColonyState) Exit(c *Connection) {
	m.System.colonizedBy = c
	c.Printf("Established colony on %v.\n", m.System)
}
