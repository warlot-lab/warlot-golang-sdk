package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/steven3002/warlot-golang-sdk/warlot-go/internal/devcli"
	"github.com/steven3002/warlot-golang-sdk/warlot-go/internal/devcli/ui"
	"github.com/steven3002/warlot-golang-sdk/warlot-go/warlot"
)

// RunInit initializes a new project.
func RunInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	owner := fs.String("owner", "", "Owner address (user)")
	includePass := fs.Bool("include-pass", true, "Include pass artifacts")
	deletable := fs.Bool("deletable", true, "Deletable project")
	g := devcli.ParseGlobalFlagsArgs(fs, args)
	if g.Err != nil {
		return g.Err
	}

	if err := devcli.RequireFlag(g.HolderID, "-holder", "provide -holder <id> or set WARLOT_HOLDER"); err != nil {
		return err
	}
	if err := devcli.RequireFlag(g.ProjectName, "-pname", "provide -pname <name> or set WARLOT_PNAME"); err != nil {
		return err
	}
	if err := devcli.RequireFlag(*owner, "-owner", "provide -owner <0xAddress>"); err != nil {
		return err
	}

	cl := devcli.NewClient(g)
	ctx, cancel := devcli.Ctx(g)
	defer cancel()

	out, err := cl.InitProject(ctx, warlot.InitProjectRequest{
		HolderID:      g.HolderID,
		ProjectName:   g.ProjectName,
		OwnerAddress:  *owner,
		EpochSet:      0,
		CycleEnd:      0,
		WritersLen:    0,
		TrackBackLen:  0,
		DraftEpochDur: 0,
		IncludePass:   *includePass,
		Deletable:     *deletable,
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

	fmt.Fprintf(os.Stdout, "%s Project initialized successfully!\n", painter.OK(fmt.Sprintf("[%s]", glyphs.Check())))
	fmt.Fprintf(os.Stdout, "  %s  %s\n", painter.Muted("Project ID:   "), painter.Strong(out.ProjectID))
	fmt.Fprintf(os.Stdout, "  %s  %s\n", painter.Muted("DB ID:        "), out.DBID)
	if out.WriterPassID != "" {
		fmt.Fprintf(os.Stdout, "  %s  %s\n", painter.Muted("Writer Pass:  "), out.WriterPassID)
	}
	if out.TxDigest != "" {
		fmt.Fprintf(os.Stdout, "  %s  %s\n", painter.Muted("Tx Digest:    "), out.TxDigest)
	}
	return nil
}
