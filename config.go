package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

// Config holds all configurations of application
type Config struct {
	Debug bool

	Threads   int
	Host      string
	Port      int
	KeepAlive time.Duration

	Tiles map[string]redis.Options
}

// config reads configuration with viper
func config() Config {
	var instance Config

	v := viper.New()
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	v.SetConfigName("config")

	if err := v.ReadInConfig(); err != nil {
		switch err.(type) {
		default:
			log.Fatalf("Fatal error loading config file: %s \n", err)
		case viper.ConfigFileNotFoundError:
			log.Printf("No config file found. Using defaults and environment variables")
		}
	}

	v.SetEnvPrefix("tiles")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.UnmarshalExact(&instance); err != nil {
		log.Printf("configuration: %s", err)
	}
	fmt.Printf("Following configuration is loaded:\n%+v\n", instance)

	return instance
}
