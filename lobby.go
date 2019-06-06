package main

import (
	"strings"
	"time"
)

var banner = `
##############################################################################################


 /$$$$$$$$                                         /$$
| $$_____/                                        | $$
| $$       /$$   /$$  /$$$$$$   /$$$$$$$  /$$$$$$ | $$  /$$$$$$  /$$$$$$$  /$$   /$$  /$$$$$$$
| $$$$$   |  $$ /$$/ /$$__  $$ /$$_____/ /$$__  $$| $$ /$$__  $$| $$__  $$| $$  | $$ /$$_____/
| $$__/    \  $$$$/ | $$  \ $$| $$      | $$  \ $$| $$| $$  \ $$| $$  \ $$| $$  | $$|  $$$$$$
| $$        >$$  $$ | $$  | $$| $$      | $$  | $$| $$| $$  | $$| $$  | $$| $$  | $$ \____  $$
| $$$$$$$$ /$$/\  $$|  $$$$$$/|  $$$$$$$|  $$$$$$/| $$|  $$$$$$/| $$  | $$|  $$$$$$/ /$$$$$$$/
|________/|__/  \__/ \______/  \_______/ \______/ |__/ \______/ |__/  |__/ \______/ |_______/
                      

                                      ~+
                              
                                               *       +
                                         '                  |
                                     ()    .-.,="''"=.    - o -
                                           '=/_       \     |
                                        *   |  '=._    |
                                             \     '=./',        '
                                          .   '=.__.=' '='      *
                                 +                         +
                                      O      *        '       .
        

A game of dark cunning in the vast unknown of space by Jordan Orelli.

##############################################################################################
`

type LobbyState struct {
	CommandSuite
	NopExit
}

func EnterLobby() ConnectionState {
	return &LobbyState{
		CommandSuite: CommandSet{
			newGameCommand,
			joinGameCommand,
		},
	}
}

func (st *LobbyState) String() string { return "Lobby" }

func (st *LobbyState) Enter(c *Connection) {
	c.Printf(strings.TrimSpace(banner))
	time.Sleep(1 * time.Second)

	for {
		c.Printf("\n\nWhat is your name, adventurer?\n")
		name, err := c.ReadString('\n')
		if err == nil {
			name = strings.TrimSpace(name)
		} else {
			log_error("player failed to connect: %v", err)
			return
		}

		if !ValidName(name) {
			c.Printf("that name is illegal.\n")
			continue
		}
		log_info("player connected: %v", name)
		profile, err := loadProfile(name)
		if err != nil {
			log_error("could not read profile: %v", err)
			profile = &Profile{name: name}
			if err := profile.Create(); err != nil {
				log_error("unable to create profile record: %v", err)
			}
			c.Printf("you look new around these parts, %s.\n", profile.name)
			c.Printf(`if you'd like a description of how to play, type the "help" command\n`)
			c.profile = profile
		} else {
			c.profile = profile
			c.Printf("Welcome back, %s.\n", profile.name)
		}
		break
	}
	c.ListCommands()
}

func (st *LobbyState) Tick(c *Connection, frame int64) ConnectionState { return st }

func (st *LobbyState) FillStatus(c *Connection, s *status) {
	s.Description = strings.TrimSpace(`
Currently in the Lobby, waiting for you to issue a "new" command to start a new
game, or a "join" command to join an existing game.
`)
}

var newGameCommand = Command{
	name:     "new",
	summary:  "starts a new game",
	arity:    0,
	variadic: false,
	handler: func(c *Connection, args ...string) {
		c.Printf("Starting a new game...\n")
		game := gm.NewGame()
		log_info("%s Created game: %s", c.profile.name, game.id)
		go game.Run()
		c.game = game
		c.Printf("Now playing in game: %s\n\n", game.id)
		c.Line()
		c.game.Join(c)
		c.SetState(SpawnRandomly())
	},
	debug: false,
}

var joinGameCommand = Command{
	name:     "join",
	summary:  "joins an existing game",
	usage:    "join [game-code]",
	arity:    1,
	variadic: false,
	handler: func(c *Connection, args ...string) {
		if len(args) == 0 {
			c.Printf(strings.TrimLeft(`
Missing game code! When a player starts a game, they will be given a code to
identify their game. Use this game to join the other player's game.

Usage: join [game-code]`, " \n\t"))
			return
		}
		id := args[0]
		c.game = gm.Get(id)
		log_info("%s Joining game: %s", c.profile.name, c.game.id)
		c.SetState(SpawnRandomly())
		c.game.Join(c)
	},
	debug: false,
}
