package autha

import "net/http"

// UserStore to save and load user
type UserStore interface {
	// Save the user to the request
	Save(string, http.ResponseWriter, *http.Request) error
	// Remove the user from the store
	Remove(http.ResponseWriter, *http.Request) error
	// Load the current user
	Load(*http.Request) (string, bool, error)
}
