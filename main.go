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

var (
	dataPath = "/projects/exo/expl.speck"
)

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
	planet.Arrive(conn)
	if planet.planets == 1 {
		fmt.Fprintf(conn, "you are in the system %s. There is %d planet here.\n", planet.name, planet.planets)
	} else {
		fmt.Fprintf(conn, "you are in the system %s. There are %d planets here.\n", planet.name, planet.planets)
	}
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
		if isCommand(parts[0]) {
			runCommand(conn, parts[0], parts[1:]...)
			continue
		}

		switch parts[0] {
		case "scan":
			for _, otherPlanet := range index {
				if otherPlanet.name == planet.name {
					continue
				}
				go func(p *System) {
					dist := planetDistance(*planet, *p)
					delay := time.Duration(int64(dist * 100000000))
					time.Sleep(delay)
					mu.Lock()
					fmt.Fprintf(conn, "PONG from planet %s (%v)\n", p.name, delay)
					mu.Unlock()
				}(otherPlanet)
			}
		case "broadcast":
			msg := strings.Join(parts[1:], " ")
			log_info("player %s is broadcasting message %s", conn.PlayerName(), msg)
			for _, otherSystem := range index {
				if otherSystem.name == planet.name {
					log_info("skpping duplicate system %s", planet.name)
					continue
				}
				go func(s *System) {
					log_info("message reached planet %s with %d inhabitants", s.name, s.NumInhabitants())
					dist := planetDistance(*planet, *s) * 0.5
					delay := time.Duration(int64(dist * 100000000))
					time.Sleep(delay)
					s.EachConn(func(conn *Connection) {
						fmt.Fprintln(conn, msg)
					})
				}(otherSystem)
			}
		case "nearby":
			neighbors, err := planet.Nearby(25)
			fmt.Fprintf(conn, "fetching nearby star systems\n")
			if err != nil {
				log_error("%v", err)
				break
			}
			fmt.Fprintf(conn, "found %d nearby systems\n", len(neighbors))
			for _, neighbor := range neighbors {
				other := index[neighbor.id]
				fmt.Fprintf(conn, "%s: %v\n", other.name, neighbor.distance)
			}
		case "quit":
			return
		default:
			fmt.Fprintf(conn, "hmm I'm not sure I know that one.\n")
		}
	}
}

func main() {
	dbconnect()
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
