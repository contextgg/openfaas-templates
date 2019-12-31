package autha

// Profile represents the profile from an external system
type Profile struct {
	ID          string      `json:"id"`
	Username    string      `json:"username,omitempty"`
	Email       string      `json:"email,omitempty"`
	DisplayName string      `json:"display_name,omitempty"`
	AvatarURL   string      `json:"avatar_url,omitempty"`
	Raw         interface{} `json:"raw,omitempty"`
}
