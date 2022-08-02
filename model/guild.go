package model

import (
	"context"
	"database/sql"
	"fmt"
)

// GuildModel wraps the connection pool.
type GuildModel struct {
	DB *sql.DB
}

// Guild represents the guild information in the database
type Guild struct {
	ID         int64  `json:"id"`
	GuildID    string `json:"guildId"`
	GuildName  string `json:"guildName"`
	OwnerID    string `json:"ownerId"`
	Version    int    `json:"-"`
	CreateTime string `json:"createTime"`
	ModTime    string `json:"modTime"`
}

// GetByGuildID retrieves the Guild details from the database based on the given Guild ID
func (m GuildModel) GetByGuildID(i string) (*Guild, error) {
	q := `SELECT g.id, g.guild_id, g.guild_name, g.owner_id, g.version, g.ctime, g.mtime
            FROM guilds g
           WHERE g.guild_id = $1`

	var g Guild
	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return &g, err
	}
	row := tx.QueryRowContext(ctx, q, i)
	err = row.Scan(&g.ID, &g.GuildID, &g.GuildName, &g.OwnerID, &g.Version, &g.CreateTime, &g.ModTime)
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			return &g, fmt.Errorf("select failed: %s, rollback failed: %w", err, txErr)
		}
		switch err {
		case sql.ErrNoRows:
			return &g, ErrGuildNotExistant
		default:
			return &g, err
		}
	}
	if err := tx.Commit(); err != nil {
		return &g, err
	}

	return &g, nil
}

// Insert adds a new Guild into the database
func (m GuildModel) Insert(g *Guild) error {
	q := `INSERT INTO guilds (guild_id, guild_name, owner_id)
               VALUES ($1, $2, $3)
            RETURNING id, ctime, mtime, version`
	v := []interface{}{g.GuildID, g.GuildName, g.OwnerID}

	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	row := tx.QueryRowContext(ctx, q, v...)
	err = row.Scan(&g.ID, &g.CreateTime, &g.ModTime, &g.Version)
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			return fmt.Errorf("insert failed: %s, rollback failed: %w", err, txErr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
