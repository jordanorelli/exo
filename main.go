package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

var options struct {
	lightSpeed  float64
	frameRate   int
	miningRate  int
	playerSpeed float64
	bombSpeed   float64
	economic    int
}

var (
	dataPath    = "/projects/exo/expl.speck"
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
	for {
		line, err := conn.ReadString('\n')
		switch err {
		case io.EOF:
			return
		case nil:
			break
		default:
			log_error("failed to read line from player %s: %v", conn.PlayerName(), err)
			return
		}
		line = strings.TrimSpace(line)

		if conn.IsMining() {
			conn.StopMining()
		}

		if line == "" {
			continue
		}
		parts := strings.Split(line, " ")

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
	listener, err := net.Listen("tcp", ":9220")
	if err != nil {
		bail(E_No_Port, "unable to start server: %v", err)
	}

	currentGame = NewGame()
	go currentGame.Run()

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
	flag.IntVar(&options.miningRate, "mining-rate", 1, "mining rate, in duckets per frame")
	flag.Float64Var(&options.playerSpeed, "player-speed", 0.8, "player travel speed, relative to C, the speed of light")
	flag.Float64Var(&options.bombSpeed, "bomb-speed", 0.9, "bomb travel speed, relattive to C, the speed of light")
	flag.IntVar(&options.economic, "economic", 25000, "amount of money needed to win economic victory")
}
