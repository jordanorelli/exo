package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Game struct {
	id          string
	start       time.Time
	end         time.Time
	done        chan interface{}
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

func init() { rand.Seed(time.Now().UnixNano()) }

func newID() string {
	chars := []rune("ABCDEEEEEEEEFGHJJJJJJJKMNPQQQQQQQRTUVWXXXXXYZZZZZ234677777789")
	id := make([]rune, 0, 4)
	for i := 0; i < cap(id); i++ {
		id = append(id, chars[rand.Intn(len(chars))])
	}
	return string(id)
}

func NewGame() *Game {
	game := &Game{
		id:          newID(),
		start:       time.Now(),
		done:        make(chan interface{}),
		connections: make(map[*Connection]bool, 32),
		elems:       make(map[GameElement]bool, 32),
	}
	if err := game.Create(); err != nil {
		log_error("unable to create game: %v", err)
	}
	for _, system := range index {
		game.Register(system)
	}
	if currentGame != nil {
		log_info("passing %d connections...", len(currentGame.connections))
		for conn, _ := range currentGame.connections {
			log_info("moving player %s to new game", conn.Name())
			currentGame.Quit(conn)
			game.Join(conn)
		}
	}
	return game
}

func (g *Game) Create() error {
	_, err := db.Exec(`
        insert into games
        (id, start)
        values
        (?, ?)
    ;`, g.id, g.start)
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
	defer close(g.done)
	g.end = time.Now()
	g.winner = winner.Name()
	g.winMethod = method
	g.Store()

	log_info("player %s has won by %s victory", winner.Name(), method)

	for conn, _ := range g.connections {
		conn.Printf("player %s has won by %s victory.\n", winner.Name(), method)
	}
}

func (g *Game) Reset() {
	connections := g.connections
	fresh := NewGame()
	*g = *fresh
	g.connections = connections
}

func (g *Game) Run() {
	ticker := time.Tick(time.Second / time.Duration(options.frameRate))
	for {
		select {
		case <-ticker:
			g.tick()
		case <-g.done:
			for conn, _ := range g.connections {
				conn.Close()
			}
			return
		}
	}
}

func (g *Game) Register(elem GameElement) {
	g.elems[elem] = true
}

func (g *Game) tick() {
	g.frame += 1
	for elem := range g.elems {
		elem.Tick(g.frame)
	}
	for elem := range g.elems {
		if elem.Dead() {
			log_info("delete game object: %v", elem)
			delete(g.elems, elem)
		}
	}
}

type GameElement interface {
	Tick(frame int64)
	Dead() bool
}
