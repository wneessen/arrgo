package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kkyr/fig"
)

// CfgOpt is a overloading function for the New() method
type CfgOpt func(parm *ConfParm)

// ConfParm sets some overridable parameters for the config file parsing
type ConfParm struct {
	cf string // represents the config file name
	cp string // represents the config path
}

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
		Password string `fig:"pass"`
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
	Timer struct {
		FHSpam   int           `fig:"flameheart_spam" default:"60"`
		TRUpdate time.Duration `fig:"traderoutes_update" default:"12h"`
		USUpdate time.Duration `fig:"userstats_update" default:"6h"`
		RCCheck  time.Duration `fig:"ratcookie_check" default:"6h"`
		DDUpdate time.Duration `fig:"dailydeed_update" default:"24h"`
		ULUpdate time.Duration `fig:"userledger_update" default:"6h"`
	}
	confPath string
	confFile string
	firstRun bool
}

// WithConfFile overrides the default config file path/name
func WithConfFile(p string) CfgOpt {
	return func(c *ConfParm) {
		cf := filepath.Base(p)
		cp := filepath.Dir(p)
		c.cf = cf
		c.cp = cp
	}
}

func New(ol ...CfgOpt) (Config, error) {
	cp := ConfParm{
		cf: "arrgo.toml",
		cp: "/arrgo/etc",
	}
	for _, o := range ol {
		if o == nil {
			continue
		}
		o(&cp)
	}
	_, err := os.Stat(fmt.Sprintf("%s/%s", cp.cp, cp.cf))
	if err != nil {
		return Config{}, fmt.Errorf("config file %q not readable: %w",
			fmt.Sprintf("%s/%s", cp.cp, cp.cf), err)
	}
	co := Config{}
	if err := fig.Load(&co, fig.Dirs(cp.cp), fig.File(cp.cf)); err != nil {
		return co, fmt.Errorf("unable to unmarshall config: %w", err)
	}
	co.confPath = cp.cp
	co.confFile = cp.cf

	return co, nil
}

// ConfFilePath returns the internal path the config file for reference
func (c *Config) ConfFilePath() string {
	return c.confFile
}

// SetFirstRun sets the fristRun flag in the config to true
func (c *Config) SetFirstRun() {
	c.firstRun = true
}

// GetFirstRun returns the state of the firstRun flag in the config
func (c *Config) GetFirstRun() bool {
	return c.firstRun
}
