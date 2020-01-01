package main

import (
	"fmt"
	"errors"
	"os"
	"strings"
	"time"
	"strconv"
	
	"github.com/contextgg/go-sdk/secrets"
)

// ErrConfigMissing when some config has been missed
var ErrConfigMissing = errors.New("Config missing")

func parseIntOrDurationValue(val string, fallback time.Duration) time.Duration {
	if len(val) > 0 {
		parsedVal, parseErr := strconv.Atoi(val)
		if parseErr == nil && parsedVal >= 0 {
			return time.Duration(parsedVal) * time.Second
		}
	}

	duration, durationErr := time.ParseDuration(val)
	if durationErr != nil {
		return fallback
	}
	return duration
}
func parseInt(val string, fallback int) int {
	if len(val) > 0 {
		parsedVal, parseErr := strconv.Atoi(val)
		if parseErr == nil {
			return parsedVal
		}
	}
	return fallback
}
func parseBool(val string, fallback bool) bool {
	if len(val) == 0 {
		return fallback
	}
	return strings.EqualFold(val, "yes") || strings.EqualFold(val, "ok") || strings.EqualFold(val, "true")
}

// Config all the info for the app
type Config struct {
	ReadTimeout time.Duration
	WriteTimeout time.Duration

	Connection string
	DNS string
	SessionSecret string
	SessionSecure bool
	UserFunctionName string

	CallbackURL string
	LoginURL string
	ErrorURL string
}

func (c *Config) Load() error {
	c.ReadTimeout = parseIntOrDurationValue(os.Getenv("read_timeout"), 10*time.Second)
	c.WriteTimeout = parseIntOrDurationValue(os.Getenv("write_timeout"), 10*time.Second)

	c.Connection = secrets.MustReadSecret("connection", "")
	if len(c.Connection) == 0 {
		return fmt.Errorf("Missing connection - %w", ErrConfigMissing)
	}
	c.DNS = secrets.MustReadSecret("dns", "")
	if len(c.DNS) == 0 {
		return fmt.Errorf("Missing dns - %w", ErrConfigMissing)
	}
	c.SessionSecret = secrets.MustReadSecret("session_secret", "")
	if len(c.SessionSecret) == 0 {
		return fmt.Errorf("Missing session_secret - %w", ErrConfigMissing)
	}
	c.SessionSecure = parseBool(secrets.MustReadSecret("session_secure", "false"), false)

	c.UserFunctionName = os.Getenv("user_function_name")
	if len(c.UserFunctionName) == 0 {
		return fmt.Errorf("Missing user_function_name - %w", ErrConfigMissing)
	}

	c.CallbackURL = os.Getenv("callback_url")
	c.LoginURL = os.Getenv("login_url")
	c.ErrorURL = os.Getenv("error_url")
	return nil
}