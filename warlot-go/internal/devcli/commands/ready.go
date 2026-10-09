package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/steven3002/warlot-golang-sdk/warlot-go/internal/devcli"
	"github.com/steven3002/warlot-golang-sdk/warlot-go/internal/devcli/ui"
)

// RunReady queries the cluster readiness probe (/readyz) to diagnose backend dependencies.
func RunReady(args []string) error {
	fs := flag.NewFlagSet("ready", flag.ContinueOnError)
	g := devcli.ParseGlobalFlagsArgs(fs, args)

	cl := devcli.NewClient(g)
	ctx, cancel := devcli.Ctx(g)
	defer cancel()

	res, err := cl.CheckReadiness(ctx)
	if res != nil && g.JSON {
		devcli.PrintJSON(res)
		if err != nil {
			return err
		}
		return nil
	}

	if res != nil {
		env := ui.DetectEnv("auto")
		painter := ui.NewPainter(env.Color)
		glyphs := ui.DefaultGlyphs(env.Unicode)
		ui.RenderReadinessCard(os.Stdout, painter, glyphs, res)
		if !res.IsReady() {
			return fmt.Errorf("one or more dependencies are degraded")
		}
		return nil
	}

	return err
}
