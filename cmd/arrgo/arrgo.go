package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/wneessen/arrgo/bot"
	"github.com/wneessen/arrgo/config"
)

// CLIFlags represents the struct that is used to handle CLI flags
type CLIFlags struct {
	c string // Path to config file
	r bool   // Remove slash commands
	m bool   // Run in SQL migration mode
	d bool   // Run in SQL downgrade mode
	f bool   // First run
}

func main() {
	cf := CLIFlags{
		c: "/arrgo/etc/arrgo.toml",
	}
	if cfe := os.Getenv("ARRGO_CONFIG"); cfe != "" {
		cf.c = cfe
	}
	flag.StringVar(&cf.c, "c", cf.c, "Path to config file")
	flag.BoolVar(&cf.r, "r", cf.r, "Remove slash commands")
	flag.BoolVar(&cf.m, "migrate", false, "Execute SQL migrations before starting the bot")
	flag.BoolVar(&cf.d, "downgrade", false, "Execute SQL downgrade migrations before "+
		"starting the bot")
	flag.BoolVar(&cf.f, "firstrun", false, "Execute first-run tasks during startup")
	flag.Parse()

	// Read/Parse config
	if cf.c == "" {
		_, _ = fmt.Fprintf(os.Stderr, "no config file provided. Aborting")
		os.Exit(1)
	}
	c, err := config.New(config.WithConfFile(cf.c))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "could not read config: %s. Aborting", err)
		os.Exit(1)
	}
	if cf.f {
		c.SetFirstRun()
	}

	// Initialize zerolog
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
		ll.Error().Msgf("failed to initialize bot: %s", err)
		os.Exit(1)
	}

	// Perform SQL migrations if requested
	if cf.m {
		if err := b.SQLMigrate(&c); err != nil {
			ll.Error().Msgf("SQL migration failed: %s", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	if cf.d {
		if err := b.SQLDowngrade(&c); err != nil {
			ll.Error().Msgf("SQL downgrade failed: %s", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	if cf.r {
		ll.Info().Msg("Removing all slash commands...")
		if err := b.RemoveSlashCommands(); err != nil {
			ll.Error().Msgf("failed to remove slash commands: %s", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Check the current DB version
	dd, err := b.CheckDBVersion(&c)
	if err != nil {
		ll.Error().Msgf("failed to check database version: %s", err)
		os.Exit(1)
	}
	if dd > 0 {
		ll.Warn().Msgf("WARNING: Your database is %d version(s) behind.", dd)
		ll.Warn().Msg("Running ArrGo behind its intended DB version can cause unexpected behaviour and crashes")
		ll.Warn().Msg("Please start the Bot using the -migrate flag to update the database")
	}

	if err := b.Run(); err != nil {
		ll.Error().Msgf("failed to run bot: %s", err)
	}
}
