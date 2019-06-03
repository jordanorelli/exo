package main

import (
	"sync"
)

var gm *GameManager

func init() {
	gm = &GameManager{
		games: make(map[string]*Game, 32),
	}
}

type GameManager struct {
	games map[string]*Game
	sync.Mutex
}

func (g *GameManager) NewGame() *Game {
	g.Lock()
	defer g.Unlock()

	game := NewGame()
	g.games[game.id] = game
	return game
}

func (g *GameManager) Get(id string) *Game {
	g.Lock()
	defer g.Unlock()

	return g.games[id]
}
