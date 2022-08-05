package bot

import "github.com/bwmarrin/discordgo"

// Requester wraps the discordgo.Member object to extend its functionality
type Requester struct {
	*discordgo.Member
}

// IsAdmin returns true if the Requester has administrative permissions on the guild
func (r *Requester) IsAdmin() bool {
	return r.Member.Permissions&discordgo.PermissionAdministrator != 0
}

// CanModerateMembers returns true if the Requester has moderator permissions on the guild
func (r *Requester) CanModerateMembers() bool {
	return r.Member.Permissions&discordgo.PermissionModerateMembers != 0
}
