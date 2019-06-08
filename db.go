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
		fi, err := os.Open(options.speckPath)
		if err != nil {
			bail(E_No_Data, "unable to open data path: %v", err)
		}
		c := make(chan System)
		go speckStream(fi, c)
		for planet := range c {
			planet.Store(db)
		}
	}
	// indexSystems()
}

func setupDb() {
	planetsTable()
	planetsData()
	profilesTable()
	gamesTable()
}
