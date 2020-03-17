package smashgg

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/contextgg/go-sdk/autha"
)

// Token struct
type Token struct {
	User *User `json:"-"`
}

var (
	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func getAvatar(name string, images []*Image) string {
	for _, image := range images {
		if image.Type == name {
			return image.URL
		}
	}
	return ""
}

type provider struct {
	service *Service
	useBio  bool
}

func (p *provider) Name() string {
	return "smashgg"
}

func (p *provider) BeginAuth(ctx context.Context, session autha.Session, params autha.Params) (string, error) {
	// if we already have a state don't regenerate.
	// sessions only last 12 hours by default
	code, _ := session.Get("state")
	if len(code) == 0 {
		// state for the oauth grant!
		code := randSeq(6)

		// set the state
		session.Set("state", code)
	}

	// returning no url won't redirect the page
	return "", nil
}

func (p *provider) Authorize(ctx context.Context, session autha.Session, params autha.Params) (autha.Token, error) {
	state, err := session.Get("state")
	if err != nil {
		return nil, fmt.Errorf("Could not load state from session %w", err)
	}
	if len(state) == 0 {
		return nil, errors.New("No state value in params")
	}

	userURL := params.Get("user-url")
	if len(userURL) == 0 {
		return nil, autha.NewWrapped("Please provider a Smashgg profile URL", autha.ErrTryAgain)
	}

	// extract the slug.
	slug, err := extractSlug(userURL)
	if err != nil {
		return nil, autha.NewWrapped(fmt.Sprintf("Invalid Smashgg URL %s", userURL), autha.ErrTryAgain)
	}

	user, err := p.service.GetUserBySlug(ctx, slug)
	if err != nil {
		return nil, autha.NewWrapped(fmt.Sprintf("Whoops, please try again later - %v", err), autha.ErrTryAgain)
	}

	if p.useBio && user.Bio != state {
		return nil, autha.NewWrapped(fmt.Sprintf("Invalid Bio for url %s; got %s, want %s", userURL, user.Bio, state), autha.ErrTryAgain)
	}

	if !p.useBio && user.Player.Prefix != state {
		return nil, autha.NewWrapped(fmt.Sprintf("Invalid Prefix for url %s; got %s, want %s", userURL, user.Player.Prefix, state), autha.ErrTryAgain)
	}

	// convert to a discord token!
	return &Token{user}, nil
}

func (p *provider) LoadProfile(ctx context.Context, token autha.Token, session autha.Session) (*autha.Profile, error) {
	t, ok := token.(*Token)
	if !ok {
		return nil, errors.New("Invalid smashgg token")
	}

	id := &autha.Profile{
		ID:          strconv.Itoa(int(t.User.ID)),
		Username:    t.User.Slug,
		DisplayName: t.User.Player.GamerTag,
		AvatarURL:   getAvatar("profile", t.User.Images),
		Raw:         t.User,
	}
	return id, nil
}

// NewProvider creates a new Provider
func NewProvider(accessToken []string, useBio bool) autha.AuthProvider {
	return &provider{
		service: NewService(accessToken...),
		useBio:  useBio,
	}
}
