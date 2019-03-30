package server

import (
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/resp"
)

func (s *Server) cmdGet(msg *Message) (resp.Value, error) {
	// aggregated objects
	objs := make([]resp.Value, 0)

	ca := make([]interface{}, len(msg.Args))
	for i, arg := range msg.Args {
		ca[i] = arg
	}

	cmd := redis.NewSliceCmd(ca...)

	var shardErr error

	s.nodes.Walk(func(s string, v interface{}) bool {
		log.Debugf("scan request for %s", s)
		if err := v.(*redis.Client).Process(cmd); err != nil {
			if err == redis.Nil {
				return false
			}
			log.Errorf("scan command error on %s: %s", s, err)
			shardErr = err
			return true
		}

		if err := cmd.Err(); err != nil {
			if err == redis.Nil {
				return false
			}
			log.Errorf("scan command error on %s: %s", s, err)
			shardErr = err
			return true
		}

		res := cmd.Val()
		log.Debugf("scan response from %s: %v", s, res)

		for _, obj := range res {
			objs = append(objs, resp.AnyValue(obj))
		}

		return true
	})

	return resp.ArrayValue(objs), shardErr
}
