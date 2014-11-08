package main

import (
	"fmt"
	"os"
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

func main() {
	setupDb()
}
