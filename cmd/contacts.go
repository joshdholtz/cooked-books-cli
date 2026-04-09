package cmd

import (
	"fmt"
	"os"

	"github.com/joshdholtz/cooked-books-cli/api"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var contactsCmd = &cobra.Command{
	Use:   "contacts",
	Short: "List contacts",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return err
		}

		data, err := client.Get("/api/v1/contacts", nil)
		if err != nil {
			return err
		}

		contacts, ok := data["data"].([]any)
		if !ok || len(contacts) == 0 {
			fmt.Println("No contacts found.")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Email", "Type"})
		table.SetBorder(false)
		table.SetColumnSeparator("")
		table.SetHeaderLine(false)

		for _, c := range contacts {
			contact, ok := c.(map[string]any)
			if !ok {
				continue
			}

			name := getString(contact, "name")
			email := getString(contact, "email")
			if email == "" {
				email = "—"
			}
			typ := getString(contact, "type")

			table.Append([]string{name, email, typ})
		}

		table.Render()
		return nil
	},
}
