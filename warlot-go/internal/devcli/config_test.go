package devcli

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseGlobalFlagsArgs_OmittedAPIKey_NoPanic_FallsBackToEnv(t *testing.T) {
	// Set test environment variable.
	const testKey = "wlt.1.env_secret_key_999"
	t.Setenv(EnvAPIKey, testKey)

	fs := flag.NewFlagSet("test-cmd", flag.ContinueOnError)
	args := []string{"-base", "https://api.test.example.com"}

	// Verify that omitting -apikey does not panic and populates from WARLOT_API_KEY.
	g := ParseGlobalFlagsArgs(fs, args)
	if g.Err != nil {
		t.Fatalf("unexpected error when -apikey is omitted: %v", g.Err)
	}
	if g.APIKey != testKey {
		t.Fatalf("expected APIKey to resolve to %q, got %q", testKey, g.APIKey)
	}
	if g.BaseURL != "https://api.test.example.com" {
		t.Fatalf("expected BaseURL to resolve to custom flag value, got %q", g.BaseURL)
	}
}

func TestParseGlobalFlagsArgs_PlaintextAPIKeyFlag_Rejected_CWE214(t *testing.T) {
	// Ensure env variable is clear.
	t.Setenv(EnvAPIKey, "")

	fs := flag.NewFlagSet("test-cmd", flag.ContinueOnError)
	args := []string{"-apikey", "wlt.1.plaintext_leaked_key"}

	g := ParseGlobalFlagsArgs(fs, args)
	// Passing plaintext API key via CLI argument must be refused.
	if g.Err == nil {
		t.Fatal("expected error when passing -apikey flag argument, but got nil")
	}
	if !strings.Contains(g.Err.Error(), "CWE-214") && !strings.Contains(g.Err.Error(), "prohibited") {
		t.Fatalf("expected CWE-214 rejection error, got: %v", g.Err)
	}
	if g.APIKey != "" {
		t.Fatalf("expected APIKey not to accept plaintext argument, but got: %q", g.APIKey)
	}
}

func TestLoadConfigFile_Mode0600_Accepted(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	content := []byte(`{"base_url": "https://config.example.com", "api_key": "wlt.1.file_key_0600"}`)
	if err := os.WriteFile(cfgPath, content, 0600); err != nil {
		t.Fatalf("failed to write test config file: %v", err)
	}

	cfg := loadConfigFileFromPath(cfgPath)
	if cfg.APIKey != "wlt.1.file_key_0600" {
		t.Fatalf("expected config to load API key from mode 0600 file, got: %q", cfg.APIKey)
	}
	if cfg.BaseURL != "https://config.example.com" {
		t.Fatalf("expected BaseURL from config file, got: %q", cfg.BaseURL)
	}
}

func TestLoadConfigFile_InsecurePermissions_Rejected(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	content := []byte(`{"api_key": "wlt.1.insecure_key"}`)
	// Permissions 0644 allow group/other read access and must be rejected.
	if err := os.WriteFile(cfgPath, content, 0644); err != nil {
		t.Fatalf("failed to write test config file: %v", err)
	}

	cfg := loadConfigFileFromPath(cfgPath)
	if cfg.APIKey != "" {
		t.Fatalf("expected insecure config file (0644) to be rejected, but loaded key: %q", cfg.APIKey)
	}
}

func TestResolveAPIKey_EnvTakesPrecedenceOverConfigFile(t *testing.T) {
	tmpHome := t.TempDir()
	warlotDir := filepath.Join(tmpHome, ".warlot")
	if err := os.MkdirAll(warlotDir, 0700); err != nil {
		t.Fatalf("failed to create fake home dir: %v", err)
	}
	cfgPath := filepath.Join(warlotDir, "config.json")
	if err := os.WriteFile(cfgPath, []byte(`{"api_key": "wlt.1.config_key"}`), 0600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	t.Setenv("HOME", tmpHome)
	t.Setenv(EnvAPIKey, "wlt.1.env_priority_key")

	key := ResolveAPIKey()
	if key != "wlt.1.env_priority_key" {
		t.Fatalf("expected env variable to take precedence, got: %q", key)
	}

	// Unset env and verify fallback to config file
	t.Setenv(EnvAPIKey, "")
	keyFallback := ResolveAPIKey()
	if keyFallback != "wlt.1.config_key" {
		t.Fatalf("expected fallback to config file, got: %q", keyFallback)
	}
}

func TestRequireAPIKey_Validation(t *testing.T) {
	if err := RequireAPIKey(""); err == nil {
		t.Fatal("expected error for empty API key")
	}
	if err := RequireAPIKey("   "); err == nil {
		t.Fatal("expected error for whitespace API key")
	}
	if err := RequireAPIKey("wlt.1.valid_key"); err != nil {
		t.Fatalf("unexpected error for valid API key: %v", err)
	}
}
