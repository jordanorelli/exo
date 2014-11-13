package main

import (
	"fmt"
	"time"
)

type Bomb struct {
	player *Connection
	origin *System
	target *System
	start  time.Time
	fti    int64 // frames to impact
}

func NewBomb(from *Connection, to *System) *Bomb {
	origin := from.System()
	dist := origin.DistanceTo(to)
	fti := int64(dist / (options.lightSpeed * options.bombSpeed))
	eta := time.Duration(fti) * time.Second / time.Duration(options.frameRate)
	log_info("bomb from: %s to: %s ETA: %v", from.PlayerName(), to.Label(), eta)
	return &Bomb{
		player: from,
		origin: origin,
		target: to,
		fti:    fti,
		start:  time.Now(),
	}
}

func (b *Bomb) Dead() bool {
	return b.fti <= 0
}

func (b *Bomb) Tick(frame int64) {
	b.fti -= 1
	if b.fti <= 0 {
		b.target.Bombed(b.player)
	}
}

func (b *Bomb) String() string {
	return fmt.Sprintf("[bomb from: %s to: %s lived: %s]", b.origin.Label(), b.target.Label(), time.Since(b.start))
}
