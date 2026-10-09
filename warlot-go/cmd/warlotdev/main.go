package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/warlot-lab/warlot-golang-sdk/warlot-go/internal/devcli"
	"github.com/warlot-lab/warlot-golang-sdk/warlot-go/internal/devcli/commands"
	"github.com/warlot-lab/warlot-golang-sdk/warlot-go/internal/devcli/ui"
)

// Entry point for the official CLI: warlotdev.
func main() {
	if len(os.Args) < 2 {
		devcli.PrintGlobalUsage("warlotdev")
		os.Exit(2)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "help", "-h", "--help":
		devcli.PrintGlobalUsage("warlotdev")
		return

	case "resolve":
		if err := commands.RunResolve(args); err != nil {
			fail(err)
		}
	case "init":
		if err := commands.RunInit(args); err != nil {
			fail(err)
		}
	case "issue-key":
		if err := commands.RunIssueKey(args); err != nil {
			fail(err)
		}
	case "sql":
		if err := commands.RunSQL(args); err != nil {
			fail(err)
		}
	case "tables":
		if err := commands.RunTables(args); err != nil {
			fail(err)
		}
	case "status":
		if err := commands.RunStatus(args); err != nil {
			fail(err)
		}
	case "commit":
		if err := commands.RunCommit(args); err != nil {
			fail(err)
		}
	case "ready":
		if err := commands.RunReady(args); err != nil {
			fail(err)
		}
	case "indexer":
		if err := commands.RunIndexer(args); err != nil {
			fail(err)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		devcli.PrintGlobalUsage("warlotdev")
		os.Exit(2)
	}
}

func fail(err error) {
	env := ui.DetectEnv("auto")
	painter := ui.NewPainter(env.Color)

	var uErr *devcli.UsageError
	if errors.As(err, &uErr) {
		fmt.Fprint(os.Stderr, painter.FormatError(uErr.Err, uErr.Hint))
		os.Exit(2)
	}

	fmt.Fprint(os.Stderr, painter.FormatError(err, ""))
	os.Exit(1)
}
