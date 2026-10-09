package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/warlot-lab/warlot-golang-sdk/warlot-go/internal/devcli"
	"github.com/warlot-lab/warlot-golang-sdk/warlot-go/internal/devcli/ui"
)

// RunCommit commits project changes to chain-backed storage.
func RunCommit(args []string) error {
	fs := flag.NewFlagSet("commit", flag.ContinueOnError)
	projectID := fs.String("project", "", "Project ID (required)")
	g := devcli.ParseGlobalFlagsArgs(fs, args)
	if g.Err != nil {
		return g.Err
	}

	if err := devcli.RequireFlag(*projectID, "-project", "provide -project <id> to trigger commit"); err != nil {
		return err
	}
	if err := devcli.EnsureAPIKey(&g); err != nil {
		return err
	}

	cl := devcli.NewClient(g)
	ctx, cancel := devcli.Ctx(g)
	defer cancel()

	out, err := cl.CommitProject(ctx, *projectID)
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

	fmt.Fprintf(os.Stdout, "%s Committed %d mutation(s) on-chain\n", painter.OK(fmt.Sprintf("[%s]", glyphs.Check())), out.CommittedOps)
	if out.TxDigest != "" {
		fmt.Fprintf(os.Stdout, "  %s  %s\n", painter.Muted("Tx Digest:    "), painter.Strong(out.TxDigest))
	}
	fmt.Fprintf(os.Stdout, "  %s  %d (Current Max: %d)\n", painter.Muted("Anchored Seq: "), out.LastUploadSeq, out.MaxSeq)
	return nil
}
