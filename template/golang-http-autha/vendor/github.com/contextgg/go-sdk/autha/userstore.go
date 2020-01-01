package autha

import "net/http"

// UserStore to save and load user
type UserStore interface {
	// Save the user to the request
	Save(string, http.ResponseWriter, *http.Request) error
	Load(*http.Request) (string, bool, error)
}
