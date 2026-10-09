package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/steven3002/warlot-golang-sdk/warlot-go/internal/devcli"
	"github.com/steven3002/warlot-golang-sdk/warlot-go/internal/devcli/ui"
	"github.com/steven3002/warlot-golang-sdk/warlot-go/warlot"
)

// RunResolve resolves a project by holder and project name.
func RunResolve(args []string) error {
	fs := flag.NewFlagSet("resolve", flag.ContinueOnError)
	g := devcli.ParseGlobalFlagsArgs(fs, args)

	if err := devcli.RequireFlag(g.HolderID, "-holder", "provide -holder <id> or set WARLOT_HOLDER"); err != nil {
		return err
	}
	if err := devcli.RequireFlag(g.ProjectName, "-pname", "provide -pname <name> or set WARLOT_PNAME"); err != nil {
		return err
	}

	cl := devcli.NewClient(g)
	ctx, cancel := devcli.Ctx(g)
	defer cancel()

	out, err := cl.ResolveProject(ctx, warlot.ResolveProjectRequest{
		HolderID:    g.HolderID,
		ProjectName: g.ProjectName,
	})
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

	fmt.Fprintf(os.Stdout, "%s Project Resolution:\n", painter.Heading("Status:"))
	fmt.Fprintf(os.Stdout, "  %s  %s\n", painter.Muted("Project ID:   "), painter.Strong(out.ProjectID))
	fmt.Fprintf(os.Stdout, "  %s  %s\n", painter.Muted("DB ID:        "), out.DBID)
	fmt.Fprintf(os.Stdout, "  %s  %s\n", painter.Muted("Meta Exists:  "), formatBool(out.ExistsMeta, painter, glyphs))
	fmt.Fprintf(os.Stdout, "  %s  %s\n", painter.Muted("Chain Exists: "), formatBool(out.ExistsChain, painter, glyphs))
	if out.Action != "" {
		fmt.Fprintf(os.Stdout, "  %s  %s\n", painter.Muted("Action:       "), painter.Hint(out.Action))
	}
	return nil
}

func formatBool(b bool, p ui.Painter, g ui.Glyphs) string {
	if b {
		return p.OK(fmt.Sprintf("%s yes", g.Check()))
	}
	return p.Muted(fmt.Sprintf("%s no", g.Cross()))
}
