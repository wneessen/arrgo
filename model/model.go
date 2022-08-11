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
	// ErrGuildNotExistent should be used in case a requested guild was not found in the database
	ErrGuildNotExistent = errors.New("requested guild not existent in database")

	// ErrUserNotExistent should be used in case a requested user was not found in the database
	ErrUserNotExistent = errors.New("requested user not existent in database")

	// ErrGuildPrefNotExistent should be returned in case a guild preference is requested that does
	// not exist in the database
	ErrGuildPrefNotExistent = errors.New("requested guild preference not existent in database")

	// ErrUserPrefNotExistent should be returned in case a user preference is requested that does
	// not exist in the database
	ErrUserPrefNotExistent = errors.New("requested user preference not existent in database")

	// ErrTradeRouteNotExistent should be used in case a requested trade route was not found in the database
	ErrTradeRouteNotExistent = errors.New("requested trade route not existent in database")

	// ErrEditConflict should be used when an UPDATE to the database ran into a race-condition
	ErrEditConflict = errors.New("a conflict occurred while updating data")

	// ErrUserStatNotExistent should be used in case a requested user stat was not found in the database
	ErrUserStatNotExistent = errors.New("requested user stat not existent in database")
)

// Model is a collection of all available models
type Model struct {
	Guild      *GuildModel
	User       *UserModel
	UserStats  *UserStatModel
	TradeRoute *TradeRouteModel
}

// New returns the collection of all available models
func New(db *sql.DB, c *config.Config) Model {
	return Model{
		Guild:      &GuildModel{DB: db, Config: c},
		User:       &UserModel{DB: db, Config: c},
		UserStats:  &UserStatModel{DB: db},
		TradeRoute: &TradeRouteModel{DB: db},
	}
}
