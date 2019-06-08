package main

import (
	"fmt"
	"strings"
	"time"
)

type Bomb struct {
	profile *Connection
	origin  *System
	target  *System
	start   time.Time
	done    bool
	fti     int64 // frames to impact
}

func NewBomb(conn *Connection, from, to *System) *Bomb {
	dist := from.DistanceTo(to)
	fti := int64(dist / (options.lightSpeed * options.bombSpeed))
	eta := time.Duration(fti) * time.Second / time.Duration(options.frameRate)
	log_info("bomb from: %v to: %v ETA: %v", from, to, eta)
	return &Bomb{
		profile: conn,
		origin:  from,
		target:  to,
		fti:     fti,
		start:   time.Now(),
	}
}

func (b *Bomb) Dead() bool {
	return b.done
}

func (b *Bomb) Tick(game *Game) {
	b.fti -= 1
	if b.fti <= 0 {
		b.target.Bombed(b.profile, game.frame)
		b.done = true
		log_info("bomb went off on %v", b.target)
	}
}

func (b *Bomb) String() string {
	return fmt.Sprintf("[bomb from: %v to: %v lived: %s]", b.origin, b.target, time.Since(b.start))
}

type MakeBombState struct {
	CommandSuite
	*System
	start int64
}

func MakeBomb(s *System) ConnectionState {
	m := &MakeBombState{System: s}
	m.CommandSuite = CommandSet{
		balCommand,
		BroadcastCommand(s),
		NearbyCommand(s),
		playersCommand,
	}
	return m
}

func (m *MakeBombState) Enter(c *Connection) {
	c.Printf("Making a bomb...\n")
	c.money -= options.bombCost
}

func (m *MakeBombState) Tick(c *Connection, frame int64) ConnectionState {
	if m.start == 0 {
		m.start = frame
	}
	if framesToDur(frame-m.start) >= options.makeBombTime {
		return Idle(m.System)
	}
	return m
}

func (MakeBombState) String() string { return "Making a Bomb" }

func (m *MakeBombState) Exit(c *Connection) {
	c.bombs += 1
	c.Printf("Done!  You now have %v bombs.\n", c.bombs)
}

func (m *MakeBombState) FillStatus(c *Connection, s *status) {
	elapsedFrames := c.game.frame - m.start
	elapsedDur := framesToDur(elapsedFrames)

	desc := fmt.Sprintf(`
Currently making a bomb!

Build time elapsed:   %v
Build time remaining: %v
`, elapsedDur, options.makeBombTime-elapsedDur)
	s.Description = strings.TrimSpace(desc)
	s.Location = m.System.String()
}
