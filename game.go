package main

import (
	"time"
)

type Game struct {
	start     time.Time
	end       time.Time
	winner    string
	winMethod string
}

func gamesTable() {
	stmnt := `create table if not exists games (
        id integer not null primary key autoincrement,
        start text not null,
        end text,
        winner text,
        win_method text
    );`
	if _, err := db.Exec(stmnt); err != nil {
		log_error("couldn't create games table: %v", err)
	}
}

func NewGame() *Game {
	game := &Game{
		start: time.Now(),
	}
	if err := game.Create(); err != nil {
		log_error("%v", err)
	}
	return game
}

func (g *Game) Create() error {
	_, err := db.Exec(`
        insert into games
        (start)
        values
        (?)
    ;`, g.start)
	return err
}
