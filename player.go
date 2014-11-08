package main

import (
	"fmt"
	"regexp"
)

var namePattern = regexp.MustCompile(`^[[:alpha:]][[:alnum:]-_]{0,19}$`)

func ValidName(name string) bool {
	return namePattern.MatchString(name)
}

type Player struct {
	id   int
	name string
}

func (p *Player) Create() error {
	_, err := db.Exec(`
        insert into players
        (name)
        values
        (?)
    ;`, p.name)
	if err != nil {
		return fmt.Errorf("unable to create player: %v", err)
	}
	return nil
}

func playersTable() {
	stmnt := `create table if not exists players (
        id integer not null primary key autoincrement,
        name text unique
    );`
	if _, err := db.Exec(stmnt); err != nil {
		log_error("couldn't create player table: %v", err)
	}
}

func loadPlayer(name string) (*Player, error) {
	row := db.QueryRow(`select * from players where name = ?`, name)
	var p Player
	if err := row.Scan(&p.id, &p.name); err != nil {
		return nil, fmt.Errorf("unable to fetch player from database: %v", err)
	}
	return &p, nil
}
