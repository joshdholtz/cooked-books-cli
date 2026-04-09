package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cooked-books",
	Short: "CLI for Cooked Books — AI-powered bookkeeping",
	Long: `Cooked Books CLI

A double-entry accounting system that works for you.
Talk to your books from the terminal.

Get started:
  cooked-books auth set-token <token>   Save your API token
  cooked-books context                  Financial overview
  cooked-books transactions             List transactions
  cooked-books pnl                      Profit & Loss report

Or set COOKED_BOOKS_TOKEN and skip auth entirely.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(contextCmd)
	rootCmd.AddCommand(transactionsCmd)
	rootCmd.AddCommand(accountsCmd)
	rootCmd.AddCommand(pnlCmd)
	rootCmd.AddCommand(balanceSheetCmd)
	rootCmd.AddCommand(invoicesCmd)
	rootCmd.AddCommand(contactsCmd)
	rootCmd.AddCommand(rulesCmd)
}
