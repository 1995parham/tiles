package server

import "time"

// Client is a remote connection into Tiles
type Client struct {
	id int

	remoteAddr string // original remote address

	opened time.Time // when the client was created/opened, unix nano
	last   time.Time // last client request/response, unix nano
}
