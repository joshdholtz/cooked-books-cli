package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/joshdholtz/cooked-books-cli/api"
)

// mockServer creates a test server that returns canned JSON for different paths
func mockServer(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.URL.Path == "/api/v1/copilot/context":
			json.NewEncoder(w).Encode(map[string]any{
				"organization_name": "Test Org",
				"transaction_counts": map[string]any{
					"new":        float64(5),
					"suggested":  float64(2),
					"reviewed":   float64(100),
					"reconciled": float64(500),
				},
				"profit_and_loss": map[string]any{
					"total_revenue":  "10000.00",
					"total_expenses": "7500.00",
					"net_income":     "2500.00",
				},
				"balance_sheet": map[string]any{
					"total_assets":      "25000.00",
					"total_liabilities": "5000.00",
					"total_equity":      "20000.00",
				},
			})

		case r.URL.Path == "/api/v1/transactions":
			json.NewEncoder(w).Encode(map[string]any{
				"data": []any{
					map[string]any{
						"date":                  "2025-06-15",
						"description":           "STRIPE TRANSFER",
						"amount":                float64(2450.00),
						"category_account_name": "Sales Revenue",
						"status":                "reconciled",
					},
					map[string]any{
						"date":                  "2025-06-14",
						"description":           "AMAZON WEB SERVICES",
						"amount":                float64(-127.50),
						"category_account_name": "",
						"status":                "new",
					},
				},
			})

		case r.URL.Path == "/api/v1/accounts":
			json.NewEncoder(w).Encode(map[string]any{
				"data": []any{
					map[string]any{"code": "1000", "name": "Cash", "type": "asset", "sub_type": "current"},
					map[string]any{"code": "4000", "name": "Sales Revenue", "type": "revenue", "sub_type": ""},
				},
			})

		case r.URL.Path == "/api/v1/reports/profit-and-loss":
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"revenue": []any{
						map[string]any{"account_name": "Sales Revenue", "total": "10000.00"},
					},
					"expenses": []any{
						map[string]any{"account_name": "Hosting", "total": "500.00"},
					},
					"total_revenue":  "10000.00",
					"total_expenses": "500.00",
					"net_income":     "9500.00",
				},
			})

		case r.URL.Path == "/api/v1/reports/balance-sheet":
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"assets":            []any{map[string]any{"account_name": "Cash", "balance": "25000.00"}},
					"liabilities":       []any{},
					"equity":            []any{map[string]any{"account_name": "Retained Earnings", "balance": "25000.00"}},
					"total_assets":      "25000.00",
					"total_liabilities": "0.00",
					"total_equity":      "25000.00",
				},
			})

		case r.URL.Path == "/api/v1/invoices":
			json.NewEncoder(w).Encode(map[string]any{
				"data": []any{
					map[string]any{
						"number":       "INV-001",
						"contact_name": "Acme Corp",
						"total":        float64(5000.00),
						"status":       "sent",
						"due_date":     "2025-07-01",
					},
				},
			})

		case r.URL.Path == "/api/v1/contacts":
			json.NewEncoder(w).Encode(map[string]any{
				"data": []any{
					map[string]any{"name": "Acme Corp", "email": "billing@acme.com", "type": "customer"},
				},
			})

		case r.URL.Path == "/api/v1/rules":
			json.NewEncoder(w).Encode(map[string]any{
				"data": []any{
					map[string]any{"pattern": "STRIPE", "account_name": "Sales Revenue", "priority": float64(10)},
				},
			})

		default:
			w.WriteHeader(404)
			json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
		}
	}))
}

// setupTestClient sets COOKED_BOOKS_TOKEN and COOKED_BOOKS_API_URL for a test
func setupTestClient(t *testing.T, serverURL string) {
	t.Helper()
	t.Setenv("COOKED_BOOKS_TOKEN", "test-token")
	t.Setenv("COOKED_BOOKS_API_URL", serverURL)
}

// captureOutput runs a cobra command and captures stdout
func captureOutput(t *testing.T, args []string) string {
	t.Helper()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	rootCmd.Execute()
	return buf.String()
}

func TestContextCommand(t *testing.T) {
	server := mockServer(t)
	defer server.Close()
	setupTestClient(t, server.URL)

	// Run context command directly
	client, _ := api.NewClient()
	data, err := client.Get("/api/v1/copilot/context", nil)
	if err != nil {
		t.Fatalf("context request failed: %v", err)
	}

	if getString(data, "organization_name") != "Test Org" {
		t.Errorf("expected org name 'Test Org', got '%s'", getString(data, "organization_name"))
	}

	txns, ok := data["transaction_counts"].(map[string]any)
	if !ok {
		t.Fatal("expected transaction_counts map")
	}
	if txns["new"] != float64(5) {
		t.Errorf("expected 5 new transactions, got %v", txns["new"])
	}
}

func TestTransactionsCommand(t *testing.T) {
	server := mockServer(t)
	defer server.Close()
	setupTestClient(t, server.URL)

	client, _ := api.NewClient()
	data, err := client.Get("/api/v1/transactions", map[string]string{"status": "new"})
	if err != nil {
		t.Fatalf("transactions request failed: %v", err)
	}

	txns, ok := data["data"].([]any)
	if !ok {
		t.Fatal("expected data array")
	}
	if len(txns) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(txns))
	}

	first := txns[0].(map[string]any)
	if first["description"] != "STRIPE TRANSFER" {
		t.Errorf("expected 'STRIPE TRANSFER', got '%s'", first["description"])
	}
}

func TestAccountsCommand(t *testing.T) {
	server := mockServer(t)
	defer server.Close()
	setupTestClient(t, server.URL)

	client, _ := api.NewClient()
	data, err := client.Get("/api/v1/accounts", map[string]string{"type": "asset"})
	if err != nil {
		t.Fatalf("accounts request failed: %v", err)
	}

	accounts, ok := data["data"].([]any)
	if !ok {
		t.Fatal("expected data array")
	}
	if len(accounts) != 2 {
		t.Errorf("expected 2 accounts, got %d", len(accounts))
	}
}

func TestPnlCommand(t *testing.T) {
	server := mockServer(t)
	defer server.Close()
	setupTestClient(t, server.URL)

	client, _ := api.NewClient()
	data, err := client.Get("/api/v1/reports/profit-and-loss", map[string]string{
		"start_date": "2025-01-01",
		"end_date":   "2025-12-31",
	})
	if err != nil {
		t.Fatalf("pnl request failed: %v", err)
	}

	report, ok := data["data"].(map[string]any)
	if !ok {
		t.Fatal("expected data map")
	}
	if report["net_income"] != "9500.00" {
		t.Errorf("expected net income 9500.00, got %v", report["net_income"])
	}
}

func TestBalanceSheetCommand(t *testing.T) {
	server := mockServer(t)
	defer server.Close()
	setupTestClient(t, server.URL)

	client, _ := api.NewClient()
	data, err := client.Get("/api/v1/reports/balance-sheet", map[string]string{
		"as_of_date": "2025-12-31",
	})
	if err != nil {
		t.Fatalf("balance sheet request failed: %v", err)
	}

	report, ok := data["data"].(map[string]any)
	if !ok {
		t.Fatal("expected data map")
	}
	if report["total_assets"] != "25000.00" {
		t.Errorf("expected total assets 25000.00, got %v", report["total_assets"])
	}
}

func TestInvoicesCommand(t *testing.T) {
	server := mockServer(t)
	defer server.Close()
	setupTestClient(t, server.URL)

	client, _ := api.NewClient()
	data, err := client.Get("/api/v1/invoices", nil)
	if err != nil {
		t.Fatalf("invoices request failed: %v", err)
	}

	invoices := data["data"].([]any)
	if len(invoices) != 1 {
		t.Errorf("expected 1 invoice, got %d", len(invoices))
	}

	inv := invoices[0].(map[string]any)
	if inv["number"] != "INV-001" {
		t.Errorf("expected INV-001, got %v", inv["number"])
	}
}

func TestContactsCommand(t *testing.T) {
	server := mockServer(t)
	defer server.Close()
	setupTestClient(t, server.URL)

	client, _ := api.NewClient()
	data, err := client.Get("/api/v1/contacts", nil)
	if err != nil {
		t.Fatalf("contacts request failed: %v", err)
	}

	contacts := data["data"].([]any)
	if len(contacts) != 1 {
		t.Errorf("expected 1 contact, got %d", len(contacts))
	}
}

func TestRulesCommand(t *testing.T) {
	server := mockServer(t)
	defer server.Close()
	setupTestClient(t, server.URL)

	client, _ := api.NewClient()
	data, err := client.Get("/api/v1/rules", nil)
	if err != nil {
		t.Fatalf("rules request failed: %v", err)
	}

	rules := data["data"].([]any)
	if len(rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(rules))
	}

	rule := rules[0].(map[string]any)
	if rule["pattern"] != "STRIPE" {
		t.Errorf("expected pattern 'STRIPE', got %v", rule["pattern"])
	}
}

func TestFormatHelpers(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() string
		expected string
	}{
		{"formatNum nil", func() string { return formatNum(nil) }, "0"},
		{"formatNum float", func() string { return formatNum(float64(42)) }, "42"},
		{"formatNum string", func() string { return formatNum("99") }, "99"},
		{"formatMoney nil", func() string { return formatMoney(nil) }, "$0.00"},
		{"formatMoney float", func() string { return formatMoney(float64(1234.56)) }, "$1234.56"},
		{"formatMoney string", func() string { return formatMoney("500.00") }, "$500.00"},
		{"getString exists", func() string { return getString(map[string]any{"k": "v"}, "k") }, "v"},
		{"getString missing", func() string { return getString(map[string]any{}, "k") }, ""},
		{"getString nil val", func() string { return getString(map[string]any{"k": nil}, "k") }, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn()
			if got != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, got)
			}
		})
	}
}

func TestEmptyResponses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": []any{}})
	}))
	defer server.Close()

	t.Setenv("COOKED_BOOKS_TOKEN", "test")
	t.Setenv("COOKED_BOOKS_API_URL", server.URL)

	client, _ := api.NewClient()

	// All list endpoints should handle empty arrays
	for _, path := range []string{"/api/v1/transactions", "/api/v1/accounts", "/api/v1/invoices", "/api/v1/contacts", "/api/v1/rules"} {
		data, err := client.Get(path, nil)
		if err != nil {
			t.Fatalf("failed on %s: %v", path, err)
		}
		items := data["data"].([]any)
		if len(items) != 0 {
			t.Errorf("expected empty array for %s", path)
		}
	}
}

func TestAuthStatusNoConfig(t *testing.T) {
	t.Setenv("COOKED_BOOKS_TOKEN", "")

	tmpDir := t.TempDir()
	api.ConfigDirOverride = tmpDir
	defer func() { api.ConfigDirOverride = "" }()

	_, err := api.NewClient()
	if err == nil {
		t.Fatal("expected error with no auth")
	}
}

func TestAuthSaveAndRetrieve(t *testing.T) {
	tmpDir := t.TempDir()
	api.ConfigDirOverride = tmpDir
	defer func() { api.ConfigDirOverride = "" }()

	cfg := &api.Config{Token: "round-trip-token", BaseURL: "https://rt.example.com"}
	if err := api.SaveConfig(cfg); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := api.LoadConfig()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded.Token != "round-trip-token" {
		t.Errorf("token mismatch: got '%s'", loaded.Token)
	}

	// Overwrite with empty (logout)
	api.SaveConfig(&api.Config{})
	_, err = api.LoadConfig()
	if err == nil {
		t.Error("expected error after logout (empty token)")
	}
}

func TestAuthConfigFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	api.ConfigDirOverride = tmpDir
	defer func() { api.ConfigDirOverride = "" }()

	api.SaveConfig(&api.Config{Token: "perms-test", BaseURL: "https://example.com"})

	info, _ := os.Stat(tmpDir + "/config.json")
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("expected 0600 permissions, got %o", perm)
	}
}
