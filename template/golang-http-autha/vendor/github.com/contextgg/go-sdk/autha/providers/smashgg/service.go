package smashgg

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/contextgg/go-sdk/httpbuilder"
)

const queries = `
fragment ImageParts on Image {
	id
	width
	height
	ratio
	type
	url
}
fragment PlayerParts on Player {
	id
	prefix
	gamerTag
}
fragment AddressParts on Address {
  id
  city
  state
  stateId
  country
  countryId
}
fragment UserParts on User {
	id
  images {
    ...ImageParts
  }
  bio
  name
  slug
  player {
    ...PlayerParts
  }
  location {
    ...AddressParts
  }
  authorizations {
    id
    externalUsername
    type
    stream {
      id
      isOnline
      name
      type
    }
    url
  }
}
query PlayerQuery($id: ID!) {
	player(id: $id){
		...PlayerParts
	}
}
query UserQuery($slug: string) {
	user(slug: $slug){
		...UserParts
	}
}
`

// Service for handling API ideas
type Service struct {
	sync.Mutex

	keys  []string
	index int
}

// NewService create a new service
func NewService(keys ...string) *Service {
	fmt.Printf("Create service with %d keys\n", len(keys))

	return &Service{
		keys: keys,
	}
}

// GetPlayer by ID
func (s *Service) GetPlayer(ctx context.Context, id int) (*Player, error) {
	body := &GraphQLRequest{
		OperationName: "PlayerQuery",
		Query:         queries,
		Variables: map[string]interface{}{
			"id": id,
		},
	}

	w := struct {
		Player *Player `json:"player"`
	}{}
	if err := s.do(ctx, body, &w); err != nil {
		return nil, err
	}
	return w.Player, nil
}

// GetUserByURL get user by profile user
func (s *Service) GetUserByURL(ctx context.Context, url string) (*User, error) {
	// extract the slug.
	slug, err := extractSlug(url)
	if err != nil {
		return nil, err
	}

	body := &GraphQLRequest{
		OperationName: "UserQuery",
		Query:         queries,
		Variables: map[string]interface{}{
			"slug": slug,
		},
	}

	w := struct {
		User *User `json:"user"`
	}{}
	if err := s.do(ctx, body, &w); err != nil {
		return nil, err
	}
	return w.User, nil
}

func (s *Service) getKey() string {
	s.Lock()
	defer s.Unlock()

	l := len(s.keys)
	if l < 1 {
		return ""
	}

	key := s.keys[s.index]

	s.index = s.index + 1
	if s.index >= l {
		s.index = 0
	}

	return key
}

func (s *Service) do(ctx context.Context, req *GraphQLRequest, data interface{}) error {
	res := &GraphQLResponse{
		Data: data,
	}
	key := s.getKey()

	var errorString string
	status, err := httpbuilder.
		New().
		SetURL("https://api.smash.gg/gql/alpha").
		SetMethod(http.MethodPost).
		SetBearerToken(key).
		AddHeader("Content-Type", "application/json; charset=utf-8").
		AddHeader("Accept", "application/json; charset=utf-8").
		SetBody(&req).
		SetOut(&res).
		SetErrorString(&errorString).
		Do(ctx)
	if err != nil {
		return err
	}
	if len(res.Errors) > 0 {
		// return first error
		return res.Errors[0]
	}
	if status != http.StatusOK {
		return fmt.Errorf("FAILED: status code %d extra info %s", status, errorString)
	}

	return nil
}
