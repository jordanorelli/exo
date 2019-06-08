package main

import (
	"math/rand"
	"sort"
	"strconv"
)

// Galaxy is a collection of systems
type Galaxy struct {
	systems map[int]*System
	names   map[string]int
}

func NewGalaxy() *Galaxy {
	g := &Galaxy{
		systems: make(map[int]*System),
		names:   make(map[string]int),
	}
	g.indexSystems()
	return g
}

func (g *Galaxy) indexSystems() {
	rows, err := db.Query(`select * from planets`)
	if err != nil {
		log_error("unable to select all planets: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		s := System{}
		if err := rows.Scan(&s.id, &s.name, &s.x, &s.y, &s.z, &s.planets); err != nil {
			log_info("unable to scan planet row: %v", err)
			continue
		}
		g.systems[s.id] = &s
		g.names[s.name] = s.id
		s.money = int64(rand.NormFloat64()*options.moneySigma + options.moneyMean)
	}
}

// GetSystem gets a system by either ID or name. If the provided string
// contains an integer, we assume the lookup is intended to be by ID.
func (g *Galaxy) GetSystem(s string) *System {
	id, err := strconv.Atoi(s)
	if err == nil {
		return g.GetSystemByID(id)
	}

	return g.GetSystemByName(s)
}

func (g *Galaxy) GetSystemByID(id int) *System {
	return g.systems[id]
}

func (g *Galaxy) GetSystemByName(name string) *System {
	id := g.SystemID(name)
	if id == 0 {
		return nil
	}
	return g.GetSystemByID(id)
}

func (g *Galaxy) SystemID(name string) int { return g.names[name] }

// Neighborhood generates the neighborhood for a given system.
func (g *Galaxy) Neighborhood(sys *System) Neighborhood {
	neighbors := make(Neighborhood, 0, len(g.systems))
	for id, sys2 := range g.systems {
		if id == sys.id {
			continue
		}
		neighbors = append(neighbors, Neighbor{id: id, distance: sys.DistanceTo(sys2)})
	}
	sort.Sort(neighbors)
	return neighbors
}

func (g *Galaxy) randomSystem() *System {
	id := rand.Intn(len(g.systems))
	return g.GetSystemByID(id)
}
