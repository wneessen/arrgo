package model

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/gob"
	"fmt"
	"github.com/wneessen/arrgo/crypto"
)

// GuildPrefKey represents a guild preferences identifier key
type GuildPrefKey string

// List of possible GuildPrefKeys
const (
	// GuildPrefScheduledFlameheart is the setting for en-/disabling the scheduled Flameheart spam
	GuildPrefScheduledFlameheart GuildPrefKey = "scheduled_fh"
)

// GetPrefString fetches a client-specific setting from the database as string type
func (m GuildModel) GetPrefString(g *Guild, k GuildPrefKey) (string, error) {
	return getPref[string](m, g, k)
}

// GetPrefStringEnc fetches an encrypted client-specific setting from the database as string type
func (m GuildModel) GetPrefStringEnc(g *Guild, k GuildPrefKey) (string, error) {
	return getPrefEnc[string](m, g, k)
}

// GetPrefInt fetches a client-specific setting from the database as string type
func (m GuildModel) GetPrefInt(g *Guild, k GuildPrefKey) (int, error) {
	return getPref[int](m, g, k)
}

// GetPrefInt64 fetches a client-specific setting from the database as string type
func (m GuildModel) GetPrefInt64(g *Guild, k GuildPrefKey) (int64, error) {
	return getPref[int64](m, g, k)
}

// GetPrefBool fetches a client-specific setting from the database as string type
func (m GuildModel) GetPrefBool(g *Guild, k GuildPrefKey) (bool, error) {
	return getPref[bool](m, g, k)
}

// PrefExists checks if a guild preference is already present in the DB
func (m GuildModel) PrefExists(g *Guild, k GuildPrefKey) (bool, error) {
	q := `SELECT COUNT(g.pref_val)
            FROM guild_prefs g
           WHERE g.guild_id = $1 AND g.pref_key = $2`

	sa := []interface{}{g.ID, k}

	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	var co int64
	row := m.DB.QueryRowContext(ctx, q, sa...)
	err := row.Scan(&co)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return false, nil
		default:
			return false, err
		}
	}
	return co > 0, nil
}

// SetPref stores a guild-specific setting in the database
func (m GuildModel) SetPref(g *Guild, k GuildPrefKey, v interface{}) error {
	var sv bytes.Buffer
	gobEnc := gob.NewEncoder(&sv)
	if err := gobEnc.Encode(v); err != nil {
		return err
	}

	q := `INSERT INTO guild_prefs (guild_id, pref_key, pref_val, is_enc) VALUES ($1, $2, $3, false)`
	sa := []interface{}{
		g.ID,
		k,
		sv.Bytes(),
	}

	pe, err := m.PrefExists(g, k)
	if err != nil {
		return err
	}
	if pe {
		q = `UPDATE guild_prefs SET pref_val = $3, mtime = NOW() WHERE pref_key = $2 AND guild_id = $1`
	}

	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	_, err = m.DB.ExecContext(ctx, q, sa...)
	if err != nil {
		return err
	}
	return nil
}

// SetPrefEnc stores an encrypted guild-specific setting in the database
func (m GuildModel) SetPrefEnc(g *Guild, k GuildPrefKey, v interface{}) error {
	var sv bytes.Buffer
	gobEnc := gob.NewEncoder(&sv)
	if err := gobEnc.Encode(v); err != nil {
		return err
	}

	ek, err := m.DecryptEncSecret(g)
	if err != nil {
		return fmt.Errorf("failed to decrypt guild encryption secret: %w", err)
	}
	ed, err := crypto.EncryptAuth(sv.Bytes(), ek, []byte(g.GuildID))
	if err != nil {
		return fmt.Errorf("failed to encrypt guild preference: %w", err)
	}

	q := `INSERT INTO guild_prefs (guild_id, pref_key, pref_val, is_enc) VALUES ($1, $2, $3, true)`
	sa := []interface{}{
		g.ID,
		k,
		ed,
	}

	pe, err := m.PrefExists(g, k)
	if err != nil {
		return err
	}
	if pe {
		q = `UPDATE guild_prefs SET pref_val = $3, mtime = NOW() WHERE pref_key = $2 AND guild_id = $1`
	}

	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	_, err = m.DB.ExecContext(ctx, q, sa...)
	if err != nil {
		return err
	}
	return nil
}

// getPref is a generic interface to fetch guild-specific settings from the database
// for different types
func getPref[V string | bool | int | int64](m GuildModel, g *Guild, k GuildPrefKey) (V, error) {
	var v V
	var bv []byte
	var ob bytes.Buffer

	q := `SELECT pref_val
            FROM guild_prefs g
           WHERE g.guild_id = $1 AND g.pref_key = $2 AND g.is_enc = false`
	sa := []interface{}{g.ID, k}

	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	row := m.DB.QueryRowContext(ctx, q, sa...)
	err := row.Scan(&bv)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return v, ErrGuildPrefNotExistant
		default:
			return v, err
		}
	}

	ob.Write(bv)
	gobDec := gob.NewDecoder(&ob)
	if err := gobDec.Decode(&v); err != nil {
		return v, err
	}
	return v, nil
}

// getPrefEnc is a generic interface to fetch encrypted guild-specific settings from the database
// for different types
func getPrefEnc[V string | bool | int | int64](m GuildModel, g *Guild, k GuildPrefKey) (V, error) {
	var v V
	var bv []byte
	var ob bytes.Buffer

	q := `SELECT pref_val
            FROM guild_prefs g
           WHERE g.guild_id = $1 AND g.pref_key = $2 AND g.is_enc = true`
	sa := []interface{}{g.ID, k}

	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	row := m.DB.QueryRowContext(ctx, q, sa...)
	err := row.Scan(&bv)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return v, ErrGuildPrefNotExistant
		default:
			return v, err
		}
	}

	ek, err := m.DecryptEncSecret(g)
	if err != nil {
		return v, fmt.Errorf("failed to decrypt guild encryption secret: %w", err)
	}
	pd, err := crypto.DecryptAuth(bv, ek, []byte(g.GuildID))
	if err != nil {
		return v, fmt.Errorf("failed to decrypt guild preference: %w", err)
	}

	ob.Write(pd)
	gobDec := gob.NewDecoder(&ob)
	if err := gobDec.Decode(&v); err != nil {
		return v, err
	}
	return v, nil
}
