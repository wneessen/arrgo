package model

import (
	"context"
	"database/sql"
	"time"
)

// GuildModel wraps the connection pool.
type GuildModel struct {
	DB *sql.DB
}

// Guild represents the guild information in the database
type Guild struct {
	ID              int64     `json:"id"`
	GuildID         string    `json:"guildId"`
	GuildName       string    `json:"guildName"`
	OwnerID         string    `json:"ownerId"`
	JoinedAt        time.Time `json:"joinedAt"`
	SystemChannelID string    `json:"systemChannelID"`
	Version         int       `json:"-"`
	CreateTime      time.Time `json:"createTime"`
	ModTime         time.Time `json:"modTime"`
}

// GetByGuildID retrieves the Guild details from the database based on the given Guild ID
func (m GuildModel) GetByGuildID(i string) (*Guild, error) {
	q := `SELECT g.id, g.guild_id, g.guild_name, g.owner_id, g.joined_at, g.version, g.ctime, g.mtime
            FROM guilds g
           WHERE g.guild_id = $1`

	var g Guild
	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	row := m.DB.QueryRowContext(ctx, q, i)
	err := row.Scan(&g.ID, &g.GuildID, &g.GuildName, &g.OwnerID, &g.JoinedAt, &g.SystemChannelID,
		&g.Version, &g.CreateTime, &g.ModTime)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return &g, ErrGuildNotExistant
		default:
			return &g, err
		}
	}
	return &g, nil
}

// Insert adds a new Guild into the database
func (m GuildModel) Insert(g *Guild) error {
	q := `INSERT INTO guilds (guild_id, guild_name, owner_id, joined_at, system_channel)
               VALUES ($1, $2, $3, $4, $5)
            RETURNING id, ctime, mtime, version`
	v := []interface{}{g.GuildID, g.GuildName, g.OwnerID, g.JoinedAt, g.SystemChannelID}

	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	row := m.DB.QueryRowContext(ctx, q, v...)
	err := row.Scan(&g.ID, &g.CreateTime, &g.ModTime, &g.Version)
	if err != nil {
		return err
	}
	return nil
}
