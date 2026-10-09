package ui

import "fmt"

// Role defines the semantic meaning of a styled string.
type Role int

const (
	RoleOK Role = iota
	RoleWarn
	RoleErr
	RoleHint
	RoleMuted
	RoleStrong
	RoleHeading

	// Semantic role aliases matching Sennit heuristics.
	rOK      = RoleOK
	rWarn    = RoleWarn
	rErr     = RoleErr
	rHint    = RoleHint
	rMuted   = RoleMuted
	rStrong  = RoleStrong
	rHeading = RoleHeading

	ROK      = RoleOK
	RWarn    = RoleWarn
	RErr     = RoleErr
	RHint    = RoleHint
	RMuted   = RoleMuted
	RStrong  = RoleStrong
	RHeading = RoleHeading
)

// Painter translates semantic roles into 16 ANSI escape sequences.
type Painter struct {
	color bool
	sgr   map[Role]string
}

// NewPainter initializes a painter configured with the 16 ANSI semantic color system.
func NewPainter(color bool) Painter {
	return Painter{
		color: color,
		sgr: map[Role]string{
			RoleOK:      "32",
			RoleWarn:    "33",
			RoleErr:     "31",
			RoleHint:    "36",
			RoleMuted:   "2",
			RoleStrong:  "1",
			RoleHeading: "1;36",
		},
	}
}

// Paint applies the ANSI styling for the specified semantic role.
func (p Painter) Paint(r Role, s string) string {
	if !p.color || s == "" || p.sgr[r] == "" {
		return s
	}
	return "\x1b[" + p.sgr[r] + "m" + s + "\x1b[0m"
}

// Label applies bold styling on the semantic color for category labels.
func (p Painter) Label(r Role, s string) string {
	if !p.color || s == "" {
		return s
	}
	code := p.sgr[r]
	if code == "1" || code == "2" {
		return "\x1b[" + code + "m" + s + "\x1b[0m"
	}
	return "\x1b[1;" + code + "m" + s + "\x1b[0m"
}

// Semantic helper methods.
func (p Painter) OK(s string) string      { return p.Paint(RoleOK, s) }
func (p Painter) Warn(s string) string    { return p.Paint(RoleWarn, s) }
func (p Painter) Err(s string) string     { return p.Paint(RoleErr, s) }
func (p Painter) Hint(s string) string    { return p.Paint(RoleHint, s) }
func (p Painter) Muted(s string) string   { return p.Paint(RoleMuted, s) }
func (p Painter) Strong(s string) string  { return p.Paint(RoleStrong, s) }
func (p Painter) Heading(s string) string { return p.Paint(RoleHeading, s) }

// Role-prefixed helper aliases.
func (p Painter) ROK(s string) string      { return p.OK(s) }
func (p Painter) RWarn(s string) string    { return p.Warn(s) }
func (p Painter) RErr(s string) string     { return p.Err(s) }
func (p Painter) RHint(s string) string    { return p.Hint(s) }
func (p Painter) RMuted(s string) string   { return p.Muted(s) }
func (p Painter) RStrong(s string) string  { return p.Strong(s) }
func (p Painter) RHeading(s string) string { return p.Heading(s) }

// FormatError formats an error with an optional diagnostic hint according to Sennit CLI heuristics.
func (p Painter) FormatError(err error, hint string) string {
	msg := fmt.Sprintf("%s %s\n", p.Label(RoleErr, "error:"), err.Error())
	if hint != "" {
		msg += fmt.Sprintf("%s %s\n", p.Label(RoleHint, "hint:"), hint)
	}
	return msg
}
