package autha

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

var (
	// StatusLogin when user has already been created
	StatusLogin = "login"
)

// Config for our common authentication pattern
type Config struct {
	connection   string
	loginURL     string
	errorURL     string
	sessionStore SessionStore
	userStore    UserStore
	authProvider AuthProvider
	userService  UserService

	debug bool
}

// NewConfig will return a new config
func NewConfig(
	connection string,
	loginURL string,
	errorURL string,
	sessionStore SessionStore,
	userStore UserStore,
	authProvider AuthProvider,
	userService UserService,
) *Config {
	return &Config{
		connection:   connection,
		loginURL:     loginURL,
		errorURL:     errorURL,
		sessionStore: sessionStore,
		userStore:    userStore,
		authProvider: authProvider,
		userService:  userService,
	}
}

// SetDebug so we can log more info
func (c *Config) SetDebug() {
	c.debug = true
}

func (c *Config) fullErrorURL(errorType string) string {
	str := c.errorURL
	if len(str) == 0 {
		str = c.loginURL
	}
	u, _ := url.Parse(str)

	q := u.Query()
	q.Set("error.type", errorType)

	u.RawQuery = q.Encode()

	return u.String()
}

// Begin the auth method
func (c *Config) Begin(w http.ResponseWriter, r *http.Request) Session {
	ctx := r.Context()

	// get the current session!
	session, err := c.sessionStore.Load(c.connection, r)
	if err != nil {
		http.Redirect(w, r, c.fullErrorURL("session"), http.StatusFound)
		log.Print(fmt.Errorf("Error Session Load: %w", err))
		return nil
	}

	// if there's an ID check with our current user!
	cid, ok, err := c.userStore.Load(r)
	if err != nil {
		http.Redirect(w, r, c.fullErrorURL("identity"), http.StatusFound)
		log.Print(fmt.Errorf("Error loading current user id: %w", err))
		return nil
	}

	paramID := r.URL.Query().Get("id")
	if len(paramID) > 0 {
		if !ok || cid != paramID {
			// wrong user id
			http.Redirect(w, r, c.fullErrorURL("user"), http.StatusFound)
			log.Print(fmt.Errorf("User IDs are wrong: want %s, got %s", paramID, cid))
			return nil
		}
	}

	// set the from in session. If no from supplied it'll reset the value
	paramFrom := r.URL.Query().Get("from")
	session.Set("from", paramFrom)

	url, err := c.authProvider.BeginAuth(ctx, session, r.URL.Query())
	if err != nil {
		http.Redirect(w, r, c.fullErrorURL("auth"), http.StatusFound)
		log.Print(fmt.Errorf("Error Begin Auth: %w", err))
		return nil
	}

	// save the session
	if err := c.sessionStore.Save(session, w, r); err != nil {
		http.Redirect(w, r, c.fullErrorURL("session"), http.StatusFound)
		log.Print(fmt.Errorf("Error Session Save: %w", err))
		return nil
	}

	if len(url) > 0 {
		http.Redirect(w, r, url, http.StatusFound)
	}

	return session
}

// Callback for the provider
func (c *Config) Callback(w http.ResponseWriter, r *http.Request) Session {
	ctx := r.Context()

	session, err := c.sessionStore.Load(c.connection, r)
	if err != nil {
		http.Redirect(w, r, c.fullErrorURL("session"), http.StatusFound)
		log.Print(fmt.Errorf("Error Session Load: %w", err))
		return nil
	}

	r.ParseForm()
	token, err := c.authProvider.Authorize(ctx, session, r.Form)
	if err != nil && errors.Is(err, ErrTryAgain) {
		session.Set("message", err.Error())
		return session
	}
	if err != nil {
		http.Redirect(w, r, c.fullErrorURL("id"), http.StatusFound)
		log.Print(fmt.Errorf("Error Authorize: %w", err))
		return nil
	}

	profile, err := c.authProvider.LoadProfile(ctx, token, session)
	if err != nil {
		http.Redirect(w, r, c.fullErrorURL("identity"), http.StatusFound)
		log.Print(fmt.Errorf("Error Load Identity: %w", err))
		return nil
	}

	// calcuate the aggregate id!
	userID := c.userService.CalculateAggregateID(c.connection, profile.ID)
	isConnecting := false
	var primaryUserID *string

	if c.debug {
		log.Printf("User ID: %s", userID)
	}

	cid, ok, err := c.userStore.Load(r)
	if err != nil {
		http.Redirect(w, r, c.fullErrorURL("identity"), http.StatusFound)
		log.Print(fmt.Errorf("Error loading current user id: %w", err))
		return nil
	}

	if c.debug {
		log.Printf("Current ID: %s", cid)
	}

	if ok && cid != userID {
		isConnecting = true
		primaryUserID = &cid
	}

	if c.debug {
		log.Printf("Is Connecting: %#v", isConnecting)
		log.Printf("Primary User ID: %#v", primaryUserID)
	}

	// if we are linking we need to tell the user this is a secondary account!
	pu := NewPersistUser(
		c.authProvider.Name(),
		c.connection,
		userID,
		token,
		profile,
		primaryUserID,
		isConnecting,
	)

	if c.debug {
		log.Printf("NewPersistUser: %#v", pu)
	}

	// if we have an id store it!
	if err := c.userService.Persist(r.Context(), userID, pu); err != nil {
		http.Redirect(w, r, c.fullErrorURL("user"), http.StatusFound)
		log.Print(fmt.Errorf("Error Login: %w", err))
		return nil
	}

	id := userID
	if isConnecting && primaryUserID != nil {
		id = *primaryUserID

		cu := NewConnectUser(
			c.authProvider.Name(),
			c.connection,
			userID,
			token,
			profile,
		)

		if c.debug {
			log.Printf("NewConnectUser: %#v", cu)
		}

		// we need to connect the accounts.
		if err := c.userService.Connect(r.Context(), id, cu); err != nil {
			http.Redirect(w, r, c.fullErrorURL("user"), http.StatusFound)
			log.Print(fmt.Errorf("Error connecting profiles: %w", err))
			return nil
		}
	}

	if c.debug {
		log.Printf("User ID to save: %#v", id)
	}

	// this is to auth the user!
	if err := c.userStore.Save(id, w, r); err != nil {
		http.Redirect(w, r, c.fullErrorURL("id"), http.StatusFound)
		log.Print(fmt.Errorf("Error Profile Store Save: %w", err))
		return nil
	}

	// save the session
	if err := c.sessionStore.Save(session, w, r); err != nil {
		http.Redirect(w, r, c.fullErrorURL("session"), http.StatusFound)
		log.Print(fmt.Errorf("Error Session Save: %w", err))
		return nil
	}

	// see if we have from in the session
	from, _ := session.Get("from")

	if len(from) > 0 {
		http.Redirect(w, r, from, http.StatusFound)
		return session
	}

	// what's the next step?
	if len(c.loginURL) > 0 {
		http.Redirect(w, r, c.loginURL, http.StatusFound)
		return session
	}

	// return session!
	return session
}
