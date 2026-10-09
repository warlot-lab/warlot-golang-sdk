package ui

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

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

	lastAt := "2026-10-09T12:00:00Z"
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
}
