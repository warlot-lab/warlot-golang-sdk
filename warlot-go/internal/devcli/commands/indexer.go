package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/steven3002/warlot-golang-sdk/warlot-go/internal/devcli"
	"github.com/steven3002/warlot-golang-sdk/warlot-go/internal/devcli/ui"
	"github.com/steven3002/warlot-golang-sdk/warlot-go/warlot"
)

// RunIndexer handles CLI queries for Warlot Indexer read projections.
func RunIndexer(args []string) error {
	if len(args) == 0 {
		return devcli.FlagErrorf("run 'warlotdev indexer <health|storage|balances|history|holders|projects> [flags]'", "subcommand is required")
	}

	sub := args[0]
	subArgs := args[1:]

	fs := flag.NewFlagSet("indexer "+sub, flag.ContinueOnError)
	indexerURL := fs.String("indexer", "http://10.43.102.80:8081", "Indexer base URL")
	addr := fs.String("address", "", "Target Sui address (0x...)")
	limit := fs.Int("limit", 10, "History item limit")
	cursor := fs.String("cursor", "", "Pagination cursor")
	g := devcli.ParseGlobalFlagsArgs(fs, subArgs)
	if g.Err != nil {
		return g.Err
	}

	idx := warlot.NewIndexerClient(*indexerURL, nil)
	ctx, cancel := devcli.Ctx(g)
	defer cancel()

	env := ui.DetectEnv("auto")
	painter := ui.NewPainter(env.Color)
	glyphs := ui.DefaultGlyphs(env.Unicode)

	switch sub {
	case "health":
		h, err := idx.CheckHealth(ctx)
		if err != nil {
			return err
		}
		if g.JSON {
			devcli.PrintJSON(h)
			return nil
		}
		fmt.Fprintf(os.Stdout, "%s %s\n", painter.Heading("Indexer Health:"), painter.OK(fmt.Sprintf("[%s OK]", glyphs.Check())))
		fmt.Fprintf(os.Stdout, "  %s  %d\n", painter.Muted("Checkpoint: "), h.Freshness.Checkpoint)
		fmt.Fprintf(os.Stdout, "  %s  %s\n", painter.Muted("Tx Digest:  "), h.Freshness.TxDigest)
		fmt.Fprintf(os.Stdout, "  %s  %s\n", painter.Muted("Updated At: "), h.Freshness.UpdatedAt.Format("2006-01-02 15:04:05 UTC"))
		return nil

	case "storage":
		if err := devcli.RequireFlag(*addr, "-address", "provide -address <0xAddress>"); err != nil {
			return err
		}
		st, err := idx.GetAddressStorage(ctx, *addr)
		if err != nil {
			return err
		}
		if g.JSON {
			devcli.PrintJSON(st)
			return nil
		}
		fmt.Fprintf(os.Stdout, "%s %s\n", painter.Heading("Storage Projections:"), painter.Strong(*addr))
		fmt.Fprintf(os.Stdout, "  %s  %d bytes\n", painter.Muted("Active Bytes:"), st.Data.StorageBytes)
		fmt.Fprintf(os.Stdout, "  %s  %d\n", painter.Muted("File Count:  "), st.Data.FileCount)
		fmt.Fprintf(os.Stdout, "  %s  %d\n", painter.Muted("Checkpoint:  "), st.Freshness.Checkpoint)
		return nil

	case "balances":
		if err := devcli.RequireFlag(*addr, "-address", "provide -address <0xAddress>"); err != nil {
			return err
		}
		bals, err := idx.GetAddressBalances(ctx, *addr)
		if err != nil {
			return err
		}
		if g.JSON {
			devcli.PrintJSON(bals)
			return nil
		}
		if len(bals.Data) == 0 {
			fmt.Fprintln(os.Stdout, painter.Muted("(no wallet balances found)"))
			return nil
		}
		fmt.Fprintf(os.Stdout, "%s %s\n\n", painter.Heading("Balances:"), painter.Strong(*addr))
		tbl := ui.NewTable(os.Stdout, painter)
		tbl.SetHeaders("Coin Type", "Balance", "Observed Tx")
		for _, b := range bals.Data {
			tbl.AddRow(b.CoinType, fmt.Sprintf("%d", b.Balance), b.ObservedTx)
		}
		return tbl.Flush()

	case "history":
		if err := devcli.RequireFlag(*addr, "-address", "provide -address <0xAddress>"); err != nil {
			return err
		}
		hist, err := idx.GetAddressHistory(ctx, *addr, *cursor, *limit)
		if err != nil {
			return err
		}
		if g.JSON {
			devcli.PrintJSON(hist)
			return nil
		}
		if len(hist.Data) == 0 {
			fmt.Fprintln(os.Stdout, painter.Muted("(no history entries found)"))
			return nil
		}
		fmt.Fprintf(os.Stdout, "%s %s (%d events)\n\n", painter.Heading("Activity History:"), painter.Strong(*addr), len(hist.Data))
		tbl := ui.NewTable(os.Stdout, painter)
		tbl.SetHeaders("Kind", "Amount", "Coin Type", "Direction", "Tx Digest")
		for _, item := range hist.Data {
			tbl.AddRow(item.Kind, fmt.Sprintf("%d", item.Amount), item.CoinType, item.Direction, item.TxDigest)
		}
		return tbl.Flush()

	case "holders":
		if err := devcli.RequireFlag(*addr, "-address", "provide -address <0xAddress>"); err != nil {
			return err
		}
		holders, err := idx.GetHoldersByAdmin(ctx, *addr)
		if err != nil {
			return err
		}
		if g.JSON {
			devcli.PrintJSON(holders)
			return nil
		}
		if len(holders.Data) == 0 {
			fmt.Fprintln(os.Stdout, painter.Muted("(no holders administered by this address)"))
			return nil
		}
		fmt.Fprintf(os.Stdout, "%s %s\n\n", painter.Heading("Administered Holders:"), painter.Strong(*addr))
		tbl := ui.NewTable(os.Stdout, painter)
		tbl.SetHeaders("Holder ID", "Admin Address", "Created Tx")
		for _, h := range holders.Data {
			tbl.AddRow(h.HolderID, h.AdminAddr, h.CreatedTx)
		}
		return tbl.Flush()

	case "projects":
		targetHolder := g.HolderID
		if targetHolder == "" && *addr != "" {
			targetHolder = *addr
		}
		if err := devcli.RequireFlag(targetHolder, "-holder", "provide -holder <0xHolderID> or -address <0xHolderID>"); err != nil {
			return err
		}
		proj, err := idx.GetHolderProjects(ctx, targetHolder)
		if err != nil {
			return err
		}
		if g.JSON {
			devcli.PrintJSON(proj)
			return nil
		}
		fmt.Fprintf(os.Stdout, "%s %s\n", painter.Heading("Holder Projects:"), painter.Strong(targetHolder))
		fmt.Fprintf(os.Stdout, "  %s  %d\n", painter.Muted("Active Projects:"), proj.Data.ActiveProjects)
		return nil

	default:
		return devcli.FlagErrorf("valid subcommands are health, storage, balances, history, holders, projects", "unknown indexer subcommand %q", sub)
	}
}
