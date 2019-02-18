package config

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

// Config holds all configurations of application
type Config struct {
	Threads int
	Host    string
	Port    int
}

var instance Config
var once sync.Once

// GetConfig returns global instance of configuration.
func GetConfig() Config {
	once.Do(func() {
		config()
	})
	return instance
}

// config reads configuration with viper
func config() {
	v := viper.New()
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.SetConfigName("config.default")

	if err := v.ReadConfig(bytes.NewReader(defaultConfig)); err != nil {
		log.Fatalf("Fatal error loading **default** config file: %s \n", err)
	}

	v.SetConfigName("config")

	if err := v.MergeInConfig(); err != nil {
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
}
