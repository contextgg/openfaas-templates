package stores

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/sessions"

	"github.com/contextgg/go-sdk/autha"
)

const format = "_ctx_auth_%s"

var (
	// ErrKeyNotFound when you can't find the key in the session
	ErrKeyNotFound = errors.New("Key not found")
	// ErrValWrongType when the value is the wrong type
	ErrValWrongType = errors.New("Value wrong type")
	// ErrSessionWrongType when the value is the wrong type
	ErrSessionWrongType = errors.New("Session wrong type")
)

type sessionStore struct {
	cookieStore *sessions.CookieStore
}

func (s *sessionStore) Load(connection string, r *http.Request) (autha.Session, error) {
	name := fmt.Sprintf(format, connection)
	sess, err := s.cookieStore.Get(r, name)
	if err != nil {
		return nil, err
	}

	return &session{inner: sess}, nil
}

func (s *sessionStore) Save(sess autha.Session, w http.ResponseWriter, r *http.Request) error {
	wrap, ok := sess.(*session)
	if !ok {
		return ErrSessionWrongType
	}

	return wrap.inner.Save(r, w)
}

type session struct {
	inner *sessions.Session
}

func (s *session) Set(key, val string) error {
	s.inner.Values[key] = val
	return nil
}

func (s *session) Get(key string) (string, error) {
	val, ok := s.inner.Values[key]
	if !ok {
		return "", ErrKeyNotFound
	}

	real, ok := val.(string)
	if !ok {
		return "", ErrValWrongType
	}

	return real, nil
}

func (s *session) Clear() error {
	for k := range s.inner.Values {
		delete(s.inner.Values, k)
	}
	return nil
}

// NewSessionStore creates a new session store
func NewSessionStore(secure bool, keypairs ...[]byte) (autha.SessionStore, error) {
	if len(keypairs) == 0 {
		return nil, ErrNeedKeys
	}

	// create a new session store!
	cookieStore := sessions.NewCookieStore(keypairs...)
	cookieStore.Options.HttpOnly = true
	cookieStore.Options.Secure = secure

	// 12 hours, set this to something because if we don't then sessions
	// may never expire as long as the browser remains opened.
	cookieStore.MaxAge(int((time.Hour * 12) / time.Second))

	return &sessionStore{
		cookieStore: cookieStore,
	}, nil
}
