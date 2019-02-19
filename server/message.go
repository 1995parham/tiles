package server

import "strings"

// Type is resp type
type Type byte

// Protocol Types
const (
	Null Type = iota
	RESP
	Telnet
	Native
	HTTP
	WebSocket
	JSON
)

// Message is a resp message
type Message struct {
	_command   string
	Args       []string
	ConnType   Type
	OutputType Type
	Auth       string
}

// Command returns the first argument as a lowercase string
func (msg *Message) Command() string {
	if msg._command == "" {
		msg._command = strings.ToLower(msg.Args[0])
	}
	return msg._command
}
