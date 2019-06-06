package main

type ConnectionState interface {
	// commands available while in this state
	CommandSuite

	// human-readable description of the state
	String() string

	// fills a status struct to be printed by the status command. The
	// ConnectionState only needs to fill in things that are unique to the
	// state itself, the common things on the connection are filled in
	// automatically
	FillStatus(*Connection, *status)

	// Triggered once when the state is entered
	Enter(c *Connection)

	// Triggered every frame in which the state is the connection's current
	// state. Returning a different ConnectionState transitions between states.
	Tick(c *Connection, frame int64) ConnectionState

	// Triggered once when this state has finished for that connection
	Exit(c *Connection)
}

// No-op enter struct, for composing connection states that have no interesitng
// Enter mechanic.
type NopEnter struct{}

func (n NopEnter) Enter(c *Connection) {}

// No-op exit struct, for composing connection states that have no interesting
// Exit mechanic.
type NopExit struct{}

func (n NopExit) Exit(c *Connection) {}


