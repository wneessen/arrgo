package model

import (
	"context"
	"database/sql"
	"time"
)

// UserReputationModel wraps the connection pool.
type UserReputationModel struct {
	DB *sql.DB
}

// UserReputation represents the user reputation in the database
type UserReputation struct {
	ID                  int64     `json:"id"`
	UserID              int64     `json:"userId"`
	Emissary            string    `json:"emissary"`
	Motto               string    `json:"motto"`
	Rank                string    `json:"rank"`
	Level               int64     `json:"lvl"`
	Experience          int64     `json:"experience"`
	NextLevel           int64     `json:"nextLevel"`
	ExperienceNextLevel int64     `json:"experienceNextLevel"`
	TitlesTotal         int64     `json:"titlesTotal"`
	TitlesUnlocked      int64     `json:"titlesUnlocked"`
	EmblemsTotal        int64     `json:"EmblemsTotal"`
	EmblemsUnlocked     int64     `json:"EmblemsUnlocked"`
	ItemsTotal          int64     `json:"ItemsTotal"`
	ItemsUnlocked       int64     `json:"ItemsUnlocked"`
	CreateTime          time.Time `json:"createTime"`
}

/*
// GetByUserID retrieves the User details from the database based on the given User ID
func (m UserStatModel) GetByUserID(i int64) (*UserStat, error) {
	q := `SELECT id, user_id, title, gold, doubloons, ancient_coins, kraken, megalodon, chests, ships, vomit, distance, ctime
            FROM user_stats s
           WHERE s.user_id = $1
           ORDER BY id DESC
           LIMIT 1`

	var us UserStat
	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	row := m.DB.QueryRowContext(ctx, q, i)
	err := row.Scan(&us.ID, &us.UserID, &us.Title, &us.Gold, &us.Doubloons, &us.AncientCoins, &us.KrakenDefeated,
		&us.MegalodonEnounter, &us.ChestsHandedIn, &us.ShipsSunk, &us.VomittedTimes, &us.DistanceSailed,
		&us.CreateTime)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return &us, ErrUserStatNotExistent
		default:
			return &us, err
		}
	}
	return &us, nil
}

// GetByUserIDAtTime retrieves the User details from the database based on the given User ID at a specific
// point of time
func (m UserStatModel) GetByUserIDAtTime(i int64, t time.Time) (*UserStat, error) {
	q := `SELECT id, user_id, title, gold, doubloons, ancient_coins, kraken, megalodon, chests, ships, vomit, distance, ctime
            FROM user_stats s
           WHERE s.user_id = $1
             AND s.ctime >= $2
           ORDER BY id
           LIMIT 1`

	var us UserStat
	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	row := m.DB.QueryRowContext(ctx, q, i, t)
	err := row.Scan(&us.ID, &us.UserID, &us.Title, &us.Gold, &us.Doubloons, &us.AncientCoins, &us.KrakenDefeated,
		&us.MegalodonEnounter, &us.ChestsHandedIn, &us.ShipsSunk, &us.VomittedTimes, &us.DistanceSailed,
		&us.CreateTime)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return &us, ErrUserStatNotExistent
		default:
			return &us, err
		}
	}
	return &us, nil
}

*/

// Insert adds a new User into the database
func (m UserReputationModel) Insert(ur *UserReputation) error {
	q := `INSERT INTO user_reputation (user_id, emissary, motto, rank, lvl, xp, next_lvl, xp_next_lvl, titlestotal, 
                             titlesunlocked, emblemstotal, emblemsunlocked, itemstotal, itemsunlocked)
               VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
            RETURNING id, ctime`
	v := []interface{}{
		ur.UserID, ur.Emissary, ur.Motto, ur.Rank, ur.Level, ur.Experience, ur.NextLevel,
		ur.ExperienceNextLevel, ur.TitlesTotal, ur.TitlesUnlocked, ur.EmblemsTotal, ur.EmblemsUnlocked,
		ur.ItemsTotal, ur.ItemsUnlocked,
	}

	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	row := m.DB.QueryRowContext(ctx, q, v...)
	err := row.Scan(&ur.ID, &ur.CreateTime)
	if err != nil {
		return err
	}
	return nil
}
