package smashgg

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
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
}

func (p *provider) Name() string {
	return "smashgg"
}

func (p *provider) BeginAuth(ctx context.Context, session autha.Session, params autha.Params) (string, error) {
	userID := params.Get("userID")
	if userID == "" {
		return "", errors.New("We require a SmashGG user id")
	}

	// state for the oauth grant!
	code := randSeq(6)

	// set the state
	session.Set("state", fmt.Sprintf("%s/%s", userID, code))

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

	split := strings.Split(state, "/")
	if len(split) != 2 {
		return nil, fmt.Errorf("State has invalid value of %s", state)
	}

	userID, err := strconv.Atoi(split[0])
	if err != nil {
		return nil, fmt.Errorf("Could not parse userID %w", err)
	}

	user, err := p.service.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("Could not find user %w", err)
	}

	prefix := user.Player.Prefix
	if prefix != split[1] {
		return nil, fmt.Errorf("Invalid prefix for user %d; got %s, want %s", userID, prefix, split[1])
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
func NewProvider(accessToken []string) autha.AuthProvider {
	return &provider{
		service: NewService(accessToken...),
	}
}
