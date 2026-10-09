package ui

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/warlot-lab/warlot-golang-sdk/warlot-go/warlot"
)

func TestPainter_ColorOnAndOff(t *testing.T) {
	pColor := NewPainter(true)
	pPlain := NewPainter(false)

	// Colored output should contain SGR escapes
	okStyled := pColor.OK("SUCCESS")
	if !strings.Contains(okStyled, "\x1b[32m") || !strings.Contains(okStyled, "\x1b[0m") {
		t.Errorf("expected ANSI green escape, got %q", okStyled)
	}

	errStyled := pColor.Err("FAILED")
	if !strings.Contains(errStyled, "\x1b[31m") {
		t.Errorf("expected ANSI red escape, got %q", errStyled)
	}

	// Plain output should not contain escape characters
	plainOK := pPlain.OK("SUCCESS")
	if plainOK != "SUCCESS" {
		t.Errorf("expected plain text SUCCESS, got %q", plainOK)
	}
	if strings.Contains(plainOK, "\x1b") {
		t.Errorf("plain output must not contain ANSI escape codes: %q", plainOK)
	}
}

func TestPainter_FormatError(t *testing.T) {
	p := NewPainter(false)
	msg := p.FormatError(errors.New("connection failed"), "check network connectivity")

	if !strings.Contains(msg, "error: connection failed") {
		t.Errorf("expected error line, got %q", msg)
	}
	if !strings.Contains(msg, "hint: check network connectivity") {
		t.Errorf("expected hint line, got %q", msg)
	}
}

func TestDetectEnv_NoColor(t *testing.T) {
	_ = os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	env := DetectEnv("auto")
	if env.Color {
		t.Errorf("expected Color to be false when NO_COLOR=1, got true")
	}
}

func TestRenderRowMaps(t *testing.T) {
	p := NewPainter(false)
	var buf bytes.Buffer

	rows := []map[string]interface{}{
		{"id": 1, "name": "Alice", "role": "admin"},
		{"id": 2, "name": "Bob", "role": "user"},
	}

	if err := RenderRowMaps(&buf, p, rows); err != nil {
		t.Fatalf("RenderRowMaps failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "ID") || !strings.Contains(out, "NAME") || !strings.Contains(out, "ROLE") {
		t.Errorf("missing headers in table output:\n%s", out)
	}
	if !strings.Contains(out, "Alice") || !strings.Contains(out, "Bob") {
		t.Errorf("missing rows in table output:\n%s", out)
	}
}

func TestRenderStatusCard(t *testing.T) {
	p := NewPainter(false)
	g := DefaultGlyphs(false)
	var buf bytes.Buffer

	lastAt := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	st := &warlot.ProjectStatus{
		ProjectID:     "proj-xyz",
		DBID:          "db-123",
		Status:        "active",
		LastUploadSeq: 100,
		MaxSeq:        105,
		Synced:        false,
		LastUploadAt:  &lastAt,
	}

	RenderStatusCard(&buf, p, g, st)
	out := buf.String()

	if !strings.Contains(out, "proj-xyz") {
		t.Errorf("missing project ID in card:\n%s", out)
	}
	if !strings.Contains(out, "Anchored Seq:") || !strings.Contains(out, "100") {
		t.Errorf("missing anchored seq in card:\n%s", out)
	}
	if !strings.Contains(out, "Current Max:") || !strings.Contains(out, "105") {
		t.Errorf("missing max seq in card:\n%s", out)
	}
	if !strings.Contains(out, "Pending Lag:") || !strings.Contains(out, "5 mutations") {
		t.Errorf("missing lag in card:\n%s", out)
	}
	if !strings.Contains(out, "[!] LAGGING: 5") {
		t.Errorf("expected lagging badge [!] LAGGING: 5 in card:\n%s", out)
	}
}

func TestPainter_16ColorsSemanticRoles(t *testing.T) {
	p := NewPainter(true)

	tests := []struct {
		name     string
		role     Role
		output   string
		expected string
	}{
		{"RoleOK", RoleOK, p.OK("ok"), "\x1b[32mok\x1b[0m"},
		{"RoleWarn", RoleWarn, p.Warn("warn"), "\x1b[33mwarn\x1b[0m"},
		{"RoleErr", RoleErr, p.Err("err"), "\x1b[31merr\x1b[0m"},
		{"RoleHint", RoleHint, p.Hint("hint"), "\x1b[36mhint\x1b[0m"},
		{"RoleMuted", RoleMuted, p.Muted("muted"), "\x1b[2mmuted\x1b[0m"},
		{"RoleStrong", RoleStrong, p.Strong("strong"), "\x1b[1mstrong\x1b[0m"},
		{"RoleHeading", RoleHeading, p.Heading("heading"), "\x1b[1;36mheading\x1b[0m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.output != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.output)
			}
		})
	}

	// Verify label bold application
	errLabel := p.Label(RoleErr, "error:")
	if errLabel != "\x1b[1;31merror:\x1b[0m" {
		t.Errorf("expected bold red label, got %q", errLabel)
	}
	hintLabel := p.Label(RoleHint, "hint:")
	if hintLabel != "\x1b[1;36mhint:\x1b[0m" {
		t.Errorf("expected bold cyan label, got %q", hintLabel)
	}
}

func TestDetectEnv_PrecedenceAndStandards(t *testing.T) {
	// 1. NO_COLOR wins over FORCE_COLOR
	t.Setenv("NO_COLOR", "1")
	t.Setenv("FORCE_COLOR", "1")
	env := DetectEnv("auto")
	if env.Color {
		t.Errorf("expected NO_COLOR to take precedence over FORCE_COLOR, but Color is true")
	}

	// 2. FORCE_COLOR forces color when NO_COLOR is empty
	t.Setenv("NO_COLOR", "")
	t.Setenv("FORCE_COLOR", "1")
	env = DetectEnv("auto")
	if !env.Color {
		t.Errorf("expected FORCE_COLOR=1 to enable color, got false")
	}

	// 3. TERM=dumb disables color
	t.Setenv("FORCE_COLOR", "")
	t.Setenv("TERM", "dumb")
	env = DetectEnv("auto")
	if env.Color {
		t.Errorf("expected TERM=dumb to disable color, got true")
	}

	// 4. mode="never" overrides all
	t.Setenv("FORCE_COLOR", "1")
	env = DetectEnv("never")
	if env.Color {
		t.Errorf("expected mode=never to disable color, got true")
	}

	// 5. mode="always" overrides all
	t.Setenv("NO_COLOR", "1")
	env = DetectEnv("always")
	if !env.Color {
		t.Errorf("expected mode=always to enable color, got false")
	}
}

func TestGlyphs_UnicodeVsASCII(t *testing.T) {
	u := DefaultGlyphs(true)
	if u.Check() != "✓" || u.Cross() != "✗" || u.Bar() != "━" || u.Warn() != "!" {
		t.Errorf("unexpected unicode glyphs: check=%s cross=%s bar=%s warn=%s", u.Check(), u.Cross(), u.Bar(), u.Warn())
	}

	a := DefaultGlyphs(false)
	if a.Check() != "ok" || a.Cross() != "x" || a.Bar() != "-" || a.Warn() != "!" {
		t.Errorf("unexpected ascii glyphs: check=%s cross=%s bar=%s warn=%s", a.Check(), a.Cross(), a.Bar(), a.Warn())
	}
}

func TestRenderStatusCard_SyncConvergenceBadges(t *testing.T) {
	p := NewPainter(false)
	gUnicode := DefaultGlyphs(true)
	gASCII := DefaultGlyphs(false)

	// In-sync status
	stInSync := &warlot.ProjectStatus{
		ProjectID:     "proj-sync",
		Status:        "active",
		LastUploadSeq: 200,
		MaxSeq:        200,
		Synced:        true,
	}
	var bufU bytes.Buffer
	RenderStatusCard(&bufU, p, gUnicode, stInSync)
	if !strings.Contains(bufU.String(), "[✓] IN SYNC") {
		t.Errorf("expected [✓] IN SYNC in unicode card:\n%s", bufU.String())
	}
	if !strings.Contains(bufU.String(), "[✓] ACTIVE") {
		t.Errorf("expected [✓] ACTIVE in unicode card:\n%s", bufU.String())
	}

	var bufA bytes.Buffer
	RenderStatusCard(&bufA, p, gASCII, stInSync)
	if !strings.Contains(bufA.String(), "[ok] IN SYNC") {
		t.Errorf("expected [ok] IN SYNC in ascii card:\n%s", bufA.String())
	}
	if !strings.Contains(bufA.String(), "[ok] ACTIVE") {
		t.Errorf("expected [ok] ACTIVE in ascii card:\n%s", bufA.String())
	}

	// Critical lag status
	stCritLag := &warlot.ProjectStatus{
		ProjectID:     "proj-crit",
		Status:        "terminated",
		LastUploadSeq: 100,
		MaxSeq:        250,
		Synced:        false,
	}
	var bufCrit bytes.Buffer
	RenderStatusCard(&bufCrit, p, gUnicode, stCritLag)
	if !strings.Contains(bufCrit.String(), "[✗] CRITICAL LAG: 150") {
		t.Errorf("expected [✗] CRITICAL LAG: 150 in unicode card:\n%s", bufCrit.String())
	}
	if !strings.Contains(bufCrit.String(), "[✗] TERMINATED") {
		t.Errorf("expected [✗] TERMINATED in unicode card:\n%s", bufCrit.String())
	}
}

func TestRenderTableSchemaCard(t *testing.T) {
	p := NewPainter(false)
	g := DefaultGlyphs(true)

	schema := &warlot.TableSchema{
		Table: "products",
		Columns: []warlot.TableColumn{
			{CID: 0, Name: "id", Type: "INTEGER", NotNull: true, PrimaryPK: true, Default: nil},
			{CID: 1, Name: "name", Type: "TEXT", NotNull: true, PrimaryPK: false, Default: nil},
			{CID: 2, Name: "price", Type: "REAL", NotNull: false, PrimaryPK: false, Default: "0.0"},
		},
	}

	var buf bytes.Buffer
	if err := RenderTableSchemaCard(&buf, p, g, schema); err != nil {
		t.Fatalf("RenderTableSchemaCard failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Table: products") {
		t.Errorf("missing table title:\n%s", out)
	}
	if !strings.Contains(out, "CID") || !strings.Contains(out, "NAME") || !strings.Contains(out, "PK") {
		t.Errorf("missing headers in schema table:\n%s", out)
	}
	if !strings.Contains(out, "id") || !strings.Contains(out, "INTEGER") || !strings.Contains(out, "✓") {
		t.Errorf("missing id row details in schema table:\n%s", out)
	}
	if !strings.Contains(out, "price") || !strings.Contains(out, "0.0") {
		t.Errorf("missing price row details in schema table:\n%s", out)
	}
}

func TestRenderReadinessCard(t *testing.T) {
	p := NewPainter(false)
	g := DefaultGlyphs(true)

	readyResp := &warlot.ReadinessResponse{
		Status: "ready",
		Dependencies: []warlot.DependencyStatus{
			{Name: "postgres", Status: "ok"},
			{Name: "aggregator", Status: "ok"},
		},
	}

	var buf bytes.Buffer
	RenderReadinessCard(&buf, p, g, readyResp)
	out := buf.String()

	if !strings.Contains(out, "Cluster Health: [✓] READY") {
		t.Errorf("missing [✓] READY badge:\n%s", out)
	}
	if !strings.Contains(out, "postgres") || !strings.Contains(out, "aggregator") {
		t.Errorf("missing dependency names:\n%s", out)
	}
}
