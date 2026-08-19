package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/fastly/fastly-go/fastly"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// https://www.fastly.com/documentation/reference/api/auth-tokens/user/
var tokensURL = "https://api.fastly.com/tokens"

var rootCmd = &cobra.Command{
	Use:   "fastly-purge",
	Short: "fastly-purge is a cli tool for performing Fastly token operations",
	Long:  "fastly-purge is a cli tool for performing Fastly token operations - creation, deletion, and management.",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Oops. An error while executing fastly-purge '%s'\n", err)
		os.Exit(1)
	}
}

var (
	tokenUsername string
	tokenPassword string
	tokenScope    string
	tokenServices []string
	tokenExpires  string
	tokenOTP      string
)

var createTokenCmd = &cobra.Command{
	Use:   "create-token",
	Short: "Create a Fastly API token scoped for purging",
	Long:  "Create a Fastly API token by authenticating with your Fastly username and password.",
	RunE:  runCreateToken,
}

func init() {
	createTokenCmd.Flags().StringVar(&tokenUsername, "username", os.Getenv("FASTLY_USERNAME"), "Fastly account username/email (or set FASTLY_USERNAME)")
	createTokenCmd.Flags().StringVar(&tokenPassword, "password", os.Getenv("FASTLY_PASSWORD"), "Fastly account password (or set FASTLY_PASSWORD); prompted securely if omitted")
	createTokenCmd.Flags().StringVar(&tokenScope, "scope", "global:read purge_all purge_select", "Token scope: space-delimited list of global:read, purge_all, purge_select, global")
	createTokenCmd.Flags().StringSliceVar(&tokenServices, "service", nil, "Restrict the token to one or more service IDs (repeatable)")
	createTokenCmd.Flags().StringVar(&tokenExpires, "expires-at", "", "Token expiration in ISO 8601 format (optional)")
	createTokenCmd.Flags().StringVar(&tokenOTP, "otp", "", "One-time password, required if two-factor authentication is enabled")

	rootCmd.AddCommand(createTokenCmd)
}

func runCreateToken(cmd *cobra.Command, args []string) error {
	if tokenUsername == "" {
		return fmt.Errorf("username is required (--username or FASTLY_USERNAME)")
	}
	if tokenPassword == "" {
		pw, err := promptPassword("Fastly password: ")
		if err != nil {
			return fmt.Errorf("reading password: %w", err)
		}
		tokenPassword = pw
	}

	form := url.Values{}
	form.Set("username", tokenUsername)
	form.Set("password", tokenPassword)
	form.Set("scope", tokenScope)
	for _, s := range tokenServices {
		form.Add("services[]", s)
	}
	if tokenExpires != "" {
		form.Set("expires_at", tokenExpires)
	}

	req, err := http.NewRequest(http.MethodPost, tokensURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	if tokenOTP != "" {
		req.Header.Set("Fastly-OTP", tokenOTP)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("calling Fastly API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 300 {
		return fmt.Errorf("create token failed: %s: %s", resp.Status, string(body))
	}

	var token fastly.TokenCreatedResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	if token.AccessToken != nil {
		fmt.Fprintf(os.Stdout, "Created token %s (scope: %s)\n", *token.Id, *token.Scope)
		fmt.Fprintf(os.Stdout, "Access token: %s\n", *token.AccessToken)
	}
	return nil
}

func promptPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stdout, prompt)
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stdout)
	if err != nil {
		return "", err
	}
	return string(pw), nil
}
