package server

import (
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/resp"
)

func (s *Server) cmdScan(msg *Message) (resp.Value, error) {
	// aggregated objects
	objs := make([]resp.Value, 0)

	// aggregated count
	var count int64

	// passes the given command into selected shard
	// based on output format only two case can happen, if
	// output format is COUNT result is a single integer otherwise
	// result will be array.
	// https://github.com/tidwall/tile38/blob/master/internal/server/scanner.go
	isOutputCount := false
	ca := make([]interface{}, len(msg.Args))
	for i, arg := range msg.Args {
		if arg == "COUNT" {
			isOutputCount = true
		}
		ca[i] = arg
	}
	var cmd redis.Cmder
	if !isOutputCount {
		cmd = redis.NewSliceCmd(ca...)
	} else {
		cmd = redis.NewIntCmd(ca...)
	}

	s.nodes.Walk(func(s string, v interface{}) bool {
		log.Debugf("scan request for %s", s)
		if err := v.(*redis.Client).Process(cmd); err != nil {
			log.Errorf("scan command error on %s: %s", s, err)
			return true
		}

		if err := cmd.Err(); err != nil {
			log.Errorf("scan command error on %s: %s", s, err)
			return true
		}

		if !isOutputCount {
			res := cmd.(*redis.SliceCmd).Val()
			log.Debugf("scan response from %s: %v", s, res)

			for _, obj := range res[1].([]interface{}) {
				objs = append(objs, resp.AnyValue(obj))
			}
		} else {
			res := cmd.(*redis.IntCmd).Val()
			log.Debugf("scan response from %s: %v", s, res)

			count += res
		}

		return false
	})

	if !isOutputCount {
		results := make([]resp.Value, 0)
		results = append(results, resp.IntegerValue(len(objs)))
		results = append(results, resp.ArrayValue(objs))
		return resp.ArrayValue(results), nil
	}
	return resp.IntegerValue(int(count)), nil
}
