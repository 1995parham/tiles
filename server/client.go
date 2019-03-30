package server

import (
	"sync"
	"time"

	"github.com/tidwall/evio"
)

// Client is a remote connection into Tiles
type Client struct {
	id         int64            // unique id
	remoteAddr string           // original remote address
	in         evio.InputStream // input stream
	pr         PipelineReader   // command reader
	out        []byte           // output write buffer

	mu     sync.Mutex // guard
	opened time.Time  // when the client was created/opened, unix nano
	last   time.Time  // last client request/response, unix nano
}

// Write writes on client output buffer
func (client *Client) Write(b []byte) (n int, err error) {
	client.out = append(client.out, b...)
	return len(b), nil
}
