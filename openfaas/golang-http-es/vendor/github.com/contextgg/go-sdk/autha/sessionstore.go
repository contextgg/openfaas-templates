package autha

import "net/http"

// Params for a request
type Params interface {
	Get(string) string
}

// Token is the returned auth
type Token interface{}

// SessionStore to save and load sessions
type SessionStore interface {
	// Load a session for a given provider
	Load(string, *http.Request) (Session, error)

	// Save the session to the request
	Save(Session, http.ResponseWriter, *http.Request) error
}

// Session represents a providers session
type Session interface {
	// Set a value in the session
	Set(key, value string) error
	// Get the value from the session
	Get(key string) (string, error)
	// Clear will remove all variables from the current session
	Clear() error
}

// SessionHasValue loads value from session and compares the value
func SessionHasValue(session Session, key, value string) bool {
	val, err := session.Get(key)
	if err != nil {
		return false
	}
	return val == value
}
