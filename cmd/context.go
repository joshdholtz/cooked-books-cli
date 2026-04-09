package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/joshdholtz/cooked-books-cli/api"
	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Financial overview — start here",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return err
		}

		data, err := client.Get("/api/v1/copilot/context", nil)
		if err != nil {
			return err
		}

		bold := color.New(color.Bold)
		heading := color.New(color.FgYellow, color.Bold)
		label := color.New(color.FgHiBlack)
		value := color.New(color.FgGreen)
		dim := color.New(color.FgHiBlack)

		orgName := getString(data, "organization_name")
		heading.Printf("%s — Financial Overview\n", orgName)
		dim.Println("══════════════════════════════════════")
		fmt.Println()

		// Transaction counts
		if txns, ok := data["transaction_counts"].(map[string]any); ok {
			bold.Println("Transactions")
			label.Print("  To review:    ")
			value.Println(formatNum(txns["new"]))
			label.Print("  AI suggested: ")
			value.Println(formatNum(txns["suggested"]))
			label.Print("  Reviewed:     ")
			value.Println(formatNum(txns["reviewed"]))
			label.Print("  Reconciled:   ")
			value.Println(formatNum(txns["reconciled"]))
			fmt.Println()
		}

		// P&L summary
		if pnl, ok := data["profit_and_loss"].(map[string]any); ok {
			bold.Println("Profit & Loss (YTD)")
			label.Print("  Revenue:  ")
			value.Println(formatMoney(pnl["total_revenue"]))
			label.Print("  Expenses: ")
			value.Println(formatMoney(pnl["total_expenses"]))
			label.Print("  Net:      ")
			net := formatMoney(pnl["net_income"])
			if getString(pnl, "net_income") != "" {
				value.Println(net)
			} else {
				fmt.Println(net)
			}
			fmt.Println()
		}

		// Cash balance
		if bs, ok := data["balance_sheet"].(map[string]any); ok {
			bold.Println("Balance Sheet")
			label.Print("  Assets:      ")
			value.Println(formatMoney(bs["total_assets"]))
			label.Print("  Liabilities: ")
			value.Println(formatMoney(bs["total_liabilities"]))
			label.Print("  Equity:      ")
			value.Println(formatMoney(bs["total_equity"]))
			fmt.Println()
		}

		// Pending rules
		if rules, ok := data["pending_rules"].(float64); ok && rules > 0 {
			label.Printf("  %d rule proposals ready to review\n", int(rules))
		}

		return nil
	},
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func formatNum(v any) string {
	if v == nil {
		return "0"
	}
	switch n := v.(type) {
	case float64:
		return fmt.Sprintf("%d", int(n))
	default:
		return fmt.Sprintf("%v", v)
	}
}

func formatMoney(v any) string {
	if v == nil {
		return "$0.00"
	}
	switch n := v.(type) {
	case float64:
		return fmt.Sprintf("$%.2f", n)
	case string:
		return "$" + n
	default:
		return fmt.Sprintf("$%v", v)
	}
}
