package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

var (
	db *sql.DB
)

func dbconnect() {
	var err error
	db, err = sql.Open("sqlite3", "./exo.db")
	if err != nil {
		bail(E_No_DB, "couldn't connect to db: %v", err)
	}
}

func planetsTable() {
	stmnt := `create table if not exists planets (
        id integer not null primary key autoincrement,
        name text,
        x integer,
        y integer,
        z integer,
        planets integer
    );`
	if _, err := db.Exec(stmnt); err != nil {
		log_error("couldn't create planets table: %v", err)
	}
}

func planetsData() {
	n, err := countPlanets()
	if err != nil {
		log_error("couldn't count planets: %v", err)
		return
	}
	if n == 0 {
		fi, err := os.Open(dataPath)
		if err != nil {
			bail(E_No_Data, "unable to open data path: %v", err)
		}
		c := make(chan System)
		go speckStream(fi, c)
		for planet := range c {
			planet.Store(db)
		}
	}
	indexPlanets(db)
}

func edgesTable() {
	stmnt := `create table if not exists edges (
        id_1 integer,
        id_2 integer,
        distance real
    );`
	if _, err := db.Exec(stmnt); err != nil {
		log_error("couldn't create distance table: %v", err)
	}
}

func setupDb() {
	planetsTable()
	planetsData()
	edgesTable()
	playersTable()
	// fillEdges(db, idx)
}

func fillEdges(db *sql.DB, planets map[int]System) {
	for i := 0; i < len(planets); i++ {
		for j := i + 1; j < len(planets); j++ {
			log_info("distance from %s to %s: %v", planets[i].name, planets[j].name, planetDistance(planets[i], planets[j]))
		}
	}
}
