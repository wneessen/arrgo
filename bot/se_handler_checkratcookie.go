package bot

import (
	"errors"
	"fmt"
	"github.com/wneessen/arrgo/model"
	"net/http"
	"time"
)

// ScheduledEventCheckRATCookies performs scheuled checks if the provided RAT cookies are still valid
func (b *Bot) ScheduledEventCheckRATCookies() error {
	ll := b.Log.With().Str("context", "bot.ScheduledEventCheckRATCookies").Logger()
	ul, err := b.Model.User.GetUsers()
	if err != nil {
		return fmt.Errorf("failed to retrieve user list from DB: %s", err)
	}

	for _, u := range ul {
		ie := false
		te, err := b.Model.User.GetPrefInt64Enc(u, model.UserPrefSoTAuthTokenExpiration)
		if err != nil {
			if !errors.Is(err, model.ErrUserPrefNotExistent) {
				ll.Error().Msgf("failed to retrieve RAT cookie expiration from DB: %s", err)
				continue
			}
			ll.Debug().Msgf("User has no RAT token configured... skipping")
			continue
		}

		// Token is expired (timestamp wise)
		if time.Now().Unix() > te {
			ie = true
		}

		// In some cases the token might be expired on the server end... let's test with a HTTP request
		if !ie {
			rq := &Requester{nil, b.Model.User, u}
			hc, err := NewHTTPClient()
			if err != nil {
				ll.Error().Msgf(ErrFailedHTTPClient, err)
				continue
			}
			c, err := rq.GetSoTRATCookie()
			if err != nil {
				ll.Error().Err(err)
				continue
			}
			r, err := hc.HttpReq(ApiURLSoTUserOverview, ReqMethodGet, nil)
			if err != nil {
				ll.Error().Err(err)
				continue
			}
			r.SetSOTRequest(c)
			_, ho, err := hc.Fetch(r)
			if err != nil {
				ll.Error().Err(err)
				continue
			}
			if ho.StatusCode == http.StatusUnauthorized {
				ie = true
			}
		}

		if ie {
			na, err := b.Model.User.GetPrefBool(u, model.UserPrefSoTAuthTokenNotified)
			if err != nil && !errors.Is(err, model.ErrUserPrefNotExistent) {
				ll.Error().Msgf("failed to retrieve RAT cookie already notified status from DB: %s", err)
				continue
			}
			if !na {
				st, err := b.Session.UserChannelCreate(u.UserID)
				if err != nil {
					ll.Error().Msgf("failed to create DM channel with user: %s", err)
					continue
				}
				_, err = b.Session.ChannelMessageSend(st.ID, "Your SoT RAT cookie has expired. Please use "+
					"the `/setrat` command to set a new one.")
				if err != nil {
					ll.Error().Msgf("failed to send DM: %s", err)
					continue
				}
				if err := b.Model.User.SetPref(u, model.UserPrefSoTAuthTokenNotified, true); err != nil {
					ll.Error().Msgf("failed to set 'user notified' user pref in DB: %s", err)
				}
			}
		}
	}
	return nil
}
