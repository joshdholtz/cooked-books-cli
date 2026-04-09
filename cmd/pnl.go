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

		resp, err := client.Get("/api/v1/reports/profit-and-loss", params)
		if err != nil {
			return err
		}

		data, _ := resp["data"].(map[string]any)
		if data == nil {
			return fmt.Errorf("unexpected response format")
		}

		bold := color.New(color.Bold)
		dim := color.New(color.FgHiBlack)

		bold.Printf("Profit & Loss: %s to %s\n", pnlStart, pnlEnd)
		dim.Println("══════════════════════════════════════")
		fmt.Println()

		rows, _ := data["rows"].([]any)

		// Split into revenue and expenses
		var revenue, expenses []map[string]any
		var totalRevenue, totalExpenses float64

		for _, r := range rows {
			row, ok := r.(map[string]any)
			if !ok {
				continue
			}
			typ := getString(row, "type")
			credits := parseFloat(getString(row, "total_credits"))
			debits := parseFloat(getString(row, "total_debits"))

			if typ == "revenue" {
				revenue = append(revenue, row)
				totalRevenue += credits - debits
			} else if typ == "expense" {
				expenses = append(expenses, row)
				totalExpenses += debits - credits
			}
		}

		printPnlSection("Revenue", revenue, totalRevenue, true)
		printPnlSection("Expenses", expenses, totalExpenses, false)

		fmt.Println()
		net := totalRevenue - totalExpenses
		bold.Printf("  Net Income: $%.2f\n", net)

		return nil
	},
}

func printPnlSection(title string, rows []map[string]any, total float64, isRevenue bool) {
	bold := color.New(color.Bold)
	bold.Println(title)

	if len(rows) == 0 {
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

	for _, row := range rows {
		name := getString(row, "name")
		var amount float64
		if isRevenue {
			amount = parseFloat(getString(row, "total_credits")) - parseFloat(getString(row, "total_debits"))
		} else {
			amount = parseFloat(getString(row, "total_debits")) - parseFloat(getString(row, "total_credits"))
		}
		if amount == 0 {
			continue
		}
		table.Append([]string{"  " + name, fmt.Sprintf("$%.2f", amount)})
	}

	table.SetFooter([]string{"  Total " + title, fmt.Sprintf("$%.2f", total)})
	table.Render()
	fmt.Println()
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

func init() {
	pnlCmd.Flags().StringVar(&pnlStart, "start", "", "Start date (YYYY-MM-DD, default: Jan 1)")
	pnlCmd.Flags().StringVar(&pnlEnd, "end", "", "End date (YYYY-MM-DD, default: today)")
}
