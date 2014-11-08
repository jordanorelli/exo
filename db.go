package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"math"
	"math/rand"
	"os"
)

var planetIndex map[int]exoSystem

func dbconnect() (*sql.DB, error) {
	return sql.Open("sqlite3", "./exo.db")
}

func setupDb() {
	db, err := dbconnect()
	if err != nil {
		bail(E_No_DB, "unable to open database: %v", err)
	}
	defer db.Close()

	stmnt := `create table if not exists planets (
        id integer not null primary key autoincrement,
        name text,
        x integer,
        y integer,
        z integer,
        planets integer
    );`
	_, err = db.Exec(stmnt)
	if err != nil {
		log_error("couldn't create table: %v", err)
		return
	}

	n, err := countPlanets(db)
	if n == 0 {
		fi, err := os.Open(dataPath)
		if err != nil {
			bail(E_No_Data, "unable to open data path: %v", err)
		}
		c := make(chan exoSystem)
		go speckStream(fi, c)
		for planet := range c {
			planet.Store(db)
		}
	}

	indexPlanets(db)
	// log_info("%v", idx)
	// fillEdges(db, idx)

	stmnt = `create table if not exists edges (
        id_1 integer,
        id_2 integer,
        distance real
    );`
	_, err = db.Exec(stmnt)
	if err != nil {
		log_error("couldn't create distance table: %v", err)
		return
	}

	stmnt = `create table if not exists players (
        id integer not null primary key autoincrement,
        name text
    );`
	_, err = db.Exec(stmnt)
	if err != nil {
		log_error("couldn't create player table: %v", err)
		return
	}

}

func sq(x float64) float64 {
	return x * x
}

func dist3d(x1, y1, z1, x2, y2, z2 float64) float64 {
	return math.Sqrt(sq(x1-x2) + sq(y1-y2) + sq(z1-z2))
}

func planetDistance(p1, p2 exoSystem) float64 {
	return dist3d(p1.x, p1.y, p1.z, p2.x, p2.y, p2.z)
}

func indexPlanets(db *sql.DB) map[int]exoSystem {
	rows, err := db.Query(`select * from planets`)
	if err != nil {
		log_error("unable to select all planets: %v", err)
		return nil
	}
	defer rows.Close()
	planetIndex = make(map[int]exoSystem, 551)
	for rows.Next() {
		var id int
		p := exoSystem{}
		if err := rows.Scan(&id, &p.name, &p.x, &p.y, &p.z, &p.planets); err != nil {
			log_info("unable to scan planet row: %v", err)
			continue
		}
		planetIndex[id] = p
	}
	return planetIndex
}

func fillEdges(db *sql.DB, planets map[int]exoSystem) {
	for i := 0; i < len(planets); i++ {
		for j := i + 1; j < len(planets); j++ {
			log_info("distance from %s to %s: %v", planets[i].name, planets[j].name, planetDistance(planets[i], planets[j]))
		}
	}
}

func randomPlanet() (*exoSystem, error) {
	n := len(planetIndex)
	if n == 0 {
		return nil, fmt.Errorf("no planets are known to exist")
	}

	pick := rand.Intn(n)
    planet := planetIndex[pick]
    return &planet, nil
}
