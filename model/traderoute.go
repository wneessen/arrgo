package model

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// TradeRouteModel wraps the connection pool.
type TradeRouteModel struct {
	DB *sql.DB
}

// TradeRoute represents the trade route information in the database
type TradeRoute struct {
	ID          int64     `json:"id"`
	Outpost     string    `json:"outpost"`
	SoughtAfter string    `json:"soughtAfter"`
	Surplus     string    `json:"surplus"`
	ValidThru   time.Time `json:"validThru"`
	Version     int       `json:"-"`
	CreateTime  time.Time `json:"createTime"`
	ModTime     time.Time `json:"modTime"`
}

// GetByOutpost retrieves the TradeRoute details from the database based on the given Outpost name
func (m TradeRouteModel) GetByOutpost(o string) (*TradeRoute, error) {
	q := `SELECT t.id, t.outpost, t.sought_after, t.surplus, t.validthru,
       		     t.version, t.ctime, t.mtime
            FROM trade_routes t
           WHERE t.outpost = $1`

	var t TradeRoute
	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	row := m.DB.QueryRowContext(ctx, q, o)
	err := row.Scan(&t.ID, &t.Outpost, &t.SoughtAfter, &t.Surplus, &t.ValidThru,
		&t.Version, &t.CreateTime, &t.ModTime)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return &t, ErrTradeRouteNotExistent
		default:
			return &t, err
		}
	}
	return &t, nil
}

// GetTradeRoutes returns a list of all trade routes registered in the database
func (m TradeRouteModel) GetTradeRoutes() ([]*TradeRoute, error) {
	q := `SELECT t.outpost
            FROM trade_routes t`

	var tl []*TradeRoute
	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()
	rows, err := m.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var o string
		err := rows.Scan(&o)
		if err != nil {
			return nil, err
		}
		t, err := m.GetByOutpost(o)
		if err != nil {
			return nil, err
		}
		tl = append(tl, t)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tl, nil
}

// Insert adds a new TradeRoute into the database
func (m TradeRouteModel) Insert(t *TradeRoute) error {
	q := `INSERT INTO trade_routes (outpost, sought_after, surplus, validthru)
               VALUES ($1, $2, $3, $4)
            RETURNING id, ctime, mtime, version`
	v := []interface{}{t.Outpost, t.SoughtAfter, t.Surplus, t.ValidThru}

	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	row := m.DB.QueryRowContext(ctx, q, v...)
	err := row.Scan(&t.ID, &t.CreateTime, &t.ModTime, &t.Version)
	if err != nil {
		return err
	}
	return nil
}

// Update takes a given TradeRoute and updates the variable values in the database
func (m TradeRouteModel) Update(t *TradeRoute) error {
	q := `UPDATE trade_routes
             SET outpost = $1, sought_after = $2, surplus = $3, validthru = $4, 
                 mtime = NOW(), version = version + 1
           WHERE id = $5 AND version = $6
       RETURNING version`

	v := []interface{}{
		t.Outpost, t.SoughtAfter, t.Surplus, t.ValidThru,
		t.ID, t.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	row := m.DB.QueryRowContext(ctx, q, v...)
	err := row.Scan(&t.Version)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

// ValidThru retrieves the maximum TradeRoute valid thru date form the database
func (m TradeRouteModel) ValidThru() (time.Time, error) {
	q := `SELECT MAX(t.validthru)
            FROM trade_routes t`

	var v time.Time
	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	row := m.DB.QueryRowContext(ctx, q)
	err := row.Scan(&v)
	if err != nil {
		return v, err
	}
	return v, nil
}
