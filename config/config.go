package config

import (
	"fmt"
	"github.com/kkyr/fig"
	"os"
)

// Config represents the global configuration struct that the config file is marshalled into
type Config struct {
	Discord struct {
		Token   string `fig:"token"`
		ShardID int    `fig:"shard_id" default:"0"`
	}
	DB struct {
		Path string `fig:"path" validate:"required"`
	}
	Log struct {
		Level string `fig:"level" default:"info"`
	}
	confFile string
}

func New(p string) (Config, error) {
	co := Config{
		confFile: p,
	}
	if p == "" {
		return co, fmt.Errorf("config path cannot be empty")
	}
	_, err := os.Stat(p)
	if err != nil {
		return co, fmt.Errorf("config file %q not readable: %w", p, err)
	}
	if err := fig.Load(&co, fig.File(p)); err != nil {
		return co, fmt.Errorf("unable to unmarshall config: %w", err)
	}

	return co, nil
}

// ConfFilePath returns the internal path the config file for reference
func (c *Config) ConfFilePath() string {
	return c.confFile
}
