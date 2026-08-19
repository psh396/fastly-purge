package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fastly/fastly-go/fastly"
)

func TestRunCreateToken_MissingUsername(t *testing.T) {
	tokenUsername = ""
	tokenPassword = "password"

	cmd := createTokenCmd
	err := runCreateToken(cmd, nil)

	if err == nil {
		t.Errorf("expected error for missing username, got nil")
	}
	if !strings.Contains(err.Error(), "username is required") {
		t.Errorf("expected 'username is required' error, got: %v", err)
	}
}

func TestRunCreateToken_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("expected form content-type, got %s", r.Header.Get("Content-Type"))
		}

		err := r.ParseForm()
		if err != nil {
			t.Fatalf("parsing form: %v", err)
		}
		if r.FormValue("username") != "test@example.com" {
			t.Errorf("expected username test@example.com, got %s", r.FormValue("username"))
		}
		if r.FormValue("password") != "testpass" {
			t.Errorf("expected password testpass, got %s", r.FormValue("password"))
		}
		if r.FormValue("scope") != "global:read purge_all purge_select" {
			t.Errorf("expected scope, got %s", r.FormValue("scope"))
		}

		w.Header().Set("Content-Type", "application/json")
		resp := fastly.TokenCreatedResponse{
			Id:          stringPtr("token123"),
			AccessToken: stringPtr("access_token_secret"),
			Scope:       stringPtr("global:read purge_all purge_select"),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Override the API endpoint for this test
	oldURL := tokensURL
	tokensURL = server.URL
	defer func() { tokensURL = oldURL }()

	tokenUsername = "test@example.com"
	tokenPassword = "testpass"
	tokenScope = "global:read purge_all purge_select"
	tokenServices = nil
	tokenExpires = ""
	tokenOTP = ""

	cmd := createTokenCmd
	err := runCreateToken(cmd, nil)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestRunCreateToken_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid_grant", "error_description": "Invalid username or password"}`))
	}))
	defer server.Close()

	oldURL := tokensURL
	tokensURL = server.URL
	defer func() { tokensURL = oldURL }()

	tokenUsername = "test@example.com"
	tokenPassword = "wrongpass"
	tokenScope = "global:read purge_all purge_select"
	tokenServices = nil
	tokenExpires = ""
	tokenOTP = ""

	cmd := createTokenCmd
	err := runCreateToken(cmd, nil)

	if err == nil {
		t.Errorf("expected error for bad request, got nil")
	}
	if !strings.Contains(err.Error(), "create token failed") {
		t.Errorf("expected 'create token failed' error, got: %v", err)
	}
}

func TestRunCreateToken_WithServices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			t.Fatalf("parsing form: %v", err)
		}

		services := r.Form["services[]"]
		if len(services) != 2 || services[0] != "service1" || services[1] != "service2" {
			t.Errorf("expected services [service1 service2], got %v", services)
		}

		w.Header().Set("Content-Type", "application/json")
		resp := fastly.TokenCreatedResponse{
			Id:          stringPtr("token456"),
			AccessToken: stringPtr("access_token_secret"),
			Scope:       stringPtr("global:read purge_all purge_select"),
			Services:    []string{"service1", "service2"},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	oldURL := tokensURL
	tokensURL = server.URL
	defer func() { tokensURL = oldURL }()

	tokenUsername = "test@example.com"
	tokenPassword = "testpass"
	tokenScope = "global:read purge_all purge_select"
	tokenServices = []string{"service1", "service2"}
	tokenExpires = ""
	tokenOTP = ""

	cmd := createTokenCmd
	err := runCreateToken(cmd, nil)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestRunCreateToken_With2FA(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Fastly-OTP") != "123456" {
			t.Errorf("expected Fastly-OTP header, got %s", r.Header.Get("Fastly-OTP"))
		}

		w.Header().Set("Content-Type", "application/json")
		resp := fastly.TokenCreatedResponse{
			Id:          stringPtr("token789"),
			AccessToken: stringPtr("access_token_secret"),
			Scope:       stringPtr("global:read purge_all purge_select"),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	oldURL := tokensURL
	tokensURL = server.URL
	defer func() { tokensURL = oldURL }()

	tokenUsername = "test@example.com"
	tokenPassword = "testpass"
	tokenScope = "global:read purge_all purge_select"
	tokenServices = nil
	tokenExpires = ""
	tokenOTP = "123456"

	cmd := createTokenCmd
	err := runCreateToken(cmd, nil)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestRunCreateToken_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{invalid json}`))
	}))
	defer server.Close()

	oldURL := tokensURL
	tokensURL = server.URL
	defer func() { tokensURL = oldURL }()

	tokenUsername = "test@example.com"
	tokenPassword = "testpass"
	tokenScope = "global:read purge_all purge_select"
	tokenServices = nil
	tokenExpires = ""
	tokenOTP = ""

	cmd := createTokenCmd
	err := runCreateToken(cmd, nil)

	if err == nil {
		t.Errorf("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "decoding response") {
		t.Errorf("expected 'decoding response' error, got: %v", err)
	}
}

func TestPromptPassword(t *testing.T) {
	// Skip test in non-TTY environment (like CI or pipes)
	if !isTerminal() {
		t.Skip("skipping password prompt test in non-TTY environment")
	}
}

func isTerminal() bool {
	return true // In real TTY; placeholder for actual check
}

func stringPtr(s string) *string {
	return &s
}
