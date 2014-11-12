package main

import (
	"fmt"
	"time"
)

type scan struct {
	start         time.Time
	origin        *System
	dist          float64
	nextHitIndex  int
	nextEchoIndex int
	results       []scanResult
}

type scanResult struct {
	system      *System
	dist        float64
	players     map[*Connection]bool
	colonizedBy *Connection
}

func NewScan(origin *System) *scan {
	return &scan{
		origin:  origin,
		start:   time.Now(),
		results: make([]scanResult, 0, len(origin.Distances())),
	}
}

func (s *scan) Tick(frame int64) {
	s.dist += lightSpeed
	s.hits()
	s.echos()
}

func (s *scan) Dead() bool {
	return s.nextEchoIndex >= len(s.origin.Distances())
}

func (s *scan) String() string {
	return fmt.Sprintf("[scan origin: %s start_time: %v]", s.origin.name, s.start)
}

func (s *scan) hits() {
	for ; s.nextHitIndex < len(s.origin.Distances()); s.nextHitIndex += 1 {
		candidate := s.origin.Distances()[s.nextHitIndex]
		if s.dist < candidate.dist {
			break
		}
		s.results = append(s.results, s.hitSystem(candidate.s, candidate.dist))
		log_info("scan hit %v. Traveled %v in %v", candidate.s.name, candidate.dist, time.Since(s.start))
	}
}

func (s *scan) hitSystem(sys *System, dist float64) scanResult {
	r := scanResult{
		system:      sys,
		colonizedBy: sys.colonizedBy,
		dist:        dist * 2.0,
	}
	if sys.players != nil {
		r.players = make(map[*Connection]bool, len(sys.players))
		for k, v := range sys.players {
			r.players[k] = v
		}
	}
	return r
}

func (s *scan) echos() {
	for ; s.nextEchoIndex < len(s.results); s.nextEchoIndex += 1 {
		res := s.results[s.nextEchoIndex]
		if s.dist < res.dist {
			break
		}
		log_info("echo from %v reached origin %v after %v", res.system.name, s.origin.name, time.Since(s.start))
	}
}
