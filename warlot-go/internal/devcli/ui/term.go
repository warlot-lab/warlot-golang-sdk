package ui

import (
	"os"
	"strings"

	"golang.org/x/term"
)

// Env captures the detected terminal capabilities and display preferences.
type Env struct {
	StdoutTTY bool
	StderrTTY bool
	StdinTTY  bool
	Color     bool
	Width     int
	Unicode   bool
}

// DetectEnv evaluates terminal capabilities, standards, and locale settings.
// mode can be "auto", "always", or "never".
func DetectEnv(mode string) Env {
	stdoutFD := int(os.Stdout.Fd())
	stderrFD := int(os.Stderr.Fd())
	stdinFD := int(os.Stdin.Fd())

	e := Env{
		StdoutTTY: term.IsTerminal(stdoutFD),
		StderrTTY: term.IsTerminal(stderrFD),
		StdinTTY:  term.IsTerminal(stdinFD),
		Width:     80,
	}

	// Determine color support.
	// NO_COLOR standard (http://no-color.org) takes precedence over terminal detection.
	switch {
	case mode == "never":
		e.Color = false
	case mode == "always":
		e.Color = true
	case os.Getenv("NO_COLOR") != "":
		e.Color = false
	case os.Getenv("FORCE_COLOR") != "":
		e.Color = true
	case os.Getenv("TERM") == "dumb":
		e.Color = false
	case !e.StdoutTTY:
		e.Color = false
	default:
		e.Color = true
	}

	// Query terminal width with fallback to 80.
	if w, _, err := term.GetSize(stdoutFD); err == nil && w > 0 {
		e.Width = w
	} else if w, _, err := term.GetSize(stderrFD); err == nil && w > 0 {
		e.Width = w
	}

	e.Unicode = isUTF8Locale()
	return e
}

// isUTF8Locale detects whether the system locale indicates UTF-8 support.
func isUTF8Locale() bool {
	for _, envVar := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		if val := strings.ToLower(os.Getenv(envVar)); val != "" {
			return strings.Contains(val, "utf-8") || strings.Contains(val, "utf8")
		}
	}
	return os.Getenv("WT_SESSION") != ""
}

// Glyphs provides character glyphs according to Unicode vs ASCII capability.
type Glyphs struct {
	Unicode bool
}

// DefaultGlyphs returns the glyph set for the given environment.
func DefaultGlyphs(unicode bool) Glyphs {
	return Glyphs{Unicode: unicode}
}

func (g Glyphs) Check() string {
	if g.Unicode {
		return "✓"
	}
	return "ok"
}

func (g Glyphs) Cross() string {
	if g.Unicode {
		return "✗"
	}
	return "x"
}

func (g Glyphs) Warn() string {
	return "!"
}

func (g Glyphs) Dot() string {
	if g.Unicode {
		return "•"
	}
	return "*"
}

func (g Glyphs) Bar() string {
	if g.Unicode {
		return "━"
	}
	return "-"
}

func (g Glyphs) Arrow() string {
	if g.Unicode {
		return "→"
	}
	return "->"
}
