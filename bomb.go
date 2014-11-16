package main

import (
	"fmt"
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

func (b *Bomb) Tick(frame int64) {
	b.fti -= 1
	if b.fti <= 0 {
		b.target.Bombed(b.profile, frame)
		b.done = true
		log_info("bomb went off on %v", b.target)
	}
}

func (b *Bomb) String() string {
	return fmt.Sprintf("[bomb from: %v to: %v lived: %s]", b.origin, b.target, time.Since(b.start))
}
