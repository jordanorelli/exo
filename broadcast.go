package main

import (
	"fmt"
	"time"
)

type broadcast struct {
	start        time.Time
	origin       *System
	dist         float64
	message      string
	neighborhood Neighborhood
}

func NewBroadcast(from *System, template string, args ...interface{}) *broadcast {
	return &broadcast{
		start:   time.Now(),
		origin:  from,
		message: fmt.Sprintf(template, args...),
	}
}

func (b *broadcast) Tick(game *Game) {
	if b.neighborhood == nil {
		log_info("setting up neighborhood for broadcast: %s", b.message)
		b.neighborhood = game.galaxy.Neighborhood(b.origin)
		log_info("nearest neighbor: %v", b.neighborhood[0])
	}

	b.dist += options.lightSpeed
	for len(b.neighborhood) > 0 && b.neighborhood[0].distance <= b.dist {
		s := game.galaxy.GetSystemByID(b.neighborhood[0].id)
		log_info("broadcast %s has reached %s from %s", b.message, s, b.origin)
		s.NotifyInhabitants("message received from system %v:\n\t%s\n", b.origin, b.message)
		if len(b.neighborhood) > 1 {
			b.neighborhood = b.neighborhood[1:]
		} else {
			b.neighborhood = nil
		}
	}
}

func (b *broadcast) Dead() bool { return b.neighborhood == nil }

func (b *broadcast) String() string {
	return fmt.Sprintf("[broadcast origin: %v message: %s]", b.origin, b.message)
}
