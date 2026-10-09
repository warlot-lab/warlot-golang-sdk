package commands

import (
	"flag"
	"os"

	"github.com/warlot-lab/warlot-golang-sdk/warlot-go/internal/devcli"
	"github.com/warlot-lab/warlot-golang-sdk/warlot-go/internal/devcli/ui"
)

// RunStatus retrieves and presents project status and replication SLIs.
func RunStatus(args []string) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	projectID := fs.String("project", "", "Project ID (required)")
	g := devcli.ParseGlobalFlagsArgs(fs, args)
	if g.Err != nil {
		return g.Err
	}

	if err := devcli.RequireFlag(*projectID, "-project", "provide -project <id> to inspect status"); err != nil {
		return err
	}
	if err := devcli.EnsureAPIKey(&g); err != nil {
		return err
	}

	cl := devcli.NewClient(g)
	ctx, cancel := devcli.Ctx(g)
	defer cancel()

	out, err := cl.GetProjectStatus(ctx, *projectID)
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
	ui.RenderStatusCard(os.Stdout, painter, glyphs, out)
	return nil
}
