package cmd

import (
	"fmt"
	"os"

	"github.com/joshdholtz/cooked-books-cli/api"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var (
	txnStatus string
	txnSearch string
	txnLimit  string
)

var transactionsCmd = &cobra.Command{
	Use:     "transactions",
	Aliases: []string{"txn", "txns"},
	Short:   "List transactions",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return err
		}

		params := map[string]string{
			"status": txnStatus,
			"search": txnSearch,
			"limit":  txnLimit,
		}

		resp, err := client.Get("/api/v1/transactions", params)
		if err != nil {
			return err
		}

		transactions, ok := resp["data"].([]any)
		if !ok || len(transactions) == 0 {
			fmt.Println("No transactions found.")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Date", "Description", "Amount", "Status"})
		table.SetBorder(false)
		table.SetColumnSeparator("")
		table.SetHeaderLine(false)

		for _, t := range transactions {
			txn, ok := t.(map[string]any)
			if !ok {
				continue
			}

			date := getString(txn, "date")
			desc := getString(txn, "description")
			if len(desc) > 45 {
				desc = desc[:42] + "..."
			}
			amount := formatMoney(txn["amount"])
			status := getString(txn, "status")

			table.Append([]string{date, desc, amount, status})
		}

		table.Render()
		return nil
	},
}

func init() {
	transactionsCmd.Flags().StringVar(&txnStatus, "status", "", "Filter by status (new, suggested, reviewed, reconciled)")
	transactionsCmd.Flags().StringVar(&txnSearch, "search", "", "Search description")
	transactionsCmd.Flags().StringVar(&txnLimit, "limit", "20", "Number of results")
}
