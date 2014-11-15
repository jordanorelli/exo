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

func (r *scanResult) Empty() bool {
	return (r.players == nil || len(r.players) == 0) && r.colonizedBy == nil
}

func (r *scanResult) playerNames() []string {
	if r.players == nil || len(r.players) == 0 {
		return nil
	}
	names := make([]string, 0, len(r.players))
	for conn := range r.players {
		names = append(names, conn.Name())
	}
	return names
}

func NewScan(origin *System) *scan {
	return &scan{
		origin:  origin,
		start:   time.Now(),
		results: make([]scanResult, 0, len(origin.Distances())),
	}
}

func (s *scan) Tick(frame int64) {
	s.dist += options.lightSpeed
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
	sys.NotifyInhabitants("scan detected from %v\n", s.origin)
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
		if res.Empty() {
			continue
		}
		s.origin.NotifyInhabitants("results from scan of %v:\n", res.system)
		s.origin.NotifyInhabitants("\tdistance: %v\n", s.origin.DistanceTo(res.system))
		inhabitants := res.playerNames()
		if inhabitants != nil {
			s.origin.NotifyInhabitants("\tinhabitants: %v\n", inhabitants)
		}
		if res.colonizedBy != nil {
			s.origin.NotifyInhabitants("\tcolonized by: %v\n", res.colonizedBy.Name())
		}
	}
}
