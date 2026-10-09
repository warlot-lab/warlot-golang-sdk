package devcli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"
)

// Environment keys for defaults.
const (
	EnvBaseURL     = "WARLOT_BASE_URL"
	EnvAPIKey      = "WARLOT_API_KEY"
	EnvHolderID    = "WARLOT_HOLDER"
	EnvProjectName = "WARLOT_PNAME"

	EnvTimeoutSec  = "WARLOT_TIMEOUT"         // seconds
	EnvRetries     = "WARLOT_RETRIES"         // int
	EnvBackoffInit = "WARLOT_BACKOFF_INIT_MS" // ms
	EnvBackoffMax  = "WARLOT_BACKOFF_MAX_MS"  // ms
)

// Reasonable defaults for production-grade operation.
const (
	DefaultTimeoutSec  = 90
	DefaultRetries     = 5
	DefaultBackoffInit = 1000 // ms
	DefaultBackoffMax  = 8000 // ms
)

// GlobalFlags captures CLI-wide settings and defaults.
type GlobalFlags struct {
	BaseURL     string
	APIKey      string
	HolderID    string
	ProjectName string

	Timeout     time.Duration
	Retries     int
	BackoffInit time.Duration
	BackoffMax  time.Duration
	Verbose     bool
	JSON        bool
	Err         error
}

// ParseGlobalFlagsArgs binds global flags to the provided FlagSet and parses args.
func ParseGlobalFlagsArgs(fs *flag.FlagSet, args []string) GlobalFlags {
	var g GlobalFlags

	fileCfg := loadConfigFile()

	// Defaults sourced from environment variables, config file, then fallback.
	defBase := getenvDefault(EnvBaseURL, fileCfg.BaseURL)
	if defBase == "" {
		defBase = "https://api.warlot.stevenhert.xyz"
	}
	defHolder := getenvDefault(EnvHolderID, fileCfg.HolderID)
	defPname := getenvDefault(EnvProjectName, fileCfg.ProjectName)

	defTO := time.Duration(atoiDefault(os.Getenv(EnvTimeoutSec), DefaultTimeoutSec)) * time.Second
	defRet := atoiDefault(os.Getenv(EnvRetries), DefaultRetries)
	defBInit := durMsDefault(os.Getenv(EnvBackoffInit), time.Duration(DefaultBackoffInit)*time.Millisecond)
	defBMax := durMsDefault(os.Getenv(EnvBackoffMax), time.Duration(DefaultBackoffMax)*time.Millisecond)

	fs.StringVar(&g.BaseURL, "base", defBase, "API base URL (env "+EnvBaseURL+")")
	fs.StringVar(&g.HolderID, "holder", defHolder, "Holder ID (env "+EnvHolderID+")")
	fs.StringVar(&g.ProjectName, "pname", defPname, "Project name (env "+EnvProjectName+")")

	var apikeyFlag string
	fs.StringVar(&apikeyFlag, "apikey", "", "API key (disabled as CLI argument for CWE-214 defense; set WARLOT_API_KEY)")

	timeoutSec := fs.Int("timeout", int(defTO/time.Second), "Request timeout seconds (env "+EnvTimeoutSec+")")
	fs.IntVar(&g.Retries, "retries", defRet, "Max retries on 429/5xx (env "+EnvRetries+")")

	backoffInit := fs.Int("backoff-init", int(defBInit/time.Millisecond), "Initial backoff ms (env "+EnvBackoffInit+")")
	backoffMax := fs.Int("backoff-max", int(defBMax/time.Millisecond), "Max backoff ms (env "+EnvBackoffMax+")")

	fs.BoolVar(&g.Verbose, "v", false, "Verbose request/response logs (credentials redacted)")
	fs.BoolVar(&g.JSON, "json", false, "Output raw machine-readable JSON")

	// Parse now.
	if err := fs.Parse(args); err != nil {
		g.Err = err
		return g
	}

	// Secret defense: prohibit passing API keys as CLI arguments (CWE-214).
	if strings.TrimSpace(apikeyFlag) != "" {
		g.Err = FlagErrorf(
			"set WARLOT_API_KEY environment variable, configure ~/.warlot/config.json (mode 0600), or enter key interactively",
			"passing plaintext API keys via -apikey CLI arguments is prohibited (CWE-214)",
		)
	} else {
		g.APIKey = ResolveAPIKey()
	}

	// Finalize computed durations.
	g.Timeout = time.Duration(*timeoutSec) * time.Second
	g.BackoffInit = time.Duration(*backoffInit) * time.Millisecond
	g.BackoffMax = time.Duration(*backoffMax) * time.Millisecond

	return g
}

// EnsureAPIKey guarantees that an API key is present for authenticated commands.
// If not resolved from WARLOT_API_KEY or ~/.warlot/config.json (mode 0600),
// it prompts the operator masked on a TTY using golang.org/x/term.ReadPassword.
func EnsureAPIKey(g *GlobalFlags) error {
	if g == nil {
		return FlagErrorf("client configuration is nil", "internal error")
	}
	if strings.TrimSpace(g.APIKey) != "" {
		return nil
	}
	if term.IsTerminal(int(os.Stdin.Fd())) {
		g.APIKey = promptMaskedAPIKey()
		if strings.TrimSpace(g.APIKey) != "" {
			return nil
		}
	}
	return RequireAPIKey(g.APIKey)
}

// ResolveAPIKey resolves the API key prioritizing WARLOT_API_KEY, falling back to
// ~/.warlot/config.json (mode 0600).
func ResolveAPIKey() string {
	// 1. Environment variable
	if key := strings.TrimSpace(os.Getenv(EnvAPIKey)); key != "" {
		return key
	}

	// 2. Fallback to ~/.warlot/config.json (mode 0600)
	cfg := loadConfigFile()
	if key := strings.TrimSpace(cfg.APIKey); key != "" {
		return key
	}

	return ""
}

func promptMaskedAPIKey() string {
	fmt.Fprint(os.Stderr, "Enter Warlot API Key: ")
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// MustNonEmpty enforces required flag presence for better operator feedback without panicking.
func MustNonEmpty(val, name string) {
	if strings.TrimSpace(val) == "" {
		ExitWithUsageError(fmt.Errorf("missing required %s", name), "provide "+name+" flag or set corresponding environment variable")
	}
}

// Helpers

func getenvDefault(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func atoiDefault(s string, d int) int {
	if s == "" {
		return d
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return d
	}
	return i
}

func durMsDefault(msStr string, d time.Duration) time.Duration {
	if msStr == "" {
		return d
	}
	ms, err := strconv.Atoi(msStr)
	if err != nil {
		return d
	}
	return time.Duration(ms) * time.Millisecond
}

type configFile struct {
	BaseURL     string `json:"base_url,omitempty"`
	APIKey      string `json:"api_key,omitempty"`
	HolderID    string `json:"holder_id,omitempty"`
	ProjectName string `json:"project_name,omitempty"`
}

func loadConfigFile() configFile {
	home, err := os.UserHomeDir()
	if err != nil {
		return configFile{}
	}
	p := filepath.Join(home, ".warlot", "config.json")
	return loadConfigFileFromPath(p)
}

func loadConfigFileFromPath(p string) configFile {
	fi, err := os.Stat(p)
	if err != nil {
		return configFile{}
	}
	// Verify that the file is a regular file and strictly protected (mode 0600).
	// Permissions must not allow group or other access (CWE-214 secret defense).
	if !fi.Mode().IsRegular() || fi.Mode().Perm()&0077 != 0 || fi.Mode().Perm() > 0600 {
		return configFile{}
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return configFile{}
	}
	var cfg configFile
	_ = json.Unmarshal(b, &cfg)
	return cfg
}
