package model

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/wneessen/arrgo/config"
	"github.com/wneessen/arrgo/crypto"
	"time"
)

// UserModel wraps the connection pool.
type UserModel struct {
	Config *config.Config
	DB     *sql.DB
}

// User represents the user information in the database
type User struct {
	ID            int64     `json:"id"`
	UserID        string    `json:"userId"`
	EncryptionKey []byte    `json:"-"`
	Version       int       `json:"-"`
	CreateTime    time.Time `json:"createTime"`
	ModTime       time.Time `json:"modTime"`
}

// GetByUserID retrieves the User details from the database based on the given User ID
func (m UserModel) GetByUserID(i string) (*User, error) {
	q := `SELECT u.id, u.user_id, u.enc_key, u.version, u.ctime, u.mtime
            FROM users u
           WHERE u.user_id = $1`

	var u User
	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	row := m.DB.QueryRowContext(ctx, q, i)
	err := row.Scan(&u.ID, &u.UserID, &u.EncryptionKey, &u.Version, &u.CreateTime, &u.ModTime)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return &u, ErrUserNotExistant
		default:
			return &u, err
		}
	}
	return &u, nil
}

// GetUsers returns a list of all users registered in the database
func (m UserModel) GetUsers() ([]*User, error) {
	q := `SELECT u.user_id
            FROM users u`

	var ul []*User
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
		g, err := m.GetByUserID(id)
		if err != nil {
			return nil, err
		}
		ul = append(ul, g)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return ul, nil
}

// Insert adds a new User into the database
func (m UserModel) Insert(u *User) error {
	q := `INSERT INTO users (user_id, enc_key)
               VALUES ($1, $2)
            RETURNING id, ctime, mtime, version`
	v := []interface{}{u.UserID, u.EncryptionKey}

	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	row := m.DB.QueryRowContext(ctx, q, v...)
	err := row.Scan(&u.ID, &u.CreateTime, &u.ModTime, &u.Version)
	if err != nil {
		return err
	}
	return nil
}

// Delete removes a User from the database
func (m UserModel) Delete(u *User) error {
	q := `DELETE FROM users u WHERE u.id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, q, u.ID)
	return err
}

// DecryptEncSecret decrypts the user specific encryption secret in the DB with the global encryption key
func (m UserModel) DecryptEncSecret(u *User) ([]byte, error) {
	ek, err := crypto.DecryptAuth(u.EncryptionKey, []byte(m.Config.Data.EncryptionKey), []byte(u.UserID))
	if err != nil {
		return []byte{}, fmt.Errorf("failed to decrypt user encryption secret: %w", err)
	}
	return ek, nil
}

// GetSoTRATCookie returns the Sea of Thieves RAT cookie from the database if present
func (m UserModel) GetSoTRATCookie(u *User) (string, error) {
	c, err := m.GetPrefStringEnc(u, UserPrefSoTAuthToken)
	if err != nil {
		return "", err
	}
	e, err := m.GetPrefInt64(u, UserPrefSoTAuthTokenExpiration)
	if err != nil {
		return "", err
	}
	if e > time.Now().Unix() {
		return "", fmt.Errorf("authentication token expired")
	}
	return c, nil
}
