package bot

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/wneessen/arrgo/config"
	"github.com/wneessen/arrgo/model"
)

// MigrationsPath defines the path where to find the sql_migrations
const MigrationsPath = "file://sql_migrations"

const (
	// ErrMigrateCloseSourceConnection should be used when a SQL migration was not able to close the source
	ErrMigrateCloseSourceConnection = "failed to close sources connection for migrate: %s"
	// ErrMigrateCloseDBConnection should be used when migrate is unable to close the DB connection
	ErrMigrateCloseDBConnection = "failed to close DB connection for migrate: %s"
)

// OpenDB tries to connect to the SQLite file and returns the sql.DB pointer
func (b *Bot) OpenDB(c *config.Config) (*sql.DB, error) {
	dsn := getDBDSN(c)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	ctx, cf := context.WithTimeout(context.Background(), model.SQLTimeout)
	defer cf()
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// CheckDBVersion compares the DB version with the SQL migrations
func (b *Bot) CheckDBVersion(c *config.Config) (uint, error) {
	ll := b.Log.With().Str("context", "bot.CheckDBVersion").Logger()
	dsn := getDBDSN(c)
	m, err := migrate.New(MigrationsPath, dsn)
	if err != nil {
		return 0, err
	}
	cv, _, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return 0, err
	}
	defer func() {
		if serr, derr := m.Close(); serr != nil || derr != nil {
			if serr != nil {
				ll.Warn().Msgf(ErrMigrateCloseSourceConnection, serr)
			}
			if derr != nil {
				ll.Warn().Msgf(ErrMigrateCloseDBConnection, derr)
			}
		}
	}()
	var mc uint = 0
	mr := "./sql_migrations"
	fileSystem := os.DirFS(mr)
	if err := fs.WalkDir(fileSystem, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(p, ".up.sql") {
			mc++
		}
		return nil
	}); err != nil {
		return 0, err
	}
	if cv < mc {
		return mc - cv, nil
	}
	return 0, nil
}

// SQLMigrate migrates the database to the latest SQL set
func (b *Bot) SQLMigrate(c *config.Config) error {
	ll := b.Log.With().Str("context", "bot.SQLMigrate").Logger()
	dsn := getDBDSN(c)
	m, err := migrate.New(MigrationsPath, dsn)
	if err != nil {
		return err
	}
	defer func() {
		if serr, derr := m.Close(); serr != nil || derr != nil {
			if serr != nil {
				ll.Warn().Msgf(ErrMigrateCloseSourceConnection, serr)
			}
			if derr != nil {
				ll.Warn().Msgf(ErrMigrateCloseDBConnection, derr)
			}
		}
	}()
	cv, _, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return err
	}
	if err := m.Up(); err != nil {
		switch {
		case errors.Is(err, migrate.ErrNoChange):
			ll.Info().Msg("database is already on the latest version")
			return nil
		default:
			return err
		}
	}
	nv, _, err := m.Version()
	if err != nil {
		return err
	}
	ll.Info().Msgf("successfully updated database from v%d to v%d", cv, nv)
	return nil
}

// SQLDowngrade migrates the database to the latest SQL set
func (b *Bot) SQLDowngrade(c *config.Config) error {
	ll := b.Log.With().Str("context", "bot.SQLDowngrade").Logger()
	dsn := getDBDSN(c)
	m, err := migrate.New(MigrationsPath, dsn)
	if err != nil {
		return err
	}
	defer func() {
		if serr, derr := m.Close(); serr != nil || derr != nil {
			if serr != nil {
				ll.Warn().Msgf(ErrMigrateCloseSourceConnection, serr)
			}
			if derr != nil {
				ll.Warn().Msgf(ErrMigrateCloseDBConnection, derr)
			}
		}
	}()
	if err := m.Steps(-1); err != nil {
		switch {
		case errors.Is(err, migrate.ErrNoChange):
			ll.Info().Msg("database is already on the latest version")
			return nil
		default:
			return err
		}
	}
	cv, _, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return err
	}
	ll.Info().Msgf("successfully downgraded database to v%d", cv)
	return nil
}

// getDBDSN returns the DB connection string based on the given config
func getDBDSN(c *config.Config) string {
	var dp string
	if c.DB.UseTLS {
		dp += "sslmode=verify-full"
	}
	if !c.DB.UseTLS {
		dp += "sslmode=disable"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?%s", c.DB.Username, c.DB.Password, c.DB.Host,
		c.DB.Port, c.DB.Database, dp)
}
