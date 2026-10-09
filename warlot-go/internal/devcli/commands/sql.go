package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/warlot-lab/warlot-golang-sdk/warlot-go/internal/devcli"
	"github.com/warlot-lab/warlot-golang-sdk/warlot-go/internal/devcli/ui"
	"github.com/warlot-lab/warlot-golang-sdk/warlot-go/warlot"
)

// RunSQL executes a SQL statement (optionally streaming rows or rendering tables).
func RunSQL(args []string) error {
	fs := flag.NewFlagSet("sql", flag.ContinueOnError)
	projectID := fs.String("project", "", "Project ID (required)")
	query := fs.String("q", "", "SQL query (required)")
	paramsJSON := fs.String("params", "", "Params JSON array, e.g. [\"Laptop\",999.99]")
	idempotency := fs.String("idempotency", "", "Idempotency key for writes")
	stream := fs.Bool("stream", false, "Stream SELECT rows as JSON")
	g := devcli.ParseGlobalFlagsArgs(fs, args)
	if g.Err != nil {
		return g.Err
	}

	if err := devcli.RequireFlag(*projectID, "-project", "provide -project <id> to execute SQL"); err != nil {
		return err
	}
	if err := devcli.RequireFlag(*query, "-q", "provide -q \"<SQL statement>\""); err != nil {
		return err
	}
	if err := devcli.EnsureAPIKey(&g); err != nil {
		return err
	}

	var params []any
	if strings.TrimSpace(*paramsJSON) != "" {
		if err := json.Unmarshal([]byte(*paramsJSON), &params); err != nil {
			return devcli.FlagErrorf("ensure -params is valid JSON array, e.g. '[\"val\", 123]'", "invalid -params JSON: %v", err)
		}
	}

	cl := devcli.NewClient(g)
	ctx, cancel := devcli.Ctx(g)
	defer cancel()

	callOpts := []warlot.CallOption{}
	if *idempotency != "" {
		callOpts = append(callOpts, warlot.WithIdempotencyKey(*idempotency))
	}

	if *stream {
		sc, err := cl.ExecSQLStream(ctx, *projectID, warlot.SQLRequest{SQL: *query, Params: params}, callOpts...)
		if err != nil {
			return err
		}
		defer sc.Close()

		var row map[string]any
		for sc.Next(&row) {
			devcli.PrintJSON(row)
			row = nil
		}
		if err := sc.Err(); err != nil {
			return fmt.Errorf("stream read error: %w", err)
		}
		return nil
	}

	proj := cl.Project(*projectID)
	res, err := proj.SQL(ctx, *query, params, callOpts...)
	if err != nil {
		return err
	}

	if g.JSON {
		devcli.PrintJSON(res)
		return nil
	}

	env := ui.DetectEnv("auto")
	painter := ui.NewPainter(env.Color)
	glyphs := ui.DefaultGlyphs(env.Unicode)

	if len(res.Rows) > 0 {
		if err := ui.RenderRowMaps(os.Stdout, painter, res.Rows); err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "\n%s %d row(s) returned\n", painter.Muted(glyphs.Dot()), len(res.Rows))
		return nil
	}

	rowCount := 0
	if res.RowCount != nil {
		rowCount = *res.RowCount
	}
	fmt.Fprintf(os.Stdout, "%s OK (%d row(s) affected)\n", painter.OK(glyphs.Check()), rowCount)
	return nil
}
