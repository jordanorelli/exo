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
	done   bool
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
	return b.done
}

func (b *Bomb) Tick(frame int64) {
	b.fti -= 1
	if b.fti <= 0 {
		b.target.Bombed(b.player)
		b.done = true
		log_info("bomb went off on %s", b.target.Label())
	}
}

func (b *Bomb) String() string {
	return fmt.Sprintf("[bomb from: %s to: %s lived: %s]", b.origin.Label(), b.target.Label(), time.Since(b.start))
}
