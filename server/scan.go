package server

import (
	"fmt"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/resp"
)

func (s *Server) scanProcess(cmd redis.Cmder) error {
	fmt.Println(cmd)
	return nil
}

func (s *Server) cmdScan(msg *Message) (resp.Value, error) {
	// result collection
	results := make([]resp.Value, 0)

	// passes the current command into selected shard
	ca := make([]interface{}, len(msg.Args))
	for i, arg := range msg.Args {
		ca[i] = arg
	}
	cmd := redis.NewScanCmd(s.scanProcess, ca...)

	s.nodes.Walk(func(s string, v interface{}) bool {
		log.Debugf("scan request for %s", s)
		if err := v.(*redis.Client).Process(cmd); err != nil {
			log.Errorf("scan command error on %s: %s", s, err)
			return true
		}

		keys, _, err := cmd.Result()
		if err != nil {
			log.Errorf("scan command error on %s: %s", s, err)
			return true
		}
		results = append(results, resp.AnyValue(keys))

		return false
	})

	return resp.ArrayValue(results), nil
}
