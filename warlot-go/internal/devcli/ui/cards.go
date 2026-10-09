package ui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/steven3002/warlot-golang-sdk/warlot-go/warlot"
)

// RenderStatusCard renders a structured SLI status card for a project.
func RenderStatusCard(out io.Writer, p Painter, g Glyphs, st *warlot.ProjectStatus) {
	if st == nil {
		fmt.Fprintln(out, p.Muted("(nil project status)"))
		return
	}

	// Status badge
	var statusBadge string
	switch strings.ToLower(st.Status) {
	case "active":
		statusBadge = p.OK(fmt.Sprintf("[%s ACTIVE]", g.Check()))
	case "deactivated":
		statusBadge = p.Warn(fmt.Sprintf("[%s DEACTIVATED]", g.Warn()))
	case "terminating":
		statusBadge = p.Warn(fmt.Sprintf("[%s TERMINATING]", g.Warn()))
	case "terminated":
		statusBadge = p.Err(fmt.Sprintf("[%s TERMINATED]", g.Cross()))
	default:
		statusBadge = p.Muted(fmt.Sprintf("[%s]", strings.ToUpper(st.Status)))
	}

	// Sync badge
	lag := st.ReplicationLag()
	var syncBadge string
	if st.Synced || lag == 0 {
		syncBadge = p.OK(fmt.Sprintf("[%s IN SYNC]", g.Check()))
	} else if lag < 50 {
		syncBadge = p.Warn(fmt.Sprintf("[%s LAGGING: %d ops]", g.Warn(), lag))
	} else {
		syncBadge = p.Err(fmt.Sprintf("[%s CRITICAL LAG: %d ops]", g.Cross(), lag))
	}

	// Format freshness
	freshnessStr := "unknown"
	if dur, ok := st.Freshness(); ok {
		freshnessStr = formatDuration(dur) + " ago"
	}

	fmt.Fprintf(out, "%s %s  %s\n", p.Heading("Project:"), p.Strong(st.ProjectID), statusBadge)
	fmt.Fprintf(out, "  %s  %s\n", p.Muted("DB ID:        "), st.DBID)
	if st.WriterPassID != "" {
		fmt.Fprintf(out, "  %s  %s\n", p.Muted("Writer Pass:  "), st.WriterPassID)
	}

	fmt.Fprintf(out, "\n%s\n", p.Heading("Durability & Replication SLIs:"))
	fmt.Fprintf(out, "  %s  %s\n", p.Muted("Sync State:   "), syncBadge)
	fmt.Fprintf(out, "  %s  %d\n", p.Muted("Anchored Seq: "), st.LastUploadSeq)
	fmt.Fprintf(out, "  %s  %d\n", p.Muted("Current Max:  "), st.MaxSeq)
	fmt.Fprintf(out, "  %s  %d mutations\n", p.Muted("Pending Lag:  "), lag)
	fmt.Fprintf(out, "  %s  %s\n", p.Muted("Last Anchor:  "), freshnessStr)

	if st.TerminationState != "" && st.TerminationState != "none" {
		fmt.Fprintf(out, "\n%s\n", p.Heading("Termination Progress:"))
		fmt.Fprintf(out, "  %s  %s\n", p.Muted("State:        "), st.TerminationState)
		fmt.Fprintf(out, "  %s  %d of %d\n", p.Muted("Files Shred:  "), st.TerminationFilesCompleted, st.TerminationFilesTotal)
	}
}

// RenderTableSchemaCard renders a table schema description.
func RenderTableSchemaCard(out io.Writer, p Painter, g Glyphs, schema *warlot.TableSchema) error {
	if schema == nil {
		fmt.Fprintln(out, p.Muted("(nil schema)"))
		return nil
	}

	fmt.Fprintf(out, "%s %s\n\n", p.Heading("Table:"), p.Strong(schema.Table))
	tbl := NewTable(out, p)
	tbl.SetHeaders("CID", "Name", "Type", "Not Null", "PK", "Default")

	for _, col := range schema.Columns {
		notNullStr := p.Muted("no")
		if col.NotNull {
			notNullStr = p.OK("yes")
		}
		pkStr := p.Muted("-")
		if col.PrimaryPK {
			pkStr = p.OK(g.Check())
		}
		defStr := p.Muted("NULL")
		if col.Default != nil {
			defStr = fmt.Sprintf("%v", col.Default)
		}

		tbl.AddRow(
			fmt.Sprintf("%d", col.CID),
			p.Strong(col.Name),
			col.Type,
			notNullStr,
			pkStr,
			defStr,
		)
	}

	return tbl.Flush()
}

// RenderReadinessCard renders the cluster health readiness diagnostics.
func RenderReadinessCard(out io.Writer, p Painter, g Glyphs, r *warlot.ReadinessResponse) {
	if r == nil {
		fmt.Fprintln(out, p.Muted("(nil readiness response)"))
		return
	}

	var statusBadge string
	if r.IsReady() {
		statusBadge = p.OK(fmt.Sprintf("[%s READY]", g.Check()))
	} else {
		statusBadge = p.Err(fmt.Sprintf("[%s NOT READY]", g.Cross()))
	}

	fmt.Fprintf(out, "%s %s\n\n", p.Heading("Cluster Health:"), statusBadge)
	tbl := NewTable(out, p)
	tbl.SetHeaders("Dependency", "Status")

	for _, d := range r.Dependencies {
		var depStatus string
		if d.Status == "ok" {
			depStatus = p.OK(fmt.Sprintf("%s ok", g.Check()))
		} else {
			depStatus = p.Err(fmt.Sprintf("%s %s", g.Cross(), d.Status))
		}
		tbl.AddRow(p.Strong(d.Name), depStatus)
	}
	_ = tbl.Flush()
}

func formatDuration(d time.Duration) string {
	d = d.Truncate(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%02ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%02dm", int(d.Hours()), int(d.Minutes())%60)
}
