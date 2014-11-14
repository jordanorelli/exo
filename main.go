package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"
)

var options struct {
	lightSpeed  float64
	frameRate   int
	moneySigma  float64
	moneyMean   float64
	playerSpeed float64
	bombSpeed   float64
	economic    int
	debug       bool
	speckPath   string
}

var (
	info_log    *log.Logger
	error_log   *log.Logger
	currentGame *Game
)

func log_error(template string, args ...interface{}) {
	error_log.Printf(template, args...)
}

func log_info(template string, args ...interface{}) {
	info_log.Printf(template, args...)
}

func bail(status int, template string, args ...interface{}) {
	if status == 0 {
		fmt.Fprintf(os.Stdout, template, args...)
	} else {
		fmt.Fprintf(os.Stderr, template, args...)
	}
	os.Exit(status)
}

func handleConnection(conn *Connection) {
	defer conn.Close()
	conn.Login()

	conn.Respawn()

	c := make(chan []string)
	go conn.ReadLines(c)

	for parts := range c {
		if isCommand(parts[0]) {
			runCommand(conn, parts[0], parts[1:]...)
			continue
		}

		switch parts[0] {
		case "quit":
			return
		default:
			fmt.Fprintf(conn, "hmm I'm not sure I know that one.\n")
		}

	}
}

func main() {
	flag.Parse()
	dbconnect()

	rand.Seed(time.Now().UnixNano())
	info_log = log.New(os.Stdout, "[INFO] ", 0)
	error_log = log.New(os.Stderr, "[ERROR] ", 0)

	setupDb()
	setupCommands()
	listener, err := net.Listen("tcp", ":9220")
	if err != nil {
		bail(E_No_Port, "unable to start server: %v", err)
	}

	go func() {
		for {
			log_info("starting new game")
			currentGame = NewGame()
			currentGame.Run()
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log_error("error accepting connection: %v", err)
			continue
		}
		go handleConnection(NewConnection(conn))
	}
}

func init() {
	flag.Float64Var(&options.lightSpeed, "light-speed", 0.01, "speed of light in parsecs per frame")
	flag.IntVar(&options.frameRate, "frame-rate", 100, "frame rate, in frames per second")
	flag.Float64Var(&options.playerSpeed, "player-speed", 0.8, "player travel speed, relative to C, the speed of light")
	flag.Float64Var(&options.bombSpeed, "bomb-speed", 0.9, "bomb travel speed, relattive to C, the speed of light")
	flag.IntVar(&options.economic, "economic", 25000, "amount of money needed to win economic victory")
	flag.Float64Var(&options.moneyMean, "money-mean", 10000, "mean amount of money on a system")
	flag.Float64Var(&options.moneySigma, "money-sigma", 1500, "standard deviation in money per system")
	flag.BoolVar(&options.debug, "debug", false, "puts the game in debug mode")
	flag.StringVar(&options.speckPath, "speck-path", "/projects/exo/expl.speck", "path to exoplanet speck file")
}
