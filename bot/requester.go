package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/wneessen/arrgo/model"
	"time"
)

// Requester wraps the discordgo.Member object to extend its functionality
type Requester struct {
	*discordgo.Member
	*model.UserModel
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
	u, err := r.UserModel.GetByUserID(r.Member.User.ID)
	if err != nil {
		return "", fmt.Errorf("failed to look up user in DB: %w", err)
	}
	c, err := r.UserModel.GetPrefStringEnc(u, model.UserPrefSoTAuthToken)
	if err != nil {
		return "", fmt.Errorf("failed to fetch rat_token from DB: %w", err)
	}
	e, err := r.UserModel.GetPrefInt64Enc(u, model.UserPrefSoTAuthTokenExpiration)
	if err != nil {
		return "", fmt.Errorf("failed to fetch rat_token_expiration from DB: %w", err)
	}
	if e < time.Now().Unix() {
		return "", fmt.Errorf("authentication token expired")
	}
	return c, nil
}
