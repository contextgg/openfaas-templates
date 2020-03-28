package twitter

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/mrjones/oauth"

	"github.com/contextgg/go-sdk/autha"
	"github.com/contextgg/go-sdk/httpbuilder"
)

var (
	requestURL      = "https://api.twitter.com/oauth/request_token"
	authorizeURL    = "https://api.twitter.com/oauth/authorize"
	authenticateURL = "https://api.twitter.com/oauth/authenticate"
	tokenURL        = "https://api.twitter.com/oauth/access_token"
	endpointProfile = "https://api.twitter.com/1.1/account/verify_credentials.json"
)

// CurrentUser the object representing the current discord user
type CurrentUser struct {
	ID              string `json:"id_str"`
	Name            string `json:"name"`
	ScreenName      string `json:"screen_name"`
	Location        string `json:"location"`
	Description     string `json:"description"`
	Email           string `json:"email"`
	ProfileImageURL string `json:"profile_image_url"`
	Verified        bool   `json:"verified"`
	Protected       bool   `json:"protected"`
	Lang            string `json:"lang"`
}

type provider struct {
	callbackURL string
	consumer    *oauth.Consumer
}

func (p *provider) Name() string {
	return "twitter"
}

func (p *provider) BeginAuth(ctx context.Context, session autha.Session, params autha.Params) (string, error) {
	reqToken, url, err := p.consumer.GetRequestTokenAndUrl(p.callbackURL)
	if err != nil {
		return "", err
	}

	session.Set("request_token", reqToken.Token)
	session.Set("request_secret", reqToken.Secret)

	return url, nil
}

func (p *provider) Authorize(ctx context.Context, session autha.Session, params autha.Params) (autha.Token, error) {
	code := params.Get("oauth_verifier")

	reqToken, err := session.Get("request_token")
	if err != nil {
		return nil, err
	}
	reqSecret, err := session.Get("request_secret")
	if err != nil {
		return nil, err
	}

	req := &oauth.RequestToken{
		Token:  reqToken,
		Secret: reqSecret,
	}
	log.Printf("ReqToken: %s; ReqSecret: %s", reqToken, reqSecret)

	accessToken, err := p.consumer.AuthorizeToken(req, code)
	if err != nil {
		return nil, err
	}
	log.Printf("AccessToken: %s; AccessSecret: %s", accessToken.Token, accessToken.Secret)

	return accessToken, err
}

func (p *provider) LoadProfile(ctx context.Context, token autha.Token, session autha.Session) (*autha.Profile, error) {
	accessToken, ok := token.(*oauth.AccessToken)
	if !ok {
		return nil, errors.New("Invalid access token")
	}

	client, err := p.consumer.MakeHttpClient(accessToken)
	if err != nil {
		return nil, err
	}

	var user CurrentUser
	status, err := httpbuilder.New().
		SetClient(client).
		SetURL(endpointProfile).
		AddQuery("include_entities", "false").
		AddQuery("skip_status", "true").
		AddQuery("include_email", "true").
		SetOut(&user).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("Invalid Status Code %d", status)
	}

	id := &autha.Profile{
		ID:          user.ID,
		Username:    user.ScreenName,
		Email:       user.Email,
		DisplayName: user.Name,
		AvatarURL:   user.ProfileImageURL,
		Raw:         user,
	}
	return id, nil
}

// NewProvider creates a new Provider
func NewProvider(clientID, clientSecret, callbackURL string) autha.AuthProvider {
	return &provider{
		callbackURL: callbackURL,
		consumer:    newConsumer(clientID, clientSecret, false),
	}
}

func newConsumer(clientID, clientSecret string, debug bool) *oauth.Consumer {
	c := oauth.NewConsumer(
		clientID,
		clientSecret,
		oauth.ServiceProvider{
			RequestTokenUrl:   requestURL,
			AuthorizeTokenUrl: authorizeURL,
			AccessTokenUrl:    tokenURL,
		})

	c.Debug(debug)
	return c
}
