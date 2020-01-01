package discord

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/contextgg/go-sdk/gen"
	"github.com/contextgg/go-sdk/httpbuilder"

	"golang.org/x/oauth2"

	"github.com/contextgg/go-sdk/autha"
)

const (
	authURL      string = "https://api.twitch.tv/kraken/oauth2/authorize"
	tokenURL     string = "https://api.twitch.tv/kraken/oauth2/token"
	userEndpoint string = "https://api.twitch.tv/kraken/user"
)

const (
	// ScopeUserRead provides read access to non-public user information, such
	// as their email address.
	ScopeUserRead string = "user_read"
	// ScopeUserBlocksEdit provides the ability to ignore or unignore on
	// behalf of a user.
	ScopeUserBlocksEdit string = "user_blocks_edit"
	// ScopeUserBlocksRead provides read access to a user's list of ignored
	// users.
	ScopeUserBlocksRead string = "user_blocks_read"
	// ScopeUserFollowsEdit provides access to manage a user's followed
	// channels.
	ScopeUserFollowsEdit string = "user_follows_edit"
	// ScopeChannelRead provides read access to non-public channel information,
	// including email address and stream key.
	ScopeChannelRead string = "channel_read"
	// ScopeChannelEditor provides write access to channel metadata (game,
	// status, etc).
	ScopeChannelEditor string = "channel_editor"
	// ScopeChannelCommercial provides access to trigger commercials on
	// channel.
	ScopeChannelCommercial string = "channel_commercial"
	// ScopeChannelStream provides the ability to reset a channel's stream key.
	ScopeChannelStream string = "channel_stream"
	// ScopeChannelSubscriptions provides read access to all subscribers to
	// your channel.
	ScopeChannelSubscriptions string = "channel_subscriptions"
	// ScopeUserSubscriptions provides read access to subscriptions of a user.
	ScopeUserSubscriptions string = "user_subscriptions"
	// ScopeChannelCheckSubscription provides read access to check if a user is
	// subscribed to your channel.
	ScopeChannelCheckSubscription string = "channel_check_subscription"
	// ScopeChatLogin provides the ability to log into chat and send messages.
	ScopeChatLogin string = "chat_login"
)

// CurrentUser the object representing the current discord user
type CurrentUser struct {
	ID               string `json:"_id"`
	Bio              string `json:"bio"`
	Name             string `json:"name"`
	DisplayName      string `json:"display_name"`
	Email            string `json:"email"`
	EmailVerified    bool   `json:"email_verified"`
	Logo             string `json:"logo"`
	Partnered        bool   `json:"partnered"`
	TwitterConnected bool   `json:"twitter_connected"`
}

type provider struct {
	config *oauth2.Config
}

func (p *provider) Name() string {
	return "twitch"
}

func (p *provider) BeginAuth(ctx context.Context, session autha.Session, params autha.Params) (string, error) {
	// state for the oauth grant!
	state := gen.RandomString(64)

	// set the state
	session.Set("state", state)

	// generate the url
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOnline), nil
}

func (p *provider) Authorize(ctx context.Context, session autha.Session, params autha.Params) (autha.Token, error) {
	state := params.Get("state")
	if len(state) == 0 {
		return nil, errors.New("No state value in params")
	}

	if !autha.SessionHasValue(session, "state", state) {
		return nil, errors.New("Invalid state")
	}

	code := params.Get("code")
	if len(code) == 0 {
		return nil, errors.New("No code value in params")
	}

	token, err := p.config.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, err
	}

	if !token.Valid() {
		return nil, errors.New("Invalid token received from provider")
	}

	return token, nil
}

func (p *provider) LoadProfile(ctx context.Context, token autha.Token, session autha.Session) (*autha.Profile, error) {
	t, ok := token.(*oauth2.Token)
	if !ok {
		return nil, errors.New("Wrong token type")
	}

	accessToken := t.AccessToken

	// todo get the user!
	var user CurrentUser
	status, err := httpbuilder.New().
		SetURL(userEndpoint).
		SetAuthToken("OAuth", accessToken).
		AddHeader("Client-ID", p.config.ClientID).
		AddHeader("Accept", "application/vnd.twitchtv.v5+json").
		SetOut(&user).
		SetLogger(log.Printf).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("Invalid Status Code %d", status)
	}

	id := &autha.Profile{
		ID:          user.ID,
		Username:    user.Name,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.Logo,
		Raw:         user,
	}
	return id, nil
}

// NewProvider creates a new Provider
func NewProvider(clientID, clientSecret, callbackURL string, scopes ...string) autha.AuthProvider {
	return &provider{
		config: newConfig(clientID, clientSecret, callbackURL, scopes),
	}
}

func newConfig(clientID, clientSecret, callbackURL string, scopes []string) *oauth2.Config {
	c := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  callbackURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
		Scopes: []string{},
	}

	if len(scopes) > 0 {
		for _, scope := range scopes {
			c.Scopes = append(c.Scopes, scope)
		}
	} else {
		c.Scopes = []string{ScopeUserRead}
	}

	return c
}
