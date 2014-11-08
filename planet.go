package main

import (
	"database/sql"
	"fmt"
	"math"
	"math/rand"
)

type Planet struct {
	x, y, z float64
	planets int
	name    string
}

func (e Planet) Store(db *sql.DB) {
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

func (e Planet) String() string {
	return fmt.Sprintf("<name: %s x: %v y: %v z: %v planets: %v>", e.name, e.x, e.y, e.z, e.planets)
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

func planetDistance(p1, p2 Planet) float64 {
	return dist3d(p1.x, p1.y, p1.z, p2.x, p2.y, p2.z)
}

func indexPlanets(db *sql.DB) map[int]Planet {
	rows, err := db.Query(`select * from planets`)
	if err != nil {
		log_error("unable to select all planets: %v", err)
		return nil
	}
	defer rows.Close()
	planetIndex = make(map[int]Planet, 551)
	for rows.Next() {
		var id int
		p := Planet{}
		if err := rows.Scan(&id, &p.name, &p.x, &p.y, &p.z, &p.planets); err != nil {
			log_info("unable to scan planet row: %v", err)
			continue
		}
		planetIndex[id] = p
	}
	return planetIndex
}

func randomPlanet() (*Planet, error) {
	n := len(planetIndex)
	if n == 0 {
		return nil, fmt.Errorf("no planets are known to exist")
	}

	pick := rand.Intn(n)
	planet := planetIndex[pick]
	return &planet, nil
}
