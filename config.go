package stegobot

import (
	"cmp"
	"os"
)

type Config struct {
	APIKey string
}

func NewConfig() *Config {
	return &Config{
		APIKey: "",
	}
}

func load(key, def string) string {
	return cmp.Or(os.Getenv(key), def)
}
