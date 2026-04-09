package cmd

import (
	"fmt"
	"os"

	"github.com/joshdholtz/cooked-books-cli/api"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var invoiceStatus string

var invoicesCmd = &cobra.Command{
	Use:   "invoices",
	Short: "List invoices",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return err
		}

		params := map[string]string{
			"status": invoiceStatus,
		}

		data, err := client.Get("/api/v1/invoices", params)
		if err != nil {
			return err
		}

		invoices, ok := data["data"].([]any)
		if !ok || len(invoices) == 0 {
			fmt.Println("No invoices found.")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Number", "Contact", "Total", "Status", "Due Date"})
		table.SetBorder(false)
		table.SetColumnSeparator("")
		table.SetHeaderLine(false)

		for _, inv := range invoices {
			i, ok := inv.(map[string]any)
			if !ok {
				continue
			}

			number := getString(i, "number")
			contact := getString(i, "contact_name")
			if contact == "" {
				contact = "—"
			}
			total := formatMoney(i["total"])
			status := getString(i, "status")
			due := getString(i, "due_date")
			if due == "" {
				due = "—"
			}

			table.Append([]string{number, contact, total, status, due})
		}

		table.Render()
		return nil
	},
}

func init() {
	invoicesCmd.Flags().StringVar(&invoiceStatus, "status", "", "Filter by status (draft, sent, paid, overdue, void)")
}
