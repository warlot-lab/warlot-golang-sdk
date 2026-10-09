package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/steven3002/warlot-golang-sdk/warlot-go/internal/devcli"
	"github.com/steven3002/warlot-golang-sdk/warlot-go/internal/devcli/ui"
)

// RunTables dispatches to list|browse|schema|count subcommands.
func RunTables(args []string) error {
	if len(args) == 0 {
		return devcli.FlagErrorf("run 'warlotdev tables <list|browse|schema|count> [flags]'", "subcommand is required")
	}
	switch args[0] {
	case "list":
		return runTablesList(args[1:])
	case "browse":
		return runTablesBrowse(args[1:])
	case "schema":
		return runTablesSchema(args[1:])
	case "count":
		return runTablesCount(args[1:])
	default:
		return devcli.FlagErrorf("valid subcommands are list, browse, schema, count", "unknown tables subcommand %q", args[0])
	}
}

func runTablesList(args []string) error {
	fs := flag.NewFlagSet("tables list", flag.ContinueOnError)
	projectID := fs.String("project", "", "Project ID (required)")
	g := devcli.ParseGlobalFlagsArgs(fs, args)

	if err := devcli.RequireFlag(*projectID, "-project", "provide -project <id> to list tables"); err != nil {
		return err
	}
	if err := devcli.RequireFlag(g.APIKey, "-apikey", "provide -apikey or set WARLOT_API_KEY"); err != nil {
		return err
	}

	cl := devcli.NewClient(g)
	ctx, cancel := devcli.Ctx(g)
	defer cancel()

	out, err := cl.ListTables(ctx, *projectID)
	if err != nil {
		return err
	}

	if g.JSON {
		devcli.PrintJSON(out)
		return nil
	}

	env := ui.DetectEnv("auto")
	painter := ui.NewPainter(env.Color)
	glyphs := ui.DefaultGlyphs(env.Unicode)

	if len(out.Tables) == 0 {
		fmt.Fprintln(os.Stdout, painter.Muted("(no tables)"))
		return nil
	}

	fmt.Fprintf(os.Stdout, "%s (%d tables)\n", painter.Heading("Tables:"), len(out.Tables))
	for _, t := range out.Tables {
		fmt.Fprintf(os.Stdout, "  %s %s\n", painter.Muted(glyphs.Dot()), painter.Strong(t))
	}
	return nil
}

func runTablesBrowse(args []string) error {
	fs := flag.NewFlagSet("tables browse", flag.ContinueOnError)
	projectID := fs.String("project", "", "Project ID (required)")
	table := fs.String("table", "", "Table name (required)")
	limit := fs.Int("limit", 10, "Row limit")
	offset := fs.Int("offset", 0, "Row offset")
	g := devcli.ParseGlobalFlagsArgs(fs, args)

	if err := devcli.RequireFlag(*projectID, "-project", "provide -project <id> to browse table"); err != nil {
		return err
	}
	if err := devcli.RequireFlag(*table, "-table", "provide -table <name> to browse table"); err != nil {
		return err
	}
	if err := devcli.RequireFlag(g.APIKey, "-apikey", "provide -apikey or set WARLOT_API_KEY"); err != nil {
		return err
	}

	cl := devcli.NewClient(g)
	ctx, cancel := devcli.Ctx(g)
	defer cancel()

	out, err := cl.BrowseRows(ctx, *projectID, *table, *limit, *offset)
	if err != nil {
		return err
	}

	if g.JSON {
		devcli.PrintJSON(out)
		return nil
	}

	env := ui.DetectEnv("auto")
	painter := ui.NewPainter(env.Color)
	glyphs := ui.DefaultGlyphs(env.Unicode)

	fmt.Fprintf(os.Stdout, "%s %s\n\n", painter.Heading("Table:"), painter.Strong(out.Table))
	if err := ui.RenderRowMaps(os.Stdout, painter, out.Rows); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "\n%s Showing %d row(s) (limit: %d, offset: %d)\n", painter.Muted(glyphs.Dot()), len(out.Rows), out.Limit, out.Offset)
	return nil
}

func runTablesSchema(args []string) error {
	fs := flag.NewFlagSet("tables schema", flag.ContinueOnError)
	projectID := fs.String("project", "", "Project ID (required)")
	table := fs.String("table", "", "Table name (required)")
	g := devcli.ParseGlobalFlagsArgs(fs, args)

	if err := devcli.RequireFlag(*projectID, "-project", "provide -project <id> to view table schema"); err != nil {
		return err
	}
	if err := devcli.RequireFlag(*table, "-table", "provide -table <name> to view table schema"); err != nil {
		return err
	}
	if err := devcli.RequireFlag(g.APIKey, "-apikey", "provide -apikey or set WARLOT_API_KEY"); err != nil {
		return err
	}

	cl := devcli.NewClient(g)
	ctx, cancel := devcli.Ctx(g)
	defer cancel()

	out, err := cl.GetTableSchema(ctx, *projectID, *table)
	if err != nil {
		return err
	}

	if g.JSON {
		devcli.PrintJSON(out)
		return nil
	}

	env := ui.DetectEnv("auto")
	painter := ui.NewPainter(env.Color)
	glyphs := ui.DefaultGlyphs(env.Unicode)
	return ui.RenderTableSchemaCard(os.Stdout, painter, glyphs, out)
}

func runTablesCount(args []string) error {
	fs := flag.NewFlagSet("tables count", flag.ContinueOnError)
	projectID := fs.String("project", "", "Project ID (required)")
	g := devcli.ParseGlobalFlagsArgs(fs, args)

	if err := devcli.RequireFlag(*projectID, "-project", "provide -project <id> to count tables"); err != nil {
		return err
	}
	if err := devcli.RequireFlag(g.APIKey, "-apikey", "provide -apikey or set WARLOT_API_KEY"); err != nil {
		return err
	}

	cl := devcli.NewClient(g)
	ctx, cancel := devcli.Ctx(g)
	defer cancel()

	out, err := cl.GetTableCount(ctx, *projectID)
	if err != nil {
		return err
	}

	if g.JSON {
		devcli.PrintJSON(out)
		return nil
	}

	env := ui.DetectEnv("auto")
	painter := ui.NewPainter(env.Color)
	fmt.Fprintf(os.Stdout, "%s %s has %s table(s)\n", painter.Heading("Project:"), painter.Strong(out.ProjectID), painter.OK(fmt.Sprintf("%d", out.TableCount)))
	return nil
}
