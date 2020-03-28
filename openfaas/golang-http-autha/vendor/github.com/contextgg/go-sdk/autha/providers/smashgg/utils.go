package smashgg

import (
	"fmt"
	"net/url"
	"strings"
)

func extractSlug(userURL string) (string, error) {
	// https://smash.gg/admin/user/46ee4f62/profile-settings

	u, err := url.Parse(userURL)
	if err != nil {
		return "", err
	}

	splits := strings.Split(u.Path, "/")
	if len(splits) < 4 {
		return "", fmt.Errorf("Invalid user url: %s", userURL)
	}

	return "user/" + splits[3], nil
}
