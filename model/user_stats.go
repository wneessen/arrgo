package model

import (
	"context"
	"database/sql"
	"time"
)

// UserStatModel wraps the connection pool.
type UserStatModel struct {
	DB *sql.DB
}

// UserStat represents the user statistics in the database
type UserStat struct {
	ID                int64     `json:"id"`
	UserID            int64     `json:"userId"`
	Title             string    `json:"title"`
	Gold              int64     `json:"gold"`
	Doubloons         int64     `json:"doubloons"`
	AncientCoins      int64     `json:"ancientCoins"`
	KrakenDefeated    int64     `json:"krakenDefeated"`
	MegalodonEnounter int64     `json:"megalodonEnounter"`
	ChestsHandedIn    int64     `json:"chestsHandedIn"`
	ShipsSunk         int64     `json:"shipsSunk"`
	VomittedTimes     int64     `json:"vomittedTimes"`
	DistanceSailed    int64     `json:"distanceSailed"`
	CreateTime        time.Time `json:"createTime"`
}

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
		switch err {
		case sql.ErrNoRows:
			return &us, ErrUserStatNotExistant
		default:
			return &us, err
		}
	}
	return &us, nil
}

// Insert adds a new User into the database
func (m UserStatModel) Insert(us *UserStat) error {
	q := `INSERT INTO user_stats (user_id, title, gold, doubloons, ancient_coins, kraken, megalodon, 
                        chests, ships, vomit, distance)
               VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
            RETURNING id, ctime`
	v := []interface{}{us.UserID, us.Title, us.Gold, us.Doubloons, us.AncientCoins, us.KrakenDefeated,
		us.MegalodonEnounter, us.ChestsHandedIn, us.ShipsSunk, us.VomittedTimes, us.DistanceSailed}

	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	row := m.DB.QueryRowContext(ctx, q, v...)
	err := row.Scan(&us.ID, &us.CreateTime)
	if err != nil {
		return err
	}
	return nil
}
