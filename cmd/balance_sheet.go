package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/joshdholtz/cooked-books-cli/api"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var bsDate string

var balanceSheetCmd = &cobra.Command{
	Use:     "balance-sheet",
	Aliases: []string{"bs"},
	Short:   "Balance Sheet report",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return err
		}

		if bsDate == "" {
			bsDate = time.Now().Format("2006-01-02")
		}

		params := map[string]string{
			"as_of_date": bsDate,
		}

		data, err := client.Get("/api/v1/reports/balance-sheet", params)
		if err != nil {
			return err
		}

		bold := color.New(color.Bold)
		dim := color.New(color.FgHiBlack)

		bold.Printf("Balance Sheet as of %s\n", bsDate)
		dim.Println("══════════════════════════════════════")
		fmt.Println()

		report, _ := data["data"].(map[string]any)

		printBSSection("Assets", report, "assets", "total_assets")
		printBSSection("Liabilities", report, "liabilities", "total_liabilities")
		printBSSection("Equity", report, "equity", "total_equity")

		return nil
	},
}

func printBSSection(title string, report map[string]any, key string, totalKey string) {
	bold := color.New(color.Bold)
	bold.Println(title)

	items, ok := report[key].([]any)
	if !ok || len(items) == 0 {
		fmt.Println("  (none)")
		fmt.Println()
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetColumnSeparator("")
	table.SetHeaderLine(false)
	table.SetAutoWrapText(false)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_RIGHT})

	for _, item := range items {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name := getString(row, "account_name")
		balance := formatMoney(row["balance"])
		table.Append([]string{"  " + name, balance})
	}

	if total, ok := report[totalKey]; ok {
		table.SetFooter([]string{"  Total " + title, formatMoney(total)})
	}

	table.Render()
	fmt.Println()
}

func init() {
	balanceSheetCmd.Flags().StringVar(&bsDate, "date", "", "As-of date (YYYY-MM-DD, default: today)")
}
