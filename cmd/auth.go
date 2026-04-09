package cmd

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/joshdholtz/cooked-books-cli/api"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
}

var authSetTokenCmd = &cobra.Command{
	Use:   "set-token [token]",
	Short: "Save an API token to ~/.config/cooked-books/",
	Long: `Save an API token for future CLI use.

Generate a token at: https://app.cookedbooks.ai/integrations (Developer tab)

Pass the token as an argument or omit it for an interactive prompt.
You can also skip this entirely and set COOKED_BOOKS_TOKEN instead.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var token string

		if len(args) > 0 {
			token = args[0]
		} else {
			dim := color.New(color.FgHiBlack)
			dim.Println("Generate a token at: https://app.cookedbooks.ai/integrations")
			fmt.Println()

			reader := bufio.NewReader(os.Stdin)
			fmt.Print("API Token: ")
			t, _ := reader.ReadString('\n')
			token = strings.TrimSpace(t)
		}

		if token == "" {
			return fmt.Errorf("token cannot be empty")
		}

		baseURL := api.DefaultBaseURL()

		// Verify the token works
		client := &api.Client{
			BaseURL: baseURL,
			Token:   token,
			HTTP:    &http.Client{},
		}

		_, err := client.Get("/api/v1/copilot/context", nil)
		if err != nil {
			return fmt.Errorf("token verification failed: %w", err)
		}

		cfg := &api.Config{
			Token:   token,
			BaseURL: baseURL,
		}

		if err := api.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		green := color.New(color.FgGreen, color.Bold)
		green.Print("✓ ")
		fmt.Println("Token saved to ~/.config/cooked-books/config.json")

		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			fmt.Println("Not authenticated.")
			dim := color.New(color.FgHiBlack)
			dim.Println("Run: cooked-books auth set-token")
			dim.Println("Or set COOKED_BOOKS_TOKEN environment variable")
			return nil
		}

		_, err = client.Get("/api/v1/copilot/context", nil)
		if err != nil {
			color.New(color.FgRed, color.Bold).Print("✗ ")
			fmt.Printf("Token is invalid: %v\n", err)
			return nil
		}

		green := color.New(color.FgGreen, color.Bold)
		green.Print("✓ ")
		fmt.Printf("Authenticated to %s\n", client.BaseURL)
		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove saved token",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := api.SaveConfig(&api.Config{}); err != nil {
			return err
		}
		fmt.Println("Token removed.")
		return nil
	},
}

func init() {
	authCmd.AddCommand(authSetTokenCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)
}
