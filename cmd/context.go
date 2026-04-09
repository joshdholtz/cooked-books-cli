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

		resp, err := client.Get("/api/v1/copilot/context", nil)
		if err != nil {
			return err
		}

		data, _ := resp["data"].(map[string]any)
		if data == nil {
			return fmt.Errorf("unexpected response format")
		}

		bold := color.New(color.Bold)
		heading := color.New(color.FgYellow, color.Bold)
		label := color.New(color.FgHiBlack)
		value := color.New(color.FgGreen)
		dim := color.New(color.FgHiBlack)

		// Org name
		orgName := ""
		if org, ok := data["organization"].(map[string]any); ok {
			orgName = getString(org, "name")
		}
		heading.Printf("%s — Financial Overview\n", orgName)
		dim.Println("══════════════════════════════════════")
		fmt.Println()

		// Transaction counts
		if summary, ok := data["summary"].(map[string]any); ok {
			if txns, ok := summary["transactions"].(map[string]any); ok {
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
		}

		// Accounts overview
		if accts, ok := data["accounts_overview"].(map[string]any); ok {
			bold.Println("Accounts")
			label.Print("  Cash balance:  ")
			value.Println(formatMoney(accts["cash_balance"]))
			label.Print("  Revenue YTD:   ")
			value.Println(formatMoney(accts["total_revenue_ytd"]))
			label.Print("  Expenses YTD:  ")
			value.Println(formatMoney(accts["total_expenses_ytd"]))
			label.Print("  Net income:    ")
			value.Println(formatMoney(accts["net_income_ytd"]))
			fmt.Println()
		}

		// Available actions
		if actions, ok := data["available_actions"].([]any); ok && len(actions) > 0 {
			bold.Println("Action Items")
			for _, a := range actions {
				action, ok := a.(map[string]any)
				if !ok {
					continue
				}
				priority := getString(action, "priority")
				desc := getString(action, "description")
				marker := "  ·"
				if priority == "high" {
					marker = "  !"
				}
				label.Print(marker + " ")
				fmt.Println(desc)
			}
			fmt.Println()
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
		if n == "" {
			return "$0.00"
		}
		return "$" + n
	default:
		return fmt.Sprintf("$%v", v)
	}
}
