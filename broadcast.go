package main

import (
	"fmt"
	"time"
)

type broadcast struct {
	NopReset
	start        time.Time
	origin       *System
	dist         float64
	nextHitIndex int
	message      string
}

func NewBroadcast(from *System, template string, args ...interface{}) *broadcast {
	return &broadcast{
		start:   time.Now(),
		origin:  from,
		message: fmt.Sprintf(template, args...),
	}
}

func (b *broadcast) Tick(frame int64) {
	b.dist += options.lightSpeed
	for ; b.nextHitIndex < len(b.origin.Distances()); b.nextHitIndex += 1 {
		candidate := b.origin.Distances()[b.nextHitIndex]
		if b.dist < candidate.dist {
			break
		}
		candidate.s.NotifyInhabitants("message received from system %s:\n\t%s\n", b.origin.Label(), b.message)
	}
}

func (b *broadcast) Dead() bool {
	return b.dist > b.origin.Distances()[len(b.origin.Distances())-1].dist
}

func (b *broadcast) String() string {
	return fmt.Sprintf("[broadcast origin: %s message: %s]", b.origin.name, b.message)
}
