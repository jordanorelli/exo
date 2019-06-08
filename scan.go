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
	neighborhood  Neighborhood
}

type scanResult struct {
	system       *System
	dist         float64
	players      map[*Connection]bool
	colonizedBy  *Connection
	shielded     bool
	shieldEnergy float64
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

func NewScan(origin *System, n Neighborhood) *scan {
	return &scan{
		origin:       origin,
		start:        time.Now(),
		results:      make([]scanResult, 0, len(origin.Distances())),
		neighborhood: n,
	}
}

func (s *scan) Tick(game *Game) {
	s.dist += options.lightSpeed
	s.hits(game)
	s.echos()
}

func (s *scan) Dead() bool {
	return s.nextEchoIndex >= len(s.origin.Distances())
}

func (s *scan) String() string {
	return fmt.Sprintf("[scan origin: %s start_time: %v]", s.origin.name, s.start)
}

func (s *scan) hits(game *Game) {
	for len(s.neighborhood) > 0 && s.neighborhood[0].distance <= s.dist {
		sys := game.galaxy.GetSystemByID(s.neighborhood[0].id)
		s.results = append(s.results, s.hitSystem(sys, s.neighborhood[0].distance))
		log_info("scan hit %v. Traveled %v in %v", sys.name, s.neighborhood[0].distance, time.Since(s.start))

		if len(s.neighborhood) > 1 {
			s.neighborhood = s.neighborhood[1:]
		} else {
			s.neighborhood = nil
		}
	}
}

func (s *scan) hitSystem(sys *System, dist float64) scanResult {
	sys.NotifyInhabitants("scan detected from %v\n", s.origin)
	r := scanResult{
		system:      sys,
		colonizedBy: sys.colonizedBy,
		dist:        dist * 2.0,
		shielded:    sys.Shield != nil,
	}
	if sys.Shield != nil {
		r.shieldEnergy = sys.Shield.energy
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
		s.origin.NotifyInhabitants("\tshielded: %v\n", res.shielded)
		if res.shielded {
			s.origin.NotifyInhabitants("\tshield energy: %v\n", res.shieldEnergy)
		}
		inhabitants := res.playerNames()
		if inhabitants != nil {
			s.origin.NotifyInhabitants("\tinhabitants: %v\n", inhabitants)
		}
		if res.colonizedBy != nil {
			s.origin.NotifyInhabitants("\tcolonized by: %v\n", res.colonizedBy.Name())
		}

	}
}
