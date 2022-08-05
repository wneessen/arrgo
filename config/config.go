package config

import (
	"fmt"
	"github.com/kkyr/fig"
	"os"
)

// Some crypto defaults
const (
	// CryptoKeyLen defines the required length of cryptographic keys
	CryptoKeyLen = 32
)

// Config represents the global configuration struct that the config file is marshalled into
type Config struct {
	Discord struct {
		Token   string `fig:"token"`
		ShardID int    `fig:"shard_id" default:"0"`
	}
	DB struct {
		Host     string `fig:"host" validate:"required"`
		Username string `fig:"user" default:"arrgo"`
		Password string `fig:"password"`
		Database string `fig:"db" default:"arrgo"`
		UseTLS   bool   `fig:"use_tls"`
		Port     int    `fig:"port" default:"5432"`
	}
	Log struct {
		Level string `fig:"level" default:"info"`
	}
	Data struct {
		EncryptionKey string `fig:"enc_key"`
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
