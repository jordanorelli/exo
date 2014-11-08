package main

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var dataPath = "/projects/exo/expl.speck"

func log_error(template string, args ...interface{}) {
	fmt.Fprint(os.Stderr, "ERROR ")
	fmt.Fprintf(os.Stderr, template+"\n", args...)
}

func log_info(template string, args ...interface{}) {
	fmt.Fprint(os.Stdout, "INFO ")
	fmt.Fprintf(os.Stdout, template+"\n", args...)
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
	var mu sync.Mutex

	defer conn.Close()
	conn.Login()

	planet, err := randomPlanet()
	if err != nil {
		log_error("player %s failed to get random planet: %v", conn.PlayerName(), err)
		return
	}
	fmt.Fprintf(conn, "you are on the planet %s\n", planet.name)
	for {
		line, err := conn.ReadString('\n')
		switch err {
		case io.EOF:
			return
		case nil:
			break
		default:
			log_error("failed to read line from player %s: %v", conn.PlayerName(), err)
		}
		line = strings.TrimSpace(line)
		parts := strings.Split(line, " ")
		switch parts[0] {
		case "scan":
			for _, otherPlanet := range planetIndex {
				if otherPlanet.name == planet.name {
					continue
				}
				go func(p exoSystem) {
					dist := planetDistance(*planet, p)
					delay := time.Duration(int64(dist * 100000000))
					time.Sleep(delay)
					mu.Lock()
					fmt.Fprintf(conn, "PONG from planet %s (%v)\n", p.name, delay)
					mu.Unlock()
				}(otherPlanet)
			}
		case "broadcast":

		case "quit":
			return
		default:
			fmt.Fprintf(conn, "hmm I'm not sure I know that one.\n")
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	setupDb()
	listener, err := net.Listen("tcp", ":9220")
	if err != nil {
		bail(E_No_Port, "unable to start server: %v", err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log_error("error accepting connection: %v", err)
			continue
		}
		go handleConnection(NewConnection(conn))
	}
}
