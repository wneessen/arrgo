package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/wneessen/arrgo/config"
	"github.com/wneessen/arrgo/crypto"
)

// GuildModel wraps the connection pool.
type GuildModel struct {
	Config *config.Config
	DB     *sql.DB
}

// Guild represents the guild information in the database
type Guild struct {
	ID              int64     `json:"id"`
	GuildID         string    `json:"guildId"`
	GuildName       string    `json:"guildName"`
	OwnerID         string    `json:"ownerId"`
	JoinedAt        time.Time `json:"joinedAt"`
	SystemChannelID string    `json:"systemChannelID"`
	EncryptionKey   []byte    `json:"-"`
	Version         int       `json:"-"`
	CreateTime      time.Time `json:"createTime"`
	ModTime         time.Time `json:"modTime"`
}

// GetByGuildID retrieves the Guild details from the database based on the given Guild ID
func (m GuildModel) GetByGuildID(i string) (*Guild, error) {
	q := `SELECT g.id, g.guild_id, g.guild_name, g.owner_id, g.joined_at, g.system_channel, g.enc_key,
       		     g.version, g.ctime, g.mtime
            FROM guilds g
           WHERE g.guild_id = $1`

	var g Guild
	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	row := m.DB.QueryRowContext(ctx, q, i)
	err := row.Scan(&g.ID, &g.GuildID, &g.GuildName, &g.OwnerID, &g.JoinedAt, &g.SystemChannelID, &g.EncryptionKey,
		&g.Version, &g.CreateTime, &g.ModTime)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return &g, ErrGuildNotExistent
		default:
			return &g, err
		}
	}
	return &g, nil
}

// GetGuilds returns a list of all guilds registered in the database
func (m GuildModel) GetGuilds() ([]*Guild, error) {
	q := `SELECT g.guild_id
            FROM guilds g`

	var gl []*Guild
	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()
	rows, err := m.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		g, err := m.GetByGuildID(id)
		if err != nil {
			return nil, err
		}
		gl = append(gl, g)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return gl, nil
}

// Insert adds a new Guild into the database
func (m GuildModel) Insert(g *Guild) error {
	q := `INSERT INTO guilds (guild_id, guild_name, owner_id, joined_at, system_channel, enc_key)
               VALUES ($1, $2, $3, $4, $5, $6)
            RETURNING id, ctime, mtime, version`
	v := []interface{}{g.GuildID, g.GuildName, g.OwnerID, g.JoinedAt, g.SystemChannelID, g.EncryptionKey}

	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	row := m.DB.QueryRowContext(ctx, q, v...)
	err := row.Scan(&g.ID, &g.CreateTime, &g.ModTime, &g.Version)
	if err != nil {
		return err
	}
	return nil
}

// Delete removes a Guild from the database
func (m GuildModel) Delete(g *Guild) error {
	q := `DELETE FROM guilds g WHERE g.id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, q, g.ID)
	return err
}

// DecryptEncSecret decrypts the guild specific encryption secret in the DB with the global encryption key
func (m GuildModel) DecryptEncSecret(g *Guild) ([]byte, error) {
	ek, err := crypto.DecryptAuth(g.EncryptionKey, []byte(m.Config.Data.EncryptionKey), []byte(g.GuildID))
	if err != nil {
		return []byte{}, fmt.Errorf("failed to decrypt guild encryption secret: %w", err)
	}
	return ek, nil
}

// AnnouceChannel will return the dedicated annouce channel or the system channel if no alternative is
// configured in the database
func (m GuildModel) AnnouceChannel(g *Guild) string {
	ch, err := m.GetPrefString(g, GuildPrefAnnounceChannel)
	if err != nil {
		return g.SystemChannelID
	}
	return ch
}
