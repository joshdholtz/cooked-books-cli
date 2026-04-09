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

var (
	pnlStart string
	pnlEnd   string
)

var pnlCmd = &cobra.Command{
	Use:     "pnl",
	Aliases: []string{"profit-and-loss", "income"},
	Short:   "Profit & Loss report",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return err
		}

		if pnlStart == "" {
			pnlStart = fmt.Sprintf("%d-01-01", time.Now().Year())
		}
		if pnlEnd == "" {
			pnlEnd = time.Now().Format("2006-01-02")
		}

		params := map[string]string{
			"start_date": pnlStart,
			"end_date":   pnlEnd,
		}

		data, err := client.Get("/api/v1/reports/profit-and-loss", params)
		if err != nil {
			return err
		}

		bold := color.New(color.Bold)
		dim := color.New(color.FgHiBlack)

		bold.Printf("Profit & Loss: %s to %s\n", pnlStart, pnlEnd)
		dim.Println("══════════════════════════════════════")
		fmt.Println()

		report, _ := data["data"].(map[string]any)

		// Revenue
		printSection("Revenue", report, "revenue")
		printSection("Expenses", report, "expenses")

		fmt.Println()
		bold.Printf("  Net Income: %s\n", formatMoney(report["net_income"]))

		return nil
	},
}

func printSection(title string, report map[string]any, key string) {
	bold := color.New(color.Bold)
	bold.Println(title)

	items, ok := report[key].([]any)
	if !ok || len(items) == 0 {
		fmt.Println("  (none)")
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
		amount := formatMoney(row["total"])
		table.Append([]string{"  " + name, amount})
	}

	if total, ok := report["total_"+key]; ok {
		table.SetFooter([]string{"  Total " + title, formatMoney(total)})
	}

	table.Render()
	fmt.Println()
}

func init() {
	pnlCmd.Flags().StringVar(&pnlStart, "start", "", "Start date (YYYY-MM-DD, default: Jan 1)")
	pnlCmd.Flags().StringVar(&pnlEnd, "end", "", "End date (YYYY-MM-DD, default: today)")
}
