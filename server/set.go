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
func (s *Server) parseSetArgs(vs []string) (obj geojson.Object, err error) {
	// these variables only store SET command section for the prasing
	var fields []string
	var values []float64
	var expires *float64

	var ok bool
	var typ []byte

	var key, id string

	if vs, key, ok = tokenval(vs); !ok || key == "" {
		err = errInvalidNumberOfArguments
		return
	}

	if vs, id, ok = tokenval(vs); !ok || id == "" {
		err = errInvalidNumberOfArguments
		return
	}

	var arg []byte
	var nvs []string
	for {
		if nvs, arg, ok = tokenvalbytes(vs); !ok || len(arg) == 0 {
			err = errInvalidNumberOfArguments
			return
		}
		if lcb(arg, "field") {
			vs = nvs
			var name string
			var svalue string
			var value float64
			if vs, name, ok = tokenval(vs); !ok || name == "" {
				err = errInvalidNumberOfArguments
				return
			}
			if isReservedFieldName(name) {
				err = errInvalidArgument(name)
				return
			}
			if vs, svalue, ok = tokenval(vs); !ok || svalue == "" {
				err = errInvalidNumberOfArguments
				return
			}
			value, err = strconv.ParseFloat(svalue, 64)
			if err != nil {
				err = errInvalidArgument(svalue)
				return
			}
			fields = append(fields, name)
			values = append(values, value)
			continue
		}
		if lcb(arg, "ex") {
			vs = nvs
			if expires != nil {
				err = errInvalidArgument(string(arg))
				return
			}
			var s string
			var v float64
			if vs, s, ok = tokenval(vs); !ok || s == "" {
				err = errInvalidNumberOfArguments
				return
			}
			v, err = strconv.ParseFloat(s, 64)
			if err != nil {
				err = errInvalidArgument(s)
				return
			}
			expires = &v
			continue
		}
		if lcb(arg, "xx") {
			err = errors.New("tiles does not support xx field")
			return
		}
		if lcb(arg, "nx") {
			err = errors.New("tiles does not support nx field")
			return
		}
		break
	}
	if vs, typ, ok = tokenvalbytes(vs); !ok || len(typ) == 0 {
		err = errInvalidNumberOfArguments
		return
	}
	if len(vs) == 0 {
		err = errInvalidNumberOfArguments
		return
	}

	switch {
	default:
		err = errInvalidArgument(string(typ))
		return
	case lcb(typ, "string"):
		err = errors.New("tiles does not support string values")
		return
	case lcb(typ, "point"):
		var slat, slon, sz string
		if vs, slat, ok = tokenval(vs); !ok || slat == "" {
			err = errInvalidNumberOfArguments
			return
		}
		if vs, slon, ok = tokenval(vs); !ok || slon == "" {
			err = errInvalidNumberOfArguments
			return
		}
		vs, sz, ok = tokenval(vs)
		if !ok || sz == "" {
			var x, y float64
			y, err = strconv.ParseFloat(slat, 64)
			if err != nil {
				err = errInvalidArgument(slat)
				return
			}
			x, err = strconv.ParseFloat(slon, 64)
			if err != nil {
				err = errInvalidArgument(slon)
				return
			}
			obj = geojson.NewPoint(geometry.Point{X: x, Y: y})
		} else {
			var x, y, z float64
			y, err = strconv.ParseFloat(slat, 64)
			if err != nil {
				err = errInvalidArgument(slat)
				return
			}
			x, err = strconv.ParseFloat(slon, 64)
			if err != nil {
				err = errInvalidArgument(slon)
				return
			}
			z, err = strconv.ParseFloat(sz, 64)
			if err != nil {
				err = errInvalidArgument(sz)
				return
			}
			obj = geojson.NewPointZ(geometry.Point{X: x, Y: y}, z)
		}
	case lcb(typ, "bounds"):
		err = errors.New("tiles does not support bounds values")
		return
	case lcb(typ, "hash"):
		var shash string
		if vs, shash, ok = tokenval(vs); !ok || shash == "" {
			err = errInvalidNumberOfArguments
			return
		}
		lat, lon := geohash.Decode(shash)
		obj = geojson.NewPoint(geometry.Point{X: lon, Y: lat})
	case lcb(typ, "object"):
		var object string
		if vs, object, ok = tokenval(vs); !ok || object == "" {
			err = errInvalidNumberOfArguments
			return
		}
		obj, err = geojson.Parse(object, nil)
		if err != nil {
			return
		}
	}
	if len(vs) != 0 {
		err = errInvalidNumberOfArguments
	}

	return
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
		return resp.StringValue(`{"ok":true,"elapsed":"` + time.Now().Sub(start).String() + "\"}"), nil
	case RESP:
		return resp.SimpleStringValue("OK"), nil
	}

	return resp.NullValue(), nil
}
