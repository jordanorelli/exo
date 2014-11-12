package main

import (
	"fmt"
	"time"
)

type Game struct {
	id          Id
	start       time.Time
	end         time.Time
	winner      string
	winMethod   string
	connections map[*Connection]bool
	frame       int64
	elems       map[GameElement]bool
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
		id:          NewId(),
		start:       time.Now(),
		connections: make(map[*Connection]bool, 32),
		elems:       make(map[GameElement]bool, 32),
	}
	if err := game.Create(); err != nil {
		log_error("unable to create game: %v", err)
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
	if err != nil {
		return fmt.Errorf("error writing sqlite insert statement to create game: %v")
	}
	return nil
}

func (g *Game) Store() error {
	_, err := db.Exec(`
        update games
        set end = ?, winner = ?, win_method = ?
        where id = ?
    ;`, g.end, g.winner, g.winMethod, g.id)
	return err
}

func (g *Game) Join(conn *Connection) {
	g.connections[conn] = true
}

func (g *Game) Quit(conn *Connection) {
	delete(g.connections, conn)
}

func (g *Game) Win(winner *Connection, method string) {
	g.end = time.Now()
	g.winner = winner.PlayerName()
	g.winMethod = method
	g.Store()

	for conn, _ := range g.connections {
		fmt.Fprintf(conn, "player %s has won by %s victory.\n", winner.PlayerName(), method)
		fmt.Fprintf(conn, "starting new game ...\n")
		conn.Reset()
	}

	ResetQueue()

	g.Reset()

	for conn, _ := range g.connections {
		conn.Respawn()
	}
}

func (g *Game) Reset() {
	connections := g.connections
	fresh := NewGame()
	*g = *fresh
	g.connections = connections
}

func (g *Game) Run() {
	ticker := time.Tick(time.Second / time.Duration(frameRate))
	for {
		select {
		case <-ticker:
			g.tick()
		}
	}
}

func (g *Game) Register(elem GameElement) {
	g.elems[elem] = true
}

func (g *Game) tick() {
	g.frame += 1
	for elem := range g.elems {
		if elem.Dead() {
			delete(g.elems, elem)
		}
	}
	for elem := range g.elems {
		elem.Tick(g.frame)
	}
}

type GameElement interface {
	Tick(frame int64)
	Dead() Mortality
}

type Mortality bool

func (m Mortality) Dead() Mortality {
	return m
}
