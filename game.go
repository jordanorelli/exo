package main

import (
	"time"
)

type Game struct {
	id        Id
	start     time.Time
	end       time.Time
	winner    string
	winMethod string
}

func gamesTable() {
	stmnt := `create table if not exists games (
        id text not null,
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
		id:    NewId(),
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
        (id, start)
        values
        (?, ?)
    ;`, g.id.String(), g.start)
	return err
}
