package server

import (
	"io"

	"github.com/tidwall/redcon"
)

// PipelineReader contains a source to parse the packets.
type PipelineReader struct {
	rd     io.Reader
	packet [0xFFFF]byte
	buf    []byte
}

// ReadMessages read from the reader and returns messages.
func (rd *PipelineReader) ReadMessages() ([]*Message, error) {
	var msgs []*Message
moreData:
	n, err := rd.rd.Read(rd.packet[:])
	if err != nil {
		return nil, err
	}
	if n == 0 {
		// need more data
		goto moreData
	}
	data := rd.packet[:n]
	if len(rd.buf) > 0 {
		data = append(rd.buf, data...)
	}
	for len(data) > 0 {
		msg := &Message{}
		complete, args, kind, leftover, err := redcon.ReadNextCommand(data, nil)
		if err != nil {
			break
		}
		if !complete {
			break
		}
		if len(args) > 0 {
			for i := 0; i < len(args); i++ {
				msg.Args = append(msg.Args, string(args[i]))
			}
			switch kind {
			case redcon.Redis:
				msg.ConnType = RESP
				msg.OutputType = RESP
			case redcon.Tile38:
				msg.ConnType = Native
				msg.OutputType = JSON
			case redcon.Telnet:
				msg.ConnType = RESP
				msg.OutputType = RESP
			}
			msgs = append(msgs, msg)
		}
		data = leftover
	}
	if len(data) > 0 {
		rd.buf = append(rd.buf[:0], data...)
	} else if len(rd.buf) > 0 {
		rd.buf = rd.buf[:0]
	}
	if err != nil && len(msgs) == 0 {
		return nil, err
	}
	return msgs, nil
}
