package main

import (
	"flag"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/wneessen/arrgo/bot"
	"github.com/wneessen/arrgo/config"
	"os"
	"strings"
	"time"
)

// ODQzODM3MjEwNTIwNTg0MjIz.GtI21P.pz_rm10f7y_8SiGdvHFR5X4lenRoFnfEqwTFaU
func main() {
	cf := "arrgo.toml"
	if cfe := os.Getenv("ARRGO_CONFIG"); cfe != "" {
		cf = cfe
	}
	flag.StringVar(&cf, "c", cf, "Path to config file")
	flag.Parse()

	// Read/Parse config
	if cf == "" {
		_, _ = fmt.Fprintf(os.Stderr, "no config file provided. Aborting")
		os.Exit(1)
	}
	c, err := config.New(cf)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "could not read config: %s. Aborting", err)
		os.Exit(1)
	}

	// Initalize zerolog
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = time.RFC3339Nano
	switch strings.ToLower(c.Log.Level) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}
	l := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("config_file", c.ConfFilePath()).Logger()
	ll := l.With().Str("context", "main").Logger()
	ll.Debug().Msg("Starting up...")

	b, err := bot.New(l, &c)
	if err != nil {
		ll.Error().Msgf("failed to initalize bot: %s", err)
		os.Exit(1)
	}

	if err := b.Run(); err != nil {
		ll.Error().Msgf("failed to run bot: %s", err)
	}
}
