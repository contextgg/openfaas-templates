package function

import (
	"net/http"

	"github.com/contextgg/go-sdk/autha"
)

// NewProvider needs to be implemented or else!
func NewProvider(callbackURL string) autha.AuthProvider {
	return nil
}

// NewHandler for routing
func NewHandler(auth *autha.Config) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) { 
		auth.Begin(w, r) 
	})
	mux.HandleFunc("/callback", func (w http.ResponseWriter, r *http.Request) { 
		auth.Callback(w, r) 
	})
	return mux
}