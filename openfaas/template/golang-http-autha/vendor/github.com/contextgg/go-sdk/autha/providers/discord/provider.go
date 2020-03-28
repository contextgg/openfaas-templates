package discord

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/contextgg/go-sdk/gen"
	"github.com/contextgg/go-sdk/httpbuilder"

	"github.com/mitchellh/mapstructure"
	"golang.org/x/oauth2"

	"github.com/contextgg/go-sdk/autha"
)

const (
	authURL        = "https://discordapp.com/api/oauth2/authorize"
	tokenURL       = "https://discordapp.com/api/oauth2/token"
	userEndpoint   = "https://discordapp.com/api/users/@me"
	guildsEndpoint = "https://discordapp.com/api/guilds"
)

const (
	// ScopeIdentify allows /users/@me without email
	ScopeIdentify string = "identify"
	// ScopeEmail enables /users/@me to return an email
	ScopeEmail string = "email"
	// ScopeConnections allows /users/@me/connections to return linked Twitch and YouTube accounts
	ScopeConnections string = "connections"
	// ScopeGuilds allows /users/@me/guilds to return basic information about all of a user's guilds
	ScopeGuilds string = "guilds"
	// ScopeJoinGuild allows /invites/{invite.id} to be used for joining a user's guild
	ScopeJoinGuild string = "guilds.join"
	// ScopeGroupDMjoin allows your app to join users to a group dm
	ScopeGroupDMjoin string = "gdm.join"
	// ScopeBot for oauth2 bots, this puts the bot in the user's selected guild by default
	ScopeBot string = "bot"
	// ScopeWebhook this generates a webhook that is returned in the oauth token response for authorization code grants
	ScopeWebhook string = "webhook.incoming"
)

// CurrentUser the object representing the current discord user
type CurrentUser struct {
	ID            string  `json:"id"`
	Username      string  `json:"username"`
	Discriminator string  `json:"discriminator"`
	Avatar        *string `json:"avatar"`
	Bot           bool    `json:"bot"`
	MFAEnabled    bool    `json:"mfa_enabled"`
	Locale        string  `json:"locale"`
	Verified      bool    `json:"verified"`
	Email         string  `json:"email"`
	Flags         int     `json:"flags"`
	PremiumType   int     `json:"premium_type"`
}

// Webhook struct
type Webhook struct {
	ID        string            `json:"id"`
	Token     string            `json:"token"`
	Name      string            `json:"name,omitempty"`
	ChannelID string            `json:"channel_id" mapstructure:"channel_id"`
	GuildID   string            `json:"guild_id" mapstructure:"guild_id"`
	Avatar    string            `json:"avatar,omitempty"`
	Type      *int              `json:"type,omitempty"`
	URL       string            `json:"url,omitempty"`
	Extra     map[string]string `json:"extra,omitempty"`
}

type Guild struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Icon    string `json:"icon"`
	OwnerID string `json:"owner_id" mapstructure:"owner_id"`
}

// Token struct
type Token struct {
	// AccessToken is the token that authorizes and authenticates
	// the requests.
	AccessToken string `json:"access_token"`

	// TokenType is the type of token.
	// The Type method returns either this or "Bearer", the default.
	TokenType string `json:"token_type,omitempty"`

	// RefreshToken is a token that's used by the application
	// (as opposed to the user) to refresh the access token
	// if it expires.
	RefreshToken string `json:"refresh_token,omitempty"`

	// Expiry is the optional expiration time of the access token.
	//
	// If zero, TokenSource implementations will reuse the same
	// token forever and RefreshToken or equivalent
	// mechanisms for that TokenSource will not be used.
	Expiry time.Time `json:"expiry,omitempty"`

	// Webhook extra information
	Webhook *Webhook `json:"webhook,omitempty"`

	// Guild extra information
	Guild *Guild `json:"guild,omitempty"`

	// GuildID only used when bot
	GuildID string `json:"guild_id,omitempty"`

	// Permissions only used when bot
	Permissions string `json:"permissions,omitempty"`

	// Error when things have errors
	Error error `json:"-"`
}

func convertToken(tk *oauth2.Token, params autha.Params) *Token {
	if tk == nil {
		return nil
	}

	t := &Token{
		AccessToken:  tk.AccessToken,
		TokenType:    tk.TokenType,
		RefreshToken: tk.RefreshToken,
		Expiry:       tk.Expiry,
		GuildID:      params.Get("guild_id"),
		Permissions:  params.Get("permissions"),
	}

	wh := tk.Extra("webhook")
	if wh != nil {
		var result Webhook
		if err := mapstructure.Decode(wh, &result); err == nil {
			t.Webhook = &result
		} else {
			t.Error = err
		}
	}

	g := tk.Extra("guild")
	if g != nil {
		var result Guild
		if err := mapstructure.Decode(g, &result); err == nil {
			t.Guild = &result
		} else {
			t.Error = err
		}
	}

	return t
}

type provider struct {
	extras map[string]string
	config *oauth2.Config
}

func (p *provider) Name() string {
	return "discord"
}

func (p *provider) BeginAuth(ctx context.Context, session autha.Session, params autha.Params) (string, error) {
	// state for the oauth grant!
	state := gen.RandomString(64)

	// set the state
	session.Set("state", state)

	opts := []oauth2.AuthCodeOption{
		oauth2.AccessTypeOnline,
	}

	if p.extras != nil {
		keys := []string{"permissions"}
		for _, key := range keys {
			if val, ok := p.extras[key]; ok {
				opts = append(opts, oauth2.SetAuthURLParam(key, val))
			}
		}
	}

	// generate the url
	return p.config.AuthCodeURL(state, opts...), nil
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

	// convert to a discord token!
	return convertToken(token, params), nil
}

func (p *provider) LoadProfile(ctx context.Context, token autha.Token, session autha.Session) (*autha.Profile, error) {
	t, ok := token.(*Token)
	if !ok {
		return nil, errors.New("Wrong token type")
	}

	profile, err := userProfile(ctx, t)
	if err != nil {
		return nil, err
	}

	if t.Webhook != nil {
		return webhookProfile(ctx, t, profile)
	}

	if t.Guild != nil {
		return botProfile(ctx, t, profile)
	}

	return profile, nil
}

func webhookProfile(ctx context.Context, t *Token, profile *autha.Profile) (*autha.Profile, error) {
	avatarURL := ""
	if len(t.Webhook.Avatar) > 0 {
		avatarURL = fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.jpg?size=512", t.Webhook.ID, t.Webhook.Avatar)
	}

	wp := &autha.Profile{
		ID:          fmt.Sprintf("%s-%s-%s", profile.ID, t.Webhook.GuildID, t.Webhook.ChannelID),
		Type:        "Webhook",
		Username:    t.Webhook.ID,
		DisplayName: t.Webhook.Name,
		AvatarURL:   avatarURL,
		Raw:         t.Webhook,
	}
	return wp, nil
}
func botProfile(ctx context.Context, t *Token, profile *autha.Profile) (*autha.Profile, error) {
	avatarURL := ""
	if len(t.Guild.Icon) > 0 {
		avatarURL = fmt.Sprintf("https://cdn.discordapp.com/icons/%s/%s.jpg?size=512", t.Guild.ID, t.Guild.Icon)
	}

	bp := &autha.Profile{
		ID:          fmt.Sprintf("%s-%s", profile.ID, t.Guild.ID),
		Type:        "Bot",
		Username:    t.Guild.ID,
		DisplayName: t.Guild.Name,
		AvatarURL:   avatarURL,
		Raw:         t.Guild,
	}
	return bp, nil
}
func userProfile(ctx context.Context, t *Token) (*autha.Profile, error) {
	authType := t.TokenType
	accessToken := t.AccessToken

	// todo get the user!
	var user CurrentUser
	status, err := httpbuilder.New().
		SetURL(userEndpoint).
		SetAuthToken(authType, accessToken).
		SetOut(&user).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("Invalid Status Code %d", status)
	}

	avatarURL := ""
	if user.Avatar != nil {
		avatarURL = fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.jpg?size=512", user.ID, *user.Avatar)
	}

	profile := &autha.Profile{
		ID:          user.ID,
		Username:    fmt.Sprintf("%s#%s", user.Username, user.Discriminator),
		Email:       user.Email,
		DisplayName: user.Username,
		AvatarURL:   avatarURL,
		Raw:         user,
	}
	return profile, nil
}

// NewProvider creates a new Provider
func NewProvider(clientID, clientSecret, callbackURL string, scopes []string, extras map[string]string) autha.AuthProvider {
	return &provider{
		extras: extras,
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
		c.Scopes = scopes
	} else {
		c.Scopes = []string{ScopeIdentify}
	}

	return c
}
