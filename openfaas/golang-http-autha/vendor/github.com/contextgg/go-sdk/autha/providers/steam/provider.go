package steam

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/contextgg/go-sdk/autha"
	"github.com/contextgg/go-sdk/httpbuilder"
)

type all struct {
	Response struct {
		Players []Player `json:"players"`
	} `json:"response"`
}

// Player is the info for a single steam player
type Player struct {
	SteamID             string `json:"steamid"`
	PersonaName         string `json:"personaname"`
	RealName            string `json:"realname"`
	AvatarFull          string `json:"avatarfull"`
	LocationCountryCode string `json:"loccountrycode"`
	LocationStateCode   string `json:"locstatecode"`
}

type steamToken struct {
	SteamID       string
	ResponseNonce string
}

const (
	// Steam API Endpoints
	apiLoginEndpoint       = "https://steamcommunity.com/openid/login"
	apiUserSummaryEndpoint = "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s"

	// OpenID settings
	openIDMode       = "checkid_setup"
	openIDNs         = "http://specs.openid.net/auth/2.0"
	openIDIdentifier = "http://specs.openid.net/auth/2.0/identifier_select"
)

type provider struct {
	apiKey      string
	callbackURL string
}

func (p *provider) Name() string {
	return "steam"
}

func (p *provider) BeginAuth(ctx context.Context, session autha.Session, params autha.Params) (string, error) {
	callbackURL, err := url.Parse(p.callbackURL)
	if err != nil {
		return "", err
	}

	urlValues := map[string]string{
		"openid.claimed_id": openIDIdentifier,
		"openid.identity":   openIDIdentifier,
		"openid.mode":       openIDMode,
		"openid.ns":         openIDNs,
		"openid.realm":      fmt.Sprintf("%s://%s", callbackURL.Scheme, callbackURL.Host),
		"openid.return_to":  callbackURL.String(),
	}

	u, err := url.Parse(apiLoginEndpoint)
	if err != nil {
		return "", err
	}

	v := u.Query()
	for key, value := range urlValues {
		v.Set(key, value)
	}
	u.RawQuery = v.Encode()

	return u.String(), nil
}

func (p *provider) Authorize(ctx context.Context, session autha.Session, params autha.Params) (autha.Token, error) {
	if params.Get("openid.mode") != "id_res" {
		return nil, errors.New("Mode must equal to \"id_res\"")
	}

	if params.Get("openid.return_to") != p.callbackURL {
		return nil, errors.New("The \"return_to url\" must match the url of current request")
	}

	v := make(url.Values)
	v.Set("openid.assoc_handle", params.Get("openid.assoc_handle"))
	v.Set("openid.signed", params.Get("openid.signed"))
	v.Set("openid.sig", params.Get("openid.sig"))
	v.Set("openid.ns", params.Get("openid.ns"))

	split := strings.Split(params.Get("openid.signed"), ",")
	for _, item := range split {
		v.Set("openid."+item, params.Get("openid."+item))
	}
	v.Set("openid.mode", "check_authentication")

	var content string
	status, err := httpbuilder.New().
		SetMethod(http.MethodPost).
		SetURL(apiLoginEndpoint).
		SetBody(v).
		SetOut(&content).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("Invalid Status Code %d", status)
	}

	response := strings.Split(content, "\n")
	if response[0] != "ns:"+openIDNs {
		return nil, errors.New("Wrong ns in the response")
	}

	if response[1] == "is_valid:false" {
		return nil, errors.New("Unable validate openId")
	}

	openIDURL := params.Get("openid.claimed_id")
	validationRegExp := regexp.MustCompile("^(http|https)://steamcommunity.com/openid/id/[0-9]{15,25}$")
	if !validationRegExp.MatchString(openIDURL) {
		return nil, errors.New("Invalid Steam ID pattern")
	}

	steamID := regexp.MustCompile("\\D+").ReplaceAllString(openIDURL, "")
	responseNonce := params.Get("openid.response_nonce")

	t := &steamToken{
		SteamID:       steamID,
		ResponseNonce: responseNonce,
	}

	return t, nil
}

func (p *provider) LoadProfile(ctx context.Context, token autha.Token, session autha.Session) (*autha.Profile, error) {
	stok, ok := token.(*steamToken)
	if !ok {
		return nil, errors.New("Invalid token")
	}

	apiURL := fmt.Sprintf(apiUserSummaryEndpoint, p.apiKey, stok.SteamID)

	var users all
	status, err := httpbuilder.New().
		SetURL(apiURL).
		SetOut(&users).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("Invalid Status Code %d", status)
	}

	// get the current users
	if l := len(users.Response.Players); l != 1 {
		return nil, fmt.Errorf("Expected one player in API response. Got %d", l)
	}

	user := users.Response.Players[0]

	id := &autha.Profile{
		ID:          user.SteamID,
		Username:    user.PersonaName,
		DisplayName: user.RealName,
		AvatarURL:   user.AvatarFull,
		Raw:         user,
	}
	return id, nil
}

// NewProvider creates a new Provider
func NewProvider(apiKey string, callbackURL string) autha.AuthProvider {
	return &provider{
		apiKey:      apiKey,
		callbackURL: callbackURL,
	}
}
