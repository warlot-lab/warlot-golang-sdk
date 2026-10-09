package devcli

import (
	"fmt"
	"strings"
)

// UsageError represents an operator error caused by missing or invalid CLI flags.
type UsageError struct {
	Err  error
	Hint string
}

func (e *UsageError) Error() string {
	return e.Err.Error()
}

// FlagErrorf constructs a UsageError with a diagnostic fix hint.
func FlagErrorf(hint, format string, a ...any) error {
	return &UsageError{
		Err:  fmt.Errorf(format, a...),
		Hint: hint,
	}
}

// RequireFlag validates that a flag string is non-empty, returning a UsageError if missing.
func RequireFlag(val, name, hint string) error {
	if strings.TrimSpace(val) == "" {
		return FlagErrorf(hint, "flag %s is required", name)
	}
	return nil
}

// RequireAPIKey validates that an API key was resolved, returning a UsageError with defense guidance if missing.
func RequireAPIKey(key string) error {
	if strings.TrimSpace(key) == "" {
		return FlagErrorf(
			"set WARLOT_API_KEY environment variable, configure ~/.warlot/config.json (mode 0600), or enter key interactively",
			"authentication credentials required",
		)
	}
	return nil
}

