package main

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Shards holds the map between shards (their connection configuration) and geo-hash
type Shards map[string]redis.Options

func shards() Shards {
	var instance Shards

	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName("shards")

	v.AddConfigPath("/etc/tiles")
	v.AddConfigPath(".")

	if err := v.MergeInConfig(); err != nil {
		logrus.Fatalf("Fatal error loading shards file: %s \n", err)
	}

	if err := v.UnmarshalExact(&instance); err != nil {
		logrus.Errorf("configuration: %s", err)
	}
	fmt.Printf("Following configuration is loaded:\n%+v\n", instance)

	return instance
}

// Config holds all configurations of application
type Config struct {
	Debug bool

	Threads   int
	Host      string
	Port      int
	KeepAlive time.Duration
}

// config reads configuration with viper
func config() Config {
	var defaultConfig = []byte(`
### configuration is in the YAML format
### and it use 2-space as tab.
debug: true
host: 0.0.0.0
port: 1372
threads: 0
keepAlive: 0s
`)

	var instance Config

	v := viper.New()
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	if err := v.ReadConfig(bytes.NewReader(defaultConfig)); err != nil {
		logrus.Fatalf("Fatal error loading **default** config array: %s \n", err)
	}

	v.SetConfigName("config")

	if err := v.MergeInConfig(); err != nil {
		switch err.(type) {
		default:
			logrus.Fatalf("Fatal error loading config file: %s \n", err)
		case viper.ConfigFileNotFoundError:
			logrus.Errorf("No config file found. Using defaults and environment variables")
		}
	}

	v.SetEnvPrefix("tiles")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.UnmarshalExact(&instance); err != nil {
		logrus.Errorf("configuration: %s", err)
	}
	fmt.Printf("Following configuration is loaded:\n%+v\n", instance)

	return instance
}
