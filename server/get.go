package server

import (
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/resp"
)

func respAnyValueWithArray(obj interface{}) resp.Value {
	if objs, ok := obj.([]interface{}); ok {
		var ress []resp.Value
		for _, obj := range objs {
			ress = append(ress, respAnyValueWithArray(obj))
		}
		return resp.ArrayValue(ress)
	}
	return resp.AnyValue(obj)
}

func (s *Server) cmdGet(msg *Message) (resp.Value, error) {
	var obj interface{}

	ca := make([]interface{}, len(msg.Args))
	for i, arg := range msg.Args {
		ca[i] = arg
	}

	cmd := redis.NewCmd(ca...)

	var shardErr error

	s.nodes.Walk(func(s string, v interface{}) bool {
		log.Debugf("get request for %s", s)
		if err := v.(*redis.Client).Process(cmd); err != nil {
			if err == redis.Nil {
				return false
			}
			log.Errorf("get command error on %s: %s", s, err)
			shardErr = err
			return true
		}

		if err := cmd.Err(); err != nil {
			if err == redis.Nil {
				return false
			}
			log.Errorf("get command error on %s: %s", s, err)
			shardErr = err
			return true
		}

		obj = cmd.Val()
		log.Debugf("get response from %s: %v", s, obj)

		return true
	})

	return respAnyValueWithArray(obj), shardErr
}
