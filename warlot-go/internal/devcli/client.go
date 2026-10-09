package devcli

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/warlot-lab/warlot-golang-sdk/warlot-go/warlot"
)

// NewClient constructs an SDK client using global flags.
func NewClient(g GlobalFlags) *warlot.Client {
	opts := []warlot.Option{
		warlot.WithBaseURL(g.BaseURL),
		warlot.WithHolderID(g.HolderID),
		warlot.WithProjectName(g.ProjectName),
		warlot.WithRetries(g.Retries),
		warlot.WithBackoff(g.BackoffInit, g.BackoffMax),
		warlot.WithHTTPClient(&http.Client{Timeout: g.Timeout}),
	}
	if g.APIKey != "" {
		opts = append(opts, warlot.WithAPIKey(g.APIKey))
	}
	if g.Verbose {
		opts = append(opts, warlot.WithLogger(func(event string, meta map[string]any) {
			// Logger redacts the API key; output directed to stderr for visibility.
			fmt.Fprintf(os.Stderr, "%s: %v\n", event, meta)
		}))
	}
	return warlot.New(opts...)
}

// Ctx returns a context with the CLI-configured timeout.
func Ctx(g GlobalFlags) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), g.Timeout)
}
