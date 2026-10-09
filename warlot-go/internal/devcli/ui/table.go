package ui

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
)

// Table wraps text/tabwriter for formatted tabular terminal output.
type Table struct {
	w       *tabwriter.Writer
	p       Painter
	headers []string
}

// NewTable initializes an aligned table writer.
func NewTable(out io.Writer, p Painter) *Table {
	return &Table{
		w: tabwriter.NewWriter(out, 0, 0, 3, ' ', 0),
		p: p,
	}
}

// SetHeaders formats and writes column headers.
func (t *Table) SetHeaders(headers ...string) {
	t.headers = headers
	styled := make([]string, len(headers))
	for i, h := range headers {
		styled[i] = t.p.Heading(strings.ToUpper(h))
	}
	fmt.Fprintln(t.w, strings.Join(styled, "\t"))
}

// AddRow writes a single row of values.
func (t *Table) AddRow(cols ...string) {
	fmt.Fprintln(t.w, strings.Join(cols, "\t"))
}

// Flush flushes all buffered rows to the output writer.
func (t *Table) Flush() error {
	return t.w.Flush()
}

// RenderRowMaps renders a slice of map rows into an aligned table.
func RenderRowMaps(out io.Writer, p Painter, rows []map[string]interface{}) error {
	if len(rows) == 0 {
		fmt.Fprintln(out, p.Muted("(0 rows)"))
		return nil
	}

	// Collect unique column keys in sorted order.
	keySet := make(map[string]struct{})
	for _, r := range rows {
		for k := range r {
			keySet[k] = struct{}{}
		}
	}
	cols := make([]string, 0, len(keySet))
	for k := range keySet {
		cols = append(cols, k)
	}
	sort.Strings(cols)

	tbl := NewTable(out, p)
	tbl.SetHeaders(cols...)

	for _, r := range rows {
		vals := make([]string, len(cols))
		for i, c := range cols {
			if val, ok := r[c]; ok && val != nil {
				vals[i] = fmt.Sprintf("%v", val)
			} else {
				vals[i] = p.Muted("NULL")
			}
		}
		tbl.AddRow(vals...)
	}

	return tbl.Flush()
}
