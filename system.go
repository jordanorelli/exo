package main

import (
	"database/sql"
	"fmt"
	"math"
	"math/rand"
)

var (
	index map[int]*System
)

type System struct {
	id      int
	x, y, z float64
	planets int
	name    string

	players map[*Connection]bool
}

func (s *System) Arrive(p *Connection) {
	p.SetSystem(s)
	log_info("player %s has arrived at system %s", p.PlayerName(), s.name)
	if s.players == nil {
		s.players = make(map[*Connection]bool, 8)
	}
	s.players[p] = true
}

func (s *System) Leave(p *Connection) {
	delete(s.players, p)
}

func (s *System) EachConn(fn func(*Connection)) {
	if s.players == nil {
		return
	}
	for conn, _ := range s.players {
		fn(conn)
	}
}

func (s *System) NumInhabitants() int {
	if s.players == nil {
		return 0
	}
	return len(s.players)
}

func (e System) Store(db *sql.DB) {
	_, err := db.Exec(`
    insert into planets
    (name, x, y, z, planets)
    values
    (?, ?, ?, ?, ?)
    ;`, e.name, e.x, e.y, e.z, e.planets)
	if err != nil {
		log_error("%v", err)
	}
}

func (s *System) DistanceTo(other *System) float64 {
	return dist3d(s.x, s.y, s.z, other.x, other.y, other.z)
}

func (e System) String() string {
	return fmt.Sprintf("<name: %s x: %v y: %v z: %v planets: %v>", e.name, e.x, e.y, e.z, e.planets)
}

type Neighbor struct {
	id       int
	distance float64
}

func (e *System) Nearby(n int) ([]Neighbor, error) {
	rows, err := db.Query(`
        select planets.id, edges.distance
        from edges
        join planets on edges.id_2 = planets.id
        where edges.id_1 = ?
        order by distance
        limit ?
    ;`, e.id, n)
	if err != nil {
		log_error("unable to get nearby systems for %s: %v", e.name, err)
		return nil, err
	}
	neighbors := make([]Neighbor, 0, n)
	for rows.Next() {
		var neighbor Neighbor
		if err := rows.Scan(&neighbor.id, &neighbor.distance); err != nil {
			log_error("error unpacking row from nearby neighbors query: %v", err)
			continue
		}
		neighbors = append(neighbors, neighbor)
	}
	return neighbors, nil
}

func countPlanets() (int, error) {
	row := db.QueryRow(`select count(*) from planets`)

	var n int
	err := row.Scan(&n)
	return n, err
}

func sq(x float64) float64 {
	return x * x
}

func dist3d(x1, y1, z1, x2, y2, z2 float64) float64 {
	return math.Sqrt(sq(x1-x2) + sq(y1-y2) + sq(z1-z2))
}

func planetDistance(p1, p2 System) float64 {
	return dist3d(p1.x, p1.y, p1.z, p2.x, p2.y, p2.z)
}

func indexPlanets(db *sql.DB) map[int]*System {
	rows, err := db.Query(`select * from planets`)
	if err != nil {
		log_error("unable to select all planets: %v", err)
		return nil
	}
	defer rows.Close()
	index = make(map[int]*System, 551)
	for rows.Next() {
		p := System{}
		if err := rows.Scan(&p.id, &p.name, &p.x, &p.y, &p.z, &p.planets); err != nil {
			log_info("unable to scan planet row: %v", err)
			continue
		}
		index[p.id] = &p
	}
	return index
}

func randomPlanet() (*System, error) {
	n := len(index)
	if n == 0 {
		return nil, fmt.Errorf("no planets are known to exist")
	}

	pick := rand.Intn(n)
	planet := index[pick]
	return planet, nil
}
