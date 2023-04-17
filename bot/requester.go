package bot

import (
	"errors"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/wneessen/arrgo/model"
)

// Requester wraps the discordgo.Member object to extend its functionality
type Requester struct {
	*discordgo.Member
	*model.UserModel
	*model.User
}

// List of Requester specific errors
var (
	ErrUserNotRegistered = errors.New("your user is not registered with the bot. Please use the " +
		"**/register** command to activate the full feature set first")
	ErrUserHasNoRATCookie = errors.New("you have not provided a Sea of Thieves authentication token. " +
		"Please store your cookie with the **/setrat** command first")
	ErrRATCookieExpired = errors.New("your Sea of Thieves authentication token is expired. " +
		"Please use the **/setrat** command to update your token")
	ErrMemberNil = errors.New("provided Member pointer must not be nil")
	ErrUserNil   = errors.New("provided User pointer must not be nil")
)

// NewRequesterFromMember returns a new *Requester pointer from a given *discordgo.Member
func NewRequesterFromMember(m *discordgo.Member, um *model.UserModel) (*Requester, error) {
	r := &Requester{UserModel: um, Member: m}
	if m == nil {
		return r, ErrMemberNil
	}
	u, err := r.UserModel.GetByUserID(m.User.ID)
	if err != nil {
		return r, ErrUserNotRegistered
	}
	r.User = u
	return r, nil
}

// NewRequesterFromUser returns a new *Requester pointer from a given *model.User
func NewRequesterFromUser(u *model.User, um *model.UserModel) (*Requester, error) {
	r := &Requester{UserModel: um, User: u}
	if u == nil {
		return r, ErrUserNil
	}
	return r, nil
}

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
	if r.User == nil {
		return "", ErrUserNil
	}
	c, err := r.UserModel.GetPrefStringEnc(r.User, model.UserPrefSoTAuthToken)
	if err != nil {
		return "", ErrUserHasNoRATCookie
	}
	e, err := r.UserModel.GetPrefInt64Enc(r.User, model.UserPrefSoTAuthTokenExpiration)
	if err != nil {
		return "", ErrUserHasNoRATCookie
	}
	if e < time.Now().Unix() {
		return "", ErrRATCookieExpired
	}
	return c, nil
}
