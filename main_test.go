package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/go-redis/redis"
)

// TestClient tests all functionally of tiles with a client.
// Please note that this function tests are written based on the following shards:
// shard_1: tn (Tehran)
// shard_2: tnke9w (Velenjak)
func TestClient(t *testing.T) {
	go main()

	time.Sleep(1 * time.Second) // wait for server to get ready

	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:1372",
	})

	// SET
	scmd := redis.NewStringCmd("SET", "fleet", "truck", "POINT", 35.8061991, 51.398658)
	assert.NoError(t, client.Process(scmd))
	v, err := scmd.Result()
	assert.NoError(t, err)
	t.Log(v)

	// Pipeline
	pipe := client.Pipeline()
	assert.NoError(t, pipe.Process(
		redis.NewStringCmd("SET", "fleet", "car-1", "FIELD", "ID", "1", "POINT", 35.8061991, 51.398658),
	))
	assert.NoError(t, pipe.Process(
		redis.NewStringCmd("SET", "fleet", "car-2", "FIELD", "ID", "2", "POINT", 35.7037415, 51.4054816),
	))
	cmds, err := pipe.Exec()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(cmds))

	// SCAN
	sccmd := redis.NewIntCmd("SCAN", "fleet", "COUNT")
	assert.NoError(t, client.Process(sccmd))
	c, err := sccmd.Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(3), c)

	// WITHIN
	wcmd := redis.NewIntCmd("WITHIN", "fleet", "COUNT", "BOUNDS", 35.7017561, 51.4043683, 35.7034704, 51.4080062)
	assert.NoError(t, client.Process(wcmd))
	w, err := wcmd.Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), w)

}
