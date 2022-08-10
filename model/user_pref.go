package model

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/gob"
	"fmt"
	"github.com/wneessen/arrgo/crypto"
)

// UserPrefKey represents a user preferences identifier key
type UserPrefKey string

// List of possible UserPrefKeys
const (
	// UserPrefSoTAuthToken is the authentication token for Sea of Thieves
	UserPrefSoTAuthToken           UserPrefKey = "rat_token"
	UserPrefSoTAuthTokenExpiration UserPrefKey = "rat_token_expire"
	UserPrefSoTAuthTokenNotified   UserPrefKey = "rat_expiry_notified"
	UserPrefPlaysSoT               UserPrefKey = "plays_sot"
	UserPrefPlaysSoTStartTime      UserPrefKey = "plays_sot_start"
)

// GetPrefString fetches a client-specific setting from the database as string type
func (m UserModel) GetPrefString(g *User, k UserPrefKey) (string, error) {
	return getUserPref[string](m, g, k)
}

// GetPrefStringEnc fetches an encrypted client-specific setting from the database as string type
func (m UserModel) GetPrefStringEnc(g *User, k UserPrefKey) (string, error) {
	return getUserPrefEnc[string](m, g, k)
}

// GetPrefInt fetches a client-specific setting from the database as string type
func (m UserModel) GetPrefInt(g *User, k UserPrefKey) (int, error) {
	return getUserPref[int](m, g, k)
}

// GetPrefIntEnc fetches an encrypted client-specific setting from the database as string type
func (m UserModel) GetPrefIntEnc(g *User, k UserPrefKey) (int, error) {
	return getUserPrefEnc[int](m, g, k)
}

// GetPrefInt64 fetches a client-specific setting from the database as string type
func (m UserModel) GetPrefInt64(g *User, k UserPrefKey) (int64, error) {
	return getUserPref[int64](m, g, k)
}

// GetPrefInt64Enc fetches an encrypted client-specific setting from the database as string type
func (m UserModel) GetPrefInt64Enc(g *User, k UserPrefKey) (int64, error) {
	return getUserPrefEnc[int64](m, g, k)
}

// GetPrefBool fetches a client-specific setting from the database as string type
func (m UserModel) GetPrefBool(g *User, k UserPrefKey) (bool, error) {
	return getUserPref[bool](m, g, k)
}

// GetPrefBoolEnc fetches an encrypted client-specific setting from the database as string type
func (m UserModel) GetPrefBoolEnc(g *User, k UserPrefKey) (bool, error) {
	return getUserPrefEnc[bool](m, g, k)
}

// PrefExists checks if a user preference is already present in the DB
func (m UserModel) PrefExists(u *User, k UserPrefKey) (bool, error) {
	q := `SELECT COUNT(u.pref_val)
            FROM user_prefs u
           WHERE u.user_id = $1 AND u.pref_key = $2`

	sa := []interface{}{u.ID, k}

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

// SetPref stores a user-specific setting in the database
func (m UserModel) SetPref(u *User, k UserPrefKey, v interface{}) error {
	var sv bytes.Buffer
	gobEnc := gob.NewEncoder(&sv)
	if err := gobEnc.Encode(v); err != nil {
		return err
	}

	q := `INSERT INTO user_prefs (user_id, pref_key, pref_val, is_enc) VALUES ($1, $2, $3, false)`
	sa := []interface{}{
		u.ID,
		k,
		sv.Bytes(),
	}

	pe, err := m.PrefExists(u, k)
	if err != nil {
		return err
	}
	if pe {
		q = `UPDATE user_prefs SET pref_val = $3, mtime = NOW() WHERE pref_key = $2 AND user_id = $1`
	}

	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	_, err = m.DB.ExecContext(ctx, q, sa...)
	if err != nil {
		return err
	}
	return nil
}

// SetPrefEnc stores an encrypted user-specific setting in the database
func (m UserModel) SetPrefEnc(u *User, k UserPrefKey, v interface{}) error {
	var sv bytes.Buffer
	gobEnc := gob.NewEncoder(&sv)
	if err := gobEnc.Encode(v); err != nil {
		return err
	}

	ek, err := m.DecryptEncSecret(u)
	if err != nil {
		return fmt.Errorf("failed to decrypt user encryption secret: %w", err)
	}
	ed, err := crypto.EncryptAuth(sv.Bytes(), ek, []byte(u.UserID))
	if err != nil {
		return fmt.Errorf("failed to encrypt user preference: %w", err)
	}

	q := `INSERT INTO user_prefs (user_id, pref_key, pref_val, is_enc) VALUES ($1, $2, $3, true)`
	sa := []interface{}{
		u.ID,
		k,
		ed,
	}

	pe, err := m.PrefExists(u, k)
	if err != nil {
		return err
	}
	if pe {
		q = `UPDATE user_prefs SET pref_val = $3, mtime = NOW() WHERE pref_key = $2 AND user_id = $1`
	}

	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	_, err = m.DB.ExecContext(ctx, q, sa...)
	if err != nil {
		return err
	}
	return nil
}

// getUserPref is a generic interface to fetch user-specific settings from the database
// for different types
func getUserPref[V string | bool | int | int64](m UserModel, u *User, k UserPrefKey) (V, error) {
	var v V
	var bv []byte
	var ob bytes.Buffer

	q := `SELECT pref_val
            FROM user_prefs u
           WHERE u.user_id = $1 AND u.pref_key = $2 AND u.is_enc = false`
	sa := []interface{}{u.ID, k}

	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	row := m.DB.QueryRowContext(ctx, q, sa...)
	err := row.Scan(&bv)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return v, ErrUserPrefNotExistant
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

// getUserPrefEnc is a generic interface to fetch encrypted user-specific settings from the database
// for different types
func getUserPrefEnc[V string | bool | int | int64](m UserModel, u *User, k UserPrefKey) (V, error) {
	var v V
	var bv []byte
	var ob bytes.Buffer

	q := `SELECT pref_val
            FROM user_prefs u
           WHERE u.user_id = $1 AND u.pref_key = $2 AND u.is_enc = true`
	sa := []interface{}{u.ID, k}

	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	row := m.DB.QueryRowContext(ctx, q, sa...)
	err := row.Scan(&bv)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return v, ErrUserPrefNotExistant
		default:
			return v, err
		}
	}

	ek, err := m.DecryptEncSecret(u)
	if err != nil {
		return v, fmt.Errorf("failed to decrypt user encryption secret: %w", err)
	}
	pd, err := crypto.DecryptAuth(bv, ek, []byte(u.UserID))
	if err != nil {
		return v, fmt.Errorf("failed to decrypt user preference: %w", err)
	}

	ob.Write(pd)
	gobDec := gob.NewDecoder(&ob)
	if err := gobDec.Decode(&v); err != nil {
		return v, err
	}
	return v, nil
}
