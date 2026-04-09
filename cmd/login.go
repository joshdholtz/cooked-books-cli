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

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with your API token",
	Long: `Authenticate with your Cooked Books API token.

Generate a token at: https://app.cookedbooks.ai/integrations (Developer tab)

You can also set COOKED_BOOKS_TOKEN as an environment variable.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		bold := color.New(color.Bold)
		dim := color.New(color.FgHiBlack)

		bold.Println("Cooked Books CLI Login")
		fmt.Println()
		dim.Println("Generate a token at: https://app.cookedbooks.ai/integrations")
		dim.Println("Go to the Developer tab and click 'Generate API Token'")
		fmt.Println()

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("API Token: ")
		token, _ := reader.ReadString('\n')
		token = strings.TrimSpace(token)

		if token == "" {
			return fmt.Errorf("token cannot be empty")
		}

		baseURL := api.DefaultBaseURL()

		// Quick verify — try to hit the context endpoint
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
		fmt.Println("Logged in. Config saved to ~/.config/cooked-books/config.json")
		fmt.Println()
		dim.Println("Try: cooked-books context")

		return nil
	},
}
