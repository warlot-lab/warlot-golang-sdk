package commands

import (
	"errors"
	"strings"
	"testing"

	"github.com/warlot-lab/warlot-golang-sdk/warlot-go/internal/devcli"
)

func TestCommands_RunSQL_RejectsAPIKeyFlag(t *testing.T) {
	err := RunSQL([]string{"-project", "proj-1", "-q", "SELECT 1", "-apikey", "wlt.1.plain"})
	if err == nil {
		t.Fatal("expected RunSQL to reject -apikey flag argument")
	}
	if !strings.Contains(err.Error(), "CWE-214") && !strings.Contains(err.Error(), "prohibited") {
		t.Fatalf("expected CWE-214 error, got: %v", err)
	}
}

func TestCommands_RunSQL_MissingRequiredFlags(t *testing.T) {
	// Missing -project
	err := RunSQL([]string{"-q", "SELECT 1"})
	if err == nil {
		t.Fatal("expected error for missing -project")
	}

	// Missing -q
	err = RunSQL([]string{"-project", "proj-1"})
	if err == nil {
		t.Fatal("expected error for missing -q")
	}

	// Missing authentication credentials
	t.Setenv(devcli.EnvAPIKey, "")
	err = RunSQL([]string{"-project", "proj-1", "-q", "SELECT 1"})
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
	if !strings.Contains(err.Error(), "credentials required") {
		t.Fatalf("expected credentials required error, got: %v", err)
	}
}

func TestCommands_RunCommit_RejectsAPIKeyFlag(t *testing.T) {
	err := RunCommit([]string{"-project", "proj-1", "-apikey", "wlt.1.plain"})
	if err == nil {
		t.Fatal("expected RunCommit to reject -apikey flag argument")
	}
	if !strings.Contains(err.Error(), "CWE-214") {
		t.Fatalf("expected CWE-214 error, got: %v", err)
	}
}

func TestCommands_RunStatus_RejectsAPIKeyFlag(t *testing.T) {
	err := RunStatus([]string{"-project", "proj-1", "-apikey", "wlt.1.plain"})
	if err == nil {
		t.Fatal("expected RunStatus to reject -apikey flag argument")
	}
	if !strings.Contains(err.Error(), "CWE-214") {
		t.Fatalf("expected CWE-214 error, got: %v", err)
	}
}

func TestCommands_RunTables_RejectsAPIKeyFlag(t *testing.T) {
	err := RunTables([]string{"list", "-project", "proj-1", "-apikey", "wlt.1.plain"})
	if err == nil {
		t.Fatal("expected RunTables to reject -apikey flag argument")
	}
	if !strings.Contains(err.Error(), "CWE-214") {
		t.Fatalf("expected CWE-214 error, got: %v", err)
	}
}

func TestCommands_MissingFlagStructuredError_NoPanic(t *testing.T) {
	// 1. RunStatus without -project
	err := RunStatus([]string{})
	if err == nil {
		t.Fatal("expected error for RunStatus without -project")
	}
	var uErr *devcli.UsageError
	if !errors.As(err, &uErr) || uErr.Hint == "" {
		t.Errorf("expected UsageError with hint, got: %v", err)
	}
	if !strings.Contains(err.Error(), "flag -project is required") {
		t.Errorf("expected missing flag error, got: %v", err)
	}

	// 2. RunTables without subcommand
	err = RunTables([]string{})
	if err == nil {
		t.Fatal("expected error for RunTables without subcommand")
	}
	if !strings.Contains(err.Error(), "subcommand is required") {
		t.Errorf("expected subcommand required error, got: %v", err)
	}

	// 3. RunTables browse without required flags
	err = RunTables([]string{"browse"})
	if err == nil {
		t.Fatal("expected error for RunTables browse without flags")
	}
	if !strings.Contains(err.Error(), "flag -project is required") {
		t.Errorf("expected -project required error, got: %v", err)
	}

	// 4. RunResolve without required flags
	t.Setenv(devcli.EnvHolderID, "")
	t.Setenv(devcli.EnvProjectName, "")
	err = RunResolve([]string{})
	if err == nil {
		t.Fatal("expected error for RunResolve without flags")
	}
	if !strings.Contains(err.Error(), "flag -holder is required") {
		t.Errorf("expected -holder required error, got: %v", err)
	}
}

func TestCommands_FlagParsing_DoubleDashAndSingleDash(t *testing.T) {
	// Verify that both --json and -json are parsed without error
	t.Setenv(devcli.EnvAPIKey, "wlt.1.testkey")
	err1 := RunStatus([]string{"-project", "proj-1", "-json"})
	err2 := RunStatus([]string{"-project", "proj-1", "--json"})

	// Since backend may not be mocked here, we only verify that flag parsing succeeded
	// and neither triggered a flag parse error
	if err1 != nil && strings.Contains(err1.Error(), "flag provided but not defined") {
		t.Errorf("unexpected flag error for -json: %v", err1)
	}
	if err2 != nil && strings.Contains(err2.Error(), "flag provided but not defined") {
		t.Errorf("unexpected flag error for --json: %v", err2)
	}
}
