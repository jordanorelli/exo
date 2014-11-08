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
	n, err := countSystems()
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
	indexSystems()
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
	fillEdges()
}

func fillEdges() {
	row := db.QueryRow(`select count(*) from edges;`)
	var n int
	if err := row.Scan(&n); err != nil {
		log_error("couldn't get number of edges: %v", err)
		return
	}
	if n > 0 {
		return
	}
	for i := 0; i < len(index); i++ {
		for j := 0; j < len(index); j++ {
			if i == j {
				continue
			}
			if index[i] == nil {
				log_error("wtf there's nil shit in here for id %d", i)
				continue
			}
			if index[j] == nil {
				log_error("wtf there's nil shit in here 2 for id %d", j)
				continue
			}
			dist := index[i].DistanceTo(index[j])
			log_info("distance from %s to %s: %v", index[i].name, index[j].name, dist)
			_, err := db.Exec(`
                insert into edges
                (id_1, id_2, distance)
                values
                (?, ?, ?)
            ;`, i, j, dist)
			if err != nil {
				log_error("unable to write edge to db: %v", err)
			}
		}
	}
}
