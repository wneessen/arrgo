package model

import (
	"database/sql"
	"time"
)

// SQLTimeout is the default timeout for SQL queries
const SQLTimeout = time.Second * 1

// Model is a collection of all available models
type Model struct {
}

// New returns the collection of all available models
func New(db *sql.DB) Model {
	return Model{}
}
