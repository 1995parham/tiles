package server

import (
	"errors"
	"fmt"
	"io"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/redcon"
	"github.com/tidwall/resp"
)

func isRespValueEmptyString(val resp.Value) bool {
	return !val.IsNull() && (val.Type() == resp.SimpleString || val.Type() == resp.BulkString) && len(val.Bytes()) == 0
}

func (s *Server) handleInputCommand(client io.Writer, msg *Message) error {
	start := time.Now()

	serializeOutput := func(res resp.Value) (string, error) {
		var resStr string
		var err error
		switch msg.OutputType {
		case JSON:
			resStr = res.String()
		case RESP:
			var resBytes []byte
			resBytes, err = res.MarshalRESP()
			resStr = string(resBytes)
		}
		return resStr, err
	}

	writeOutput := func(res string) error {
		switch msg.ConnType {
		default:
			err := fmt.Errorf("unsupported conn type: %v", msg.ConnType)
			log.Error(err)
			return err
		case RESP:
			var err error
			if msg.OutputType == JSON {
				_, err = fmt.Fprintf(client, "$%d\r\n%s\r\n", len(res), res)
			} else {
				_, err = io.WriteString(client, res)
			}
			return err
		case Native:
			_, err := fmt.Fprintf(client, "$%d %s\r\n", len(res), res)
			return err
		}
	}

	// Ping. Just send back the response. No need to put through the pipeline.
	if msg.Command() == "ping" || msg.Command() == "echo" {
		switch msg.OutputType {
		case JSON:
			if len(msg.Args) > 1 {
				return writeOutput(`{"ok":true,"` + msg.Command() + `":` + jsonString(msg.Args[1]) + `,"elapsed":"` + time.Now().Sub(start).String() + `"}`)
			}
			return writeOutput(`{"ok":true,"` + msg.Command() + `":"pong","elapsed":"` + time.Since(start).String() + `"}`)
		case RESP:
			if len(msg.Args) > 1 {
				data := redcon.AppendBulkString(nil, msg.Args[1])
				return writeOutput(string(data))
			}
			return writeOutput("+PONG\r\n")
		}
		return nil
	}

	// Command. Just send back the ok response to have a simple redis cli.
	if msg.Command() == "command" {
		if msg.OutputType == RESP {
			return writeOutput("+OK\r\n")
		}
		return nil
	}

	writeErr := func(errMsg string) error {
		switch msg.OutputType {
		case JSON:
			return writeOutput(`{"ok":false,"err":` + jsonString(errMsg) + `,"elapsed":"` + time.Now().Sub(start).String() + "\"}")
		case RESP:
			if errMsg == errInvalidNumberOfArguments.Error() {
				return writeOutput("-ERR wrong number of arguments for '" + msg.Command() + "' command\r\n")
			}
			v, _ := resp.ErrorValue(errors.New("ERR " + errMsg)).MarshalRESP()
			return writeOutput(string(v))
		}
		return nil
	}

	res, err := s.command(msg)
	if res.Type() == resp.Error {
		return writeErr(res.String())
	}
	if err != nil {
		return writeErr(err.Error())
	}

	if !isRespValueEmptyString(res) {
		var resStr string
		resStr, err := serializeOutput(res)
		if err != nil {
			return err
		}
		if err := writeOutput(resStr); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) command(msg *Message) (res resp.Value, err error) {
	switch msg.Command() {
	default:
		err = fmt.Errorf("unknown command '%s'", msg.Args[0])
	case "set":
		res, err = s.cmdSet(msg)
	case "scan":
		res, err = s.cmdScan(msg)
	case "within":
		// within is the same as scan for sharding. These requests are handled in each shard later.
		res, err = s.cmdScan(msg)
	}
	return
}
