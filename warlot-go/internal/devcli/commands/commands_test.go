package commands

import (
	"strings"
	"testing"

	"github.com/steven3002/warlot-golang-sdk/warlot-go/internal/devcli"
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
