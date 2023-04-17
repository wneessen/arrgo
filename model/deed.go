package model

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// Different deed types
const (
	DeedTypeStandard      DeedType = "standard"
	DeedTypeDailyStandard DeedType = "daily_standard"
	DeedTypeDailySwift    DeedType = "daily_swift"
	DeedTypeUnknown       DeedType = "unknown"
)

// Different reward types
const (
	RewardGold      RewardType = "gold"
	RewardDoubloons RewardType = "doubloons"
	// RewardUnknown   RewardType = "unknown"
)

// DeedModel wraps the connection pool.
type DeedModel struct {
	DB *sql.DB
}

// DeedType is a wrapper for a string
type DeedType string

// RewardType is a wrapper for a string
type RewardType string

// Deed represents the deed information in the database
type Deed struct {
	ID           int64      `json:"id"`
	DeedType     DeedType   `json:"deedType"`
	Description  string     `json:"description"`
	ValidFrom    time.Time  `json:"validFrom"`
	ValidThru    time.Time  `json:"validThru"`
	RewardType   RewardType `json:"rewardType"`
	RewardAmount int        `json:"rewardAmount"`
	RewardIcon   string     `json:"rewardIcon"`
	ImageURL     string     `json:"imageURL"`
	CreateTime   time.Time  `json:"createTime"`
}

// GetByDeedID retrieves the Deed details from the database based on the given Deed ID
func (m DeedModel) GetByDeedID(i int64) (*Deed, error) {
	q := `SELECT d.id, d.deed_type, d.description, d.valid_from, d.valid_thru, d.reward_type, d.reward_amount,
       d.reward_icon, d.image_url, d.ctime
            FROM deeds d
           WHERE d.id = $1`

	var d Deed
	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	row := m.DB.QueryRowContext(ctx, q, i)
	err := row.Scan(&d.ID, &d.DeedType, &d.Description, &d.ValidFrom, &d.ValidThru, &d.RewardType,
		&d.RewardAmount, &d.RewardIcon, &d.ImageURL, &d.CreateTime)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return &d, ErrDeedNotExistent
		default:
			return &d, err
		}
	}
	return &d, nil
}

// GetByDeedsAtTime retrieves the list of Deed details from the database based on a given time
func (m DeedModel) GetByDeedsAtTime(t time.Time) ([]*Deed, error) {
	q := `SELECT d.id
            FROM deeds d
           WHERE d.valid_from <= $1
             AND d.valid_thru >= $1`

	var dl []*Deed
	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()
	rows, err := m.DB.QueryContext(ctx, q, t)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var id int64
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		d, err := m.GetByDeedID(id)
		if err != nil {
			return nil, err
		}
		dl = append(dl, d)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return dl, nil
}

// Insert adds a new Guild into the database
func (m DeedModel) Insert(d *Deed) error {
	q := `INSERT INTO deeds (deed_type, description, valid_from, valid_thru, reward_type, reward_amount, 
                   reward_icon, image_url)
               VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
            RETURNING id, ctime`
	v := []interface{}{
		d.DeedType, d.Description, d.ValidFrom, d.ValidThru, d.RewardType, d.RewardAmount,
		d.RewardIcon, d.ImageURL,
	}

	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	row := m.DB.QueryRowContext(ctx, q, v...)
	err := row.Scan(&d.ID, &d.CreateTime)
	if err != nil {
		switch err.Error() {
		case `pq: duplicate key value violates unique constraint "deeds_deed_type_valid_from_valid_thru_key"`:
			return ErrDeedDuplicate
		default:
			return err
		}
	}
	return nil
}
