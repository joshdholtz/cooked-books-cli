package cmd

import (
	"fmt"
	"os"

	"github.com/joshdholtz/cooked-books-cli/api"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "List categorization rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return err
		}

		data, err := client.Get("/api/v1/rules", nil)
		if err != nil {
			return err
		}

		rules, ok := data["data"].([]any)
		if !ok || len(rules) == 0 {
			fmt.Println("No rules found.")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Pattern", "Account", "Priority"})
		table.SetBorder(false)
		table.SetColumnSeparator("")
		table.SetHeaderLine(false)

		for _, r := range rules {
			rule, ok := r.(map[string]any)
			if !ok {
				continue
			}

			pattern := getString(rule, "pattern")
			if len(pattern) > 40 {
				pattern = pattern[:37] + "..."
			}
			account := getString(rule, "account_name")
			priority := formatNum(rule["priority"])

			table.Append([]string{pattern, account, priority})
		}

		table.Render()
		return nil
	},
}
