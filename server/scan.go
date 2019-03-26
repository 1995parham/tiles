package server

import (
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/resp"
)

func (s *Server) cmdScan(msg *Message) (resp.Value, error) {
	// result collection with the following forma
	results := make([]resp.Value, 0)

	// aggregated objects (scan * OBJECTS)
	objs := make([]resp.Value, 0)

	// passes the current command into selected shard
	ca := make([]interface{}, len(msg.Args))
	for i, arg := range msg.Args {
		ca[i] = arg
	}
	cmd := redis.NewSliceCmd(ca...)

	s.nodes.Walk(func(s string, v interface{}) bool {
		log.Debugf("scan request for %s", s)
		if err := v.(*redis.Client).Process(cmd); err != nil {
			log.Errorf("scan command error on %s: %s", s, err)
			return true
		}

		res, err := cmd.Result()
		if err != nil {
			log.Errorf("scan command error on %s: %s", s, err)
			return true
		}

		log.Debugf("scan response from %s: %v", s, res)

		if len(res) == 2 {
			for _, obj := range res[1].([]interface{}) {
				objs = append(objs, resp.AnyValue(obj))
			}
		}

		return false
	})

	results = append(results, resp.IntegerValue(len(objs)))
	results = append(results, resp.ArrayValue(objs))
	return resp.ArrayValue(results), nil
}
