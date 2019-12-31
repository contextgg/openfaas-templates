package autha

import (
	"context"
)

// BaseUser struct
type BaseUser struct {
	Provider   string   `json:"provider"`
	Connection string   `json:"connection"`
	ID         string   `json:"id"`
	Token      Token    `json:"token"`
	Profile    *Profile `json:"profile"`
}

// PersistUser struct
type PersistUser struct {
	*BaseUser

	PrimaryUserID      *string `json:"primary_user_id,omitempty"`
	IsConnectedProfile bool    `json:"is_connected_profile"`
}

// ConnectUser struct
type ConnectUser struct {
	*BaseUser
}

// NewPersistUser return a new persist user struct
func NewPersistUser(provider, connection, id string, token Token, profile *Profile, primaryUserID *string, isConnectedProfile bool) *PersistUser {
	return &PersistUser{
		BaseUser: &BaseUser{
			Provider:   provider,
			Connection: connection,
			ID:         id,
			Token:      token,
			Profile:    profile,
		},
		PrimaryUserID:      primaryUserID,
		IsConnectedProfile: isConnectedProfile,
	}
}

// NewConnectUser return a new persist user struct
func NewConnectUser(provider, connection, id string, token Token, profile *Profile) *ConnectUser {
	return &ConnectUser{
		BaseUser: &BaseUser{
			Provider:   provider,
			Connection: connection,
			ID:         id,
			Token:      token,
			Profile:    profile,
		},
	}
}

// UserService is the common interface for users
type UserService interface {
	SetDebug()
	CalculateAggregateID(string, string) string
	Persist(context.Context, string, *PersistUser) error
	Connect(context.Context, string, *ConnectUser) error
}
