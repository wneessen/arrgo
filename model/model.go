package model

import (
	"database/sql"
	"errors"
	"time"
)

// SQLTimeout is the default timeout for SQL queries
const SQLTimeout = time.Second * 1

// List of model specific errors
var (
	// ErrGuildNotExistant should be used in case a requested guild was not found in the database
	ErrGuildNotExistant = errors.New("requested guild not existant in database")
)

// Model is a collection of all available models
type Model struct {
	Guild *GuildModel
}

// New returns the collection of all available models
func New(db *sql.DB) Model {
	return Model{
		Guild: &GuildModel{DB: db},
	}
}
