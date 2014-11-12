package main

import (
	"time"
)

type scan struct {
	Mortality
	start        time.Time
	origin       *System
	dist         float64
	nextHitIndex int
}

func (s *scan) Tick(frame int64) {
	s.dist += lightSpeed
	for {
		candidate := s.origin.Distances()[s.nextHitIndex]
		if s.dist < candidate.dist {
			break
		}
		log_info("scan hit %v. Traveled %v in %v", candidate.s.name, candidate.dist, time.Since(s.start))
		s.nextHitIndex += 1
		if s.nextHitIndex >= len(s.origin.Distances()) {
			s.Mortality = true
			log_info("scan complete")
			break
		}
	}
}
