package main

import (
	"regexp"
)

var namePattern = regexp.MustCompile(`^[[:alpha:]][[:alnum:]-_]{0,19}$`)

func ValidName(name string) bool {
	return namePattern.MatchString(name)
}

type Player struct {
	name string
}

func (p *Player) Load() {

}
