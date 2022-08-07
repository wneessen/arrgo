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

	// ErrUserPrefNotExistant should be returned in case a user preference is requested that does
	// not exist in the database
	ErrUserPrefNotExistant = errors.New("requested user preference not existant in database")

	// ErrTradeRouteNotExistant should be used in case a requested trade route was not found in the database
	ErrTradeRouteNotExistant = errors.New("requested trade route not existant in database")

	// ErrEditConflict should be used when an UPDATE to the database ran into a race-condition
	ErrEditConflict = errors.New("a conflict occured while updating data")

	// ErrUserStatNotExistant should be used in case a requested user stat was not found in the database
	ErrUserStatNotExistant = errors.New("requested user stat not existant in database")
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
