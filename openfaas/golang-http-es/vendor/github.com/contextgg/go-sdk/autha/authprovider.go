package autha

import "context"

// AuthProvider is the common interface for doing auth
type AuthProvider interface {
	// Name of the provider IE Discord, Twitter, Twitch
	Name() string

	// BeginAuth the start of a token exchange
	BeginAuth(context.Context, Session, Params) (string, error)

	// Authorize confirm everything is ok
	Authorize(context.Context, Session, Params) (Token, error)

	// LoadProfile will try to load the current users profile
	LoadProfile(context.Context, Token, Session) (*Profile, error)
}
