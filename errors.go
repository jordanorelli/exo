package main

import (
    "strings"
)

const (
	E_Ok int = iota
	E_No_Data
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
