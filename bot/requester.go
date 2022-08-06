package bot

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/wneessen/arrgo/model"
	"time"
)

// Requester wraps the discordgo.Member object to extend its functionality
type Requester struct {
	*discordgo.Member
	*model.UserModel
}

// List of Requester specific errors
var (
	ErrUserNotRegistered = errors.New("your user is not registered with the bot. Please use the " +
		"**/register** command to activate the full feature set first")
	ErrUserHasNoRATCookie = errors.New("you have not provided a Sea of Thieves authentication token. " +
		"Please store your cookie with the **/setrat** command first")
	ErrRATCookieExpired = errors.New("your Sea of Thieves authentication token is expired. " +
		"Please use the **/setrat** command to update your token")
)

// IsAdmin returns true if the Requester has administrative permissions on the guild
func (r *Requester) IsAdmin() bool {
	return r.Member.Permissions&discordgo.PermissionAdministrator != 0
}

// CanModerateMembers returns true if the Requester has moderator permissions on the guild
func (r *Requester) CanModerateMembers() bool {
	return r.Member.Permissions&discordgo.PermissionModerateMembers != 0
}

// GetSoTRATCookie checks if the Requester has a SoT RAT cookie and reads it from the DB
func (r *Requester) GetSoTRATCookie() (string, error) {
	u, err := r.UserModel.GetByUserID(r.Member.User.ID)
	if err != nil {
		return "", ErrUserNotRegistered
	}
	c, err := r.UserModel.GetPrefStringEnc(u, model.UserPrefSoTAuthToken)
	if err != nil {
		return "", ErrUserHasNoRATCookie
	}
	e, err := r.UserModel.GetPrefInt64Enc(u, model.UserPrefSoTAuthTokenExpiration)
	if err != nil {
		return "", ErrUserHasNoRATCookie
	}
	if e < time.Now().Unix() {
		return "", ErrRATCookieExpired
	}
	return c, nil
}
