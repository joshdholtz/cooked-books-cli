package cmd

import (
	"fmt"
	"os"

	"github.com/joshdholtz/cooked-books-cli/api"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var accountType string

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "List chart of accounts",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return err
		}

		params := map[string]string{
			"type": accountType,
		}

		data, err := client.Get("/api/v1/accounts", params)
		if err != nil {
			return err
		}

		accounts, ok := data["data"].([]any)
		if !ok || len(accounts) == 0 {
			fmt.Println("No accounts found.")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Code", "Name", "Type", "Sub Type"})
		table.SetBorder(false)
		table.SetColumnSeparator("")
		table.SetHeaderLine(false)

		for _, a := range accounts {
			acct, ok := a.(map[string]any)
			if !ok {
				continue
			}

			code := getString(acct, "code")
			name := getString(acct, "name")
			typ := getString(acct, "type")
			subType := getString(acct, "sub_type")
			if subType == "" {
				subType = "—"
			}

			table.Append([]string{code, name, typ, subType})
		}

		table.Render()
		return nil
	},
}

func init() {
	accountsCmd.Flags().StringVar(&accountType, "type", "", "Filter by type (asset, liability, equity, revenue, expense)")
}
