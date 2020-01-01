package smashgg

import (
	"encoding/json"
	"strconv"
)

// GraphQLRequest when making a request to graphql
type GraphQLRequest struct {
	OperationName string                 `json:"operationName"`
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
}

// GraphQLError an error that is sent back
type GraphQLError struct {
	Message string `json:"message"`
}

func (e GraphQLError) Error() string {
	return "graphql: " + e.Message
}

// GraphQLResponse the response from a graphql request
type GraphQLResponse struct {
	Data   interface{}     `json:"data"`
	Errors []*GraphQLError `json:"errors"`
}

// SmashggID helps convert time
type SmashggID int

// UnmarshalJSON needs to support strings and ints
func (sid *SmashggID) UnmarshalJSON(b []byte) error {
	if b[0] == '"' {
		var str string
		if err := json.Unmarshal(b, &str); err != nil {
			return err
		}

		out, err := strconv.Atoi(str)
		if err != nil {
			return err
		}

		*sid = SmashggID(out)
		return nil
	}

	var id int
	if err := json.Unmarshal(b, &id); err != nil {
		return err
	}
	*sid = SmashggID(id)
	return nil
}

// Image standard type for smashgg
type Image struct {
	ID     SmashggID `json:"id"`
	Width  int       `json:"width"`
	Height int       `json:"height"`
	Ratio  float32   `json:"ratio"`
	Type   string    `json:"type"`
	URL    string    `json:"url"`
}

// Player in smashgg
type Player struct {
	ID            SmashggID `json:"id"`
	Images        []*Image  `json:"images"`
	Prefix        string    `json:"prefix"`
	GamerTag      string    `json:"gamerTag"`
	Color         string    `json:"color"`
	TwitchStream  string    `json:"twitchStream"`
	TwitterHandle string    `json:"twitterHandle"`
	YouTube       string    `json:"youtube"`
	Region        string    `json:"region"`
	State         string    `json:"state"`
	Country       string    `json:"country"`
}
