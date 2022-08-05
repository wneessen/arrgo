package model

import (
	"database/sql"
	"errors"
	"github.com/wneessen/arrgo/config"
	"time"
)

// SQLTimeout is the default timeout for SQL queries
const SQLTimeout = time.Second * 1

// List of model specific errors
var (
	// ErrGuildNotExistant should be used in case a requested guild was not found in the database
	ErrGuildNotExistant = errors.New("requested guild not existant in database")

	// ErrUserNotExistant should be used in case a requested user was not found in the database
	ErrUserNotExistant = errors.New("requested user not existant in database")

	// ErrGuildPrefNotExistant should be returned in case a guild preference is requested that does
	// not exist in the database
	ErrGuildPrefNotExistant = errors.New("requested guild preference not existant in database")
)

// Model is a collection of all available models
type Model struct {
	Guild *GuildModel
	User  *UserModel
}

// New returns the collection of all available models
func New(db *sql.DB, c *config.Config) Model {
	return Model{
		Guild: &GuildModel{DB: db, Config: c},
		User:  &UserModel{DB: db},
	}
}
