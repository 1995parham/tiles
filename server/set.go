package server

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/mmcloughlin/geohash"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/geojson"
	"github.com/tidwall/geojson/geometry"
	"github.com/tidwall/resp"
)

// parseSetArgs parses arguments of SET command. it extracts all sections of the command but only
// returns its location part.
// set command has the following format:
// SET fleet truck1 FIELD speed 90 POINT 33.5123 -112.2693
// SET fleet truck1 FIELD speed 90 FIELD age 21 POINT 33.5123 -112.2693
// SET fleet truck1 OBJECT {"type":"Point","coordinates":[33.5123,-112.2693,115]}
func (s *Server) parseSetArgs(vs []string) (geojson.Object, error) {
	// these variables only store SET command section for the prasing
	var ok bool

	var err error
	var obj geojson.Object

	var key, id string

	// read key and remove it from arguments
	if vs, key, ok = tokenval(vs); !ok || key == "" {
		return nil, errInvalidNumberOfArguments
	}

	// read id and remove it from arguments
	if vs, id, ok = tokenval(vs); !ok || id == "" {
		return nil, errInvalidNumberOfArguments
	}

	// read set options and remove them from arguments
	vs, err = parseSetArgsOptions(vs)
	if err != nil {
		return nil, err
	}

	// read set location and remove it from arguments
	vs, obj, err = parseSetArgsLocation(vs)
	if err != nil {
		return nil, err
	}

	if len(vs) != 0 {
		return nil, errInvalidNumberOfArguments
	}

	return obj, nil
}

func parseSetArgsLocation(vs []string) ([]string, geojson.Object, error) {
	var ok bool
	var typ []byte

	if vs, typ, ok = tokenvalbytes(vs); !ok || len(typ) == 0 {
		return nil, nil, errInvalidNumberOfArguments
	}

	if len(vs) == 0 {
		return nil, nil, errInvalidNumberOfArguments
	}

	switch {
	case lcb(typ, "string"):
		return nil, nil, errors.New("tiles does not support string values")
	case lcb(typ, "point"):
		var slat, slon, sz string
		if vs, slat, ok = tokenval(vs); !ok || slat == "" {
			return nil, nil, errInvalidNumberOfArguments
		}
		if vs, slon, ok = tokenval(vs); !ok || slon == "" {
			return nil, nil, errInvalidNumberOfArguments
		}
		vs, sz, ok = tokenval(vs)
		if !ok || sz == "" {
			y, err := strconv.ParseFloat(slat, 64)
			if err != nil {
				return nil, nil, errInvalidArgument(slat)
			}
			x, err := strconv.ParseFloat(slon, 64)
			if err != nil {
				return nil, nil, errInvalidArgument(slon)
			}
			return vs, geojson.NewPoint(geometry.Point{X: x, Y: y}), nil
		}
		y, err := strconv.ParseFloat(slat, 64)
		if err != nil {
			return nil, nil, errInvalidArgument(slat)
		}
		x, err := strconv.ParseFloat(slon, 64)
		if err != nil {
			return nil, nil, errInvalidArgument(slon)
		}
		z, err := strconv.ParseFloat(sz, 64)
		if err != nil {
			return nil, nil, errInvalidArgument(sz)
		}
		return vs, geojson.NewPointZ(geometry.Point{X: x, Y: y}, z), nil
	case lcb(typ, "bounds"):
		return nil, nil, errors.New("tiles does not support bounds values")
	case lcb(typ, "hash"):
		var shash string
		if vs, shash, ok = tokenval(vs); !ok || shash == "" {
			return nil, nil, errInvalidNumberOfArguments
		}
		lat, lon := geohash.Decode(shash)
		return vs, geojson.NewPoint(geometry.Point{X: lon, Y: lat}), nil
	case lcb(typ, "object"):
		var object string
		if vs, object, ok = tokenval(vs); !ok || object == "" {
			return nil, nil, errInvalidNumberOfArguments
		}

		obj, err := geojson.Parse(object, nil)
		if err != nil {
			return vs, obj, nil
		}
	}

	return nil, nil, errInvalidArgument(string(typ))
}

func parseSetArgsOptions(vs []string) ([]string, error) {
	var ok bool

	var expires *float64

	var arg []byte
	var nvs []string

	for {
		// read first argument in bytes and remote it from arguments
		// please note that argument list will be updated when there is a match
		if nvs, arg, ok = tokenvalbytes(vs); !ok || len(arg) == 0 {
			return nil, errInvalidNumberOfArguments
		}

		if lcb(arg, "field") {
			vs = nvs

			var name string
			var value string

			if vs, name, ok = tokenval(vs); !ok || name == "" {
				return nil, errInvalidNumberOfArguments
			}
			if isReservedFieldName(name) {
				return nil, errInvalidArgument(name)
			}
			if vs, value, ok = tokenval(vs); !ok || value == "" {
				return nil, errInvalidNumberOfArguments
			}
			continue
		}

		if lcb(arg, "ex") {
			vs = nvs

			if expires != nil {
				return nil, errInvalidArgument(string(arg))
			}

			var s string
			var v float64
			if vs, s, ok = tokenval(vs); !ok || s == "" {
				return nil, errInvalidNumberOfArguments
			}

			v, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return nil, errInvalidArgument(s)
			}

			expires = &v
			continue
		}

		if lcb(arg, "xx") {
			return nil, errors.New("tiles does not support xx field")
		}

		if lcb(arg, "nx") {
			return nil, errors.New("tiles does not support nx field")
		}

		break
	}

	return vs, nil
}

func (s *Server) cmdSet(msg *Message) (resp.Value, error) {
	// let's calculate the elapsed time
	start := time.Now()

	vs := msg.Args[1:]
	d, err := s.parseSetArgs(vs)
	if err != nil {
		return resp.NullValue(), err
	}

	// passes the current command into selected shard
	ca := make([]interface{}, len(msg.Args))
	for i, arg := range msg.Args {
		ca[i] = arg
	}
	cmd := redis.NewStringCmd(ca...)

	gh := geohash.Encode(d.Center().Y, d.Center().X)
	// longest prefix matching with radix tree
	m, c, ok := s.nodes.LongestPrefix(gh)
	if !ok {
		return resp.NullValue(), fmt.Errorf("there is no shard available for geohash: %s", gh)
	}
	log.Debugf("Geohash %s is matched with %s", gh, m)
	if err := c.(*redis.Client).Process(cmd); err != nil {
		return resp.NullValue(), err
	}

	switch msg.OutputType {
	default:
	case JSON:
		return resp.StringValue(`{"ok":true,"elapsed":"` + time.Since(start).String() + "\"}"), nil
	case RESP:
		return resp.SimpleStringValue("OK"), nil
	}

	return resp.NullValue(), nil
}
