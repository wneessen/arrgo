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

	// ErrDeedNotExistent should be used in case a requested deed was not found in the database
	ErrDeedNotExistent = errors.New("requested deed not existent in database")

	// ErrDeedDuplicate should be used in case a deed to be inserted into the database already exists
	ErrDeedDuplicate = errors.New("deed already existent in database")
)

// Model is a collection of all available models
type Model struct {
	Deed       *DeedModel
	Guild      *GuildModel
	TradeRoute *TradeRouteModel
	User       *UserModel
	UserStats  *UserStatModel
}

// New returns the collection of all available models
func New(db *sql.DB, c *config.Config) Model {
	return Model{
		Deed:       &DeedModel{DB: db},
		Guild:      &GuildModel{DB: db, Config: c},
		TradeRoute: &TradeRouteModel{DB: db},
		User:       &UserModel{DB: db, Config: c},
		UserStats:  &UserStatModel{DB: db},
	}
}
