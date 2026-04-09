package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestNewClient_FromEnvVar(t *testing.T) {
	t.Setenv("COOKED_BOOKS_TOKEN", "test-token-123")
	t.Setenv("COOKED_BOOKS_API_URL", "https://test.example.com")

	client, err := NewClient()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if client.Token != "test-token-123" {
		t.Errorf("expected token 'test-token-123', got '%s'", client.Token)
	}
	if client.BaseURL != "https://test.example.com" {
		t.Errorf("expected base URL 'https://test.example.com', got '%s'", client.BaseURL)
	}
}

func TestNewClient_FromConfigFile(t *testing.T) {
	t.Setenv("COOKED_BOOKS_TOKEN", "")

	tmpDir := t.TempDir()
	ConfigDirOverride = tmpDir
	defer func() { ConfigDirOverride = "" }()

	cfg := &Config{Token: "file-token-456", BaseURL: "https://file.example.com"}
	err := SaveConfig(cfg)
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	client, err := NewClient()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if client.Token != "file-token-456" {
		t.Errorf("expected token 'file-token-456', got '%s'", client.Token)
	}
	if client.BaseURL != "https://file.example.com" {
		t.Errorf("expected base URL 'https://file.example.com', got '%s'", client.BaseURL)
	}
}

func TestNewClient_EnvTakesPriority(t *testing.T) {
	t.Setenv("COOKED_BOOKS_TOKEN", "env-token")
	t.Setenv("COOKED_BOOKS_API_URL", "https://env.example.com")

	tmpDir := t.TempDir()
	ConfigDirOverride = tmpDir
	defer func() { ConfigDirOverride = "" }()

	SaveConfig(&Config{Token: "file-token", BaseURL: "https://file.example.com"})

	client, err := NewClient()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if client.Token != "env-token" {
		t.Errorf("expected env token, got '%s'", client.Token)
	}
	if client.BaseURL != "https://env.example.com" {
		t.Errorf("expected env base URL, got '%s'", client.BaseURL)
	}
}

func TestNewClient_NoAuth(t *testing.T) {
	t.Setenv("COOKED_BOOKS_TOKEN", "")

	tmpDir := t.TempDir()
	ConfigDirOverride = tmpDir
	defer func() { ConfigDirOverride = "" }()

	_, err := NewClient()
	if err == nil {
		t.Fatal("expected error when no auth configured")
	}
}

func TestClient_Get_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected Bearer auth, got '%s'", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("expected Accept: application/json")
		}
		if r.URL.Query().Get("status") != "new" {
			t.Errorf("expected status=new param, got '%s'", r.URL.Query().Get("status"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": []any{"item1"}})
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, Token: "test-token", HTTP: &http.Client{}}

	result, err := client.Get("/api/v1/transactions", map[string]string{"status": "new"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	data, ok := result["data"].([]any)
	if !ok || len(data) != 1 {
		t.Errorf("expected data with 1 item, got: %v", result)
	}
}

func TestClient_Get_SkipsEmptyParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "status=new" {
			t.Errorf("expected only status param, got query: '%s'", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{})
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, Token: "t", HTTP: &http.Client{}}
	_, err := client.Get("/test", map[string]string{"status": "new", "search": ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_Get_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"error": "invalid token"}`))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, Token: "bad", HTTP: &http.Client{}}
	_, err := client.Get("/test", nil)
	if err == nil {
		t.Fatal("expected error on 401")
	}
	if err.Error() != "unauthorized. Run: cooked-books auth set-token" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestClient_Get_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		w.Write([]byte(`requires Pro plan`))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, Token: "t", HTTP: &http.Client{}}
	_, err := client.Get("/test", nil)
	if err == nil {
		t.Fatal("expected error on 403")
	}
	if err.Error() != "forbidden: requires Pro plan" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestClient_Get_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`internal error`))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, Token: "t", HTTP: &http.Client{}}
	_, err := client.Get("/test", nil)
	if err == nil {
		t.Fatal("expected error on 500")
	}
}

func TestClient_Get_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`not json`))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, Token: "t", HTTP: &http.Client{}}
	_, err := client.Get("/test", nil)
	if err == nil {
		t.Fatal("expected error on invalid JSON")
	}
}

func TestClient_Post_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type: application/json")
		}

		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["key"] != "value" {
			t.Errorf("expected payload key=value, got: %v", body)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, Token: "t", HTTP: &http.Client{}}
	result, err := client.Post("/test", map[string]any{"key": "value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("expected status=ok, got: %v", result)
	}
}

func TestClient_Post_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, Token: "bad", HTTP: &http.Client{}}
	_, err := client.Post("/test", nil)
	if err == nil {
		t.Fatal("expected error on 401")
	}
}

func TestSaveConfig_And_LoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	ConfigDirOverride = tmpDir
	defer func() { ConfigDirOverride = "" }()

	cfg := &Config{Token: "save-test-token", BaseURL: "https://save.example.com"}
	err := SaveConfig(cfg)
	if err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Check file permissions
	info, err := os.Stat(filepath.Join(tmpDir, "config.json"))
	if err != nil {
		t.Fatalf("config file not created: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected 0600 permissions, got %o", info.Mode().Perm())
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}
	if loaded.Token != "save-test-token" {
		t.Errorf("expected token 'save-test-token', got '%s'", loaded.Token)
	}
	if loaded.BaseURL != "https://save.example.com" {
		t.Errorf("expected base URL 'https://save.example.com', got '%s'", loaded.BaseURL)
	}
}

func TestLoadConfig_EmptyToken(t *testing.T) {
	tmpDir := t.TempDir()
	ConfigDirOverride = tmpDir
	defer func() { ConfigDirOverride = "" }()

	SaveConfig(&Config{Token: "", BaseURL: "https://example.com"})

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error when token is empty")
	}
}

func TestLoadConfig_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	ConfigDirOverride = tmpDir
	defer func() { ConfigDirOverride = "" }()

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error when no config file")
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	ConfigDirOverride = tmpDir
	defer func() { ConfigDirOverride = "" }()

	os.WriteFile(filepath.Join(tmpDir, "config.json"), []byte("not json"), 0600)

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error on invalid JSON")
	}
}

func TestDefaultBaseURL(t *testing.T) {
	t.Setenv("COOKED_BOOKS_API_URL", "")
	if DefaultBaseURL() != "https://api.cookedbooks.ai" {
		t.Errorf("expected default URL, got '%s'", DefaultBaseURL())
	}

	t.Setenv("COOKED_BOOKS_API_URL", "http://localhost:4847")
	if DefaultBaseURL() != "http://localhost:4847" {
		t.Errorf("expected override URL, got '%s'", DefaultBaseURL())
	}
}

func TestConfigDir_Override(t *testing.T) {
	ConfigDirOverride = "/tmp/test-dir"
	defer func() { ConfigDirOverride = "" }()

	if ConfigDir() != "/tmp/test-dir" {
		t.Errorf("expected override dir, got '%s'", ConfigDir())
	}
}

func TestConfigDir_Default(t *testing.T) {
	ConfigDirOverride = ""
	dir := ConfigDir()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "cooked-books")
	if dir != expected {
		t.Errorf("expected '%s', got '%s'", expected, dir)
	}
}
