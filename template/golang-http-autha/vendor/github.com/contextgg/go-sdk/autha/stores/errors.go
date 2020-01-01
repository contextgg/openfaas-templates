package stores

import (
	"errors"
)

var (
	// ErrNeedKeys when no secret key is passed through
	ErrNeedKeys = errors.New("Please provide secret keys for session encryption")
)
