package main

import (
	"fmt"
	"strings"
)

const (
	E_Ok int = iota
	E_No_Data
	E_No_DB
	E_No_Port
	E_Bad_Duration
	E_Missing_Slack_OAuth_Token
)

type errorGroup []error

func (e errorGroup) Error() string {
	messages := make([]string, 0, len(e))
	for i, _ := range e {
		messages[i] = e[i].Error()
	}
	return strings.Join(messages, " && ")
}

func (g *errorGroup) AddError(err error) {
	if err == nil {
		return
	}
	if g == nil {
		panic("fart")
		*g = make([]error, 0, 4)
	}
	*g = append(*g, err)
}

// ErrorState represents a valid client state indicating that the client has
// hit an error.  On tick, the client will be disconnected.  ErrorState is both
// a valid ConnectionState and a valid error value.
type ErrorState struct {
	CommandSuite
	error
	NopEnter
	NopExit
}

func NewErrorState(e error) *ErrorState {
	return &ErrorState{error: e}
}

func (e *ErrorState) Tick(c *Connection, frame int64) ConnectionState {
	c.Printf("something went wrong: %v", e.error)
	log_error("player hit error: %v", e.error)
	c.Close()
	return nil
}

func (e *ErrorState) String() string {
	return fmt.Sprintf("error state: %v", e.error)
}

func (e *ErrorState) RunCommand(c *Connection, name string, args ...string) ConnectionState {
	return e
}
