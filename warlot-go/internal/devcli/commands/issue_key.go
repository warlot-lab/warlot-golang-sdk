package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/steven3002/warlot-golang-sdk/warlot-go/internal/devcli"
	"github.com/steven3002/warlot-golang-sdk/warlot-go/internal/devcli/ui"
	"github.com/steven3002/warlot-golang-sdk/warlot-go/warlot"
)

// RunIssueKey issues an API key for a project.
func RunIssueKey(args []string) error {
	fs := flag.NewFlagSet("issue-key", flag.ContinueOnError)
	projectID := fs.String("project", "", "Project ID (required)")
	userAddr := fs.String("user", "", "User address (owner) (required)")
	g := devcli.ParseGlobalFlagsArgs(fs, args)
	if g.Err != nil {
		return g.Err
	}

	if err := devcli.RequireFlag(*projectID, "-project", "provide -project <id> to issue an API key"); err != nil {
		return err
	}
	if err := devcli.RequireFlag(*userAddr, "-user", "provide -user <0xAddress>"); err != nil {
		return err
	}

	cl := devcli.NewClient(g)
	ctx, cancel := devcli.Ctx(g)
	defer cancel()

	out, err := cl.IssueAPIKey(ctx, warlot.IssueKeyRequest{
		ProjectID:     *projectID,
		ProjectHolder: g.HolderID,
		ProjectName:   g.ProjectName,
		User:          *userAddr,
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

	fmt.Fprintf(os.Stdout, "%s API Key issued successfully!\n", painter.OK(fmt.Sprintf("[%s]", glyphs.Check())))
	fmt.Fprintf(os.Stdout, "  %s  %s\n", painter.Muted("API Key: "), painter.Strong(out.APIKey))
	if out.URL != "" {
		fmt.Fprintf(os.Stdout, "  %s  %s\n", painter.Muted("URL:     "), out.URL)
	}
	return nil
}
