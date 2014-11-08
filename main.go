package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
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

func handleConnection(conn net.Conn) {
    namePattern := regexp.MustCompile(`^[[:alpha:]][[:alnum:]-_]{0,19}$`)
	r := bufio.NewReader(conn)
	fmt.Fprintf(conn, "what is your name, adventurer?\n")
	name, err := r.ReadString('\n')
	if err == nil {
		name = strings.TrimSpace(name)
		log_info("player connected: %v", name)
	} else {
		log_error("player failed to connect: %v", err)
	}
    if !namePattern.MatchString(name) {
        fmt.Fprintf(conn, "that name is illegal.\n")
    }

    planet, err := randomPlanet()
    if err != nil {
        log_error("player %s failed to get random planet: %v", name, err)
        return
    }
    fmt.Fprintf(conn, "you are on the planet %s\n", planet.name)
}

func main() {
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
		go handleConnection(conn)
	}
}
