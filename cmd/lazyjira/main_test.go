package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/config"
	"github.com/textfuel/lazyjira/v2/pkg/tui"
)

func assertEqual[T comparable](t *testing.T, label string, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %v, want %v", label, got, want)
	}
}

func stubServer(t *testing.T, status int) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = io.WriteString(w, "{}")
	}))
	t.Cleanup(server.Close)
	return server
}

func TestIsCloudType(t *testing.T) {
	t.Parallel()
	cases := []struct {
		serverType string
		want       bool
	}{
		{"", true},
		{"cloud", true},
		{"server", false},
		{"datacenter", false},
	}
	for _, tc := range cases {
		assertEqual(t, "isCloudType("+tc.serverType+")", isCloudType(tc.serverType), tc.want)
	}
}

func TestBuildHTTPClient_NoCustomTLSReturnsNil(t *testing.T) {
	t.Parallel()
	client, err := buildHTTPClient(&config.Config{})
	if err != nil {
		t.Fatalf("buildHTTPClient: %v", err)
	}
	if client != nil {
		t.Errorf("client = %v, want nil when no custom TLS", client)
	}
}

func TestBuildHTTPClient_InsecureBuildsClient(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{}
	cfg.Jira.TLS.Insecure = true
	client, err := buildHTTPClient(cfg)
	if err != nil {
		t.Fatalf("buildHTTPClient: %v", err)
	}
	if client == nil {
		t.Error("client = nil, want non-nil for insecure TLS")
	}
}

func TestBuildHTTPClient_MissingCAFileErrors(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{}
	cfg.Jira.TLS.CAFile = filepath.Join(t.TempDir(), "absent.pem")
	if _, err := buildHTTPClient(cfg); err == nil {
		t.Error("expected error for missing CA file")
	}
}

func TestMakeClient_CloudUsesBasicAuthAndV3(t *testing.T) {
	t.Parallel()
	client, err := makeClient(&config.Config{}, "https://acme.atlassian.net", "user@acme.com", "tok", "cloud")
	if err != nil {
		t.Fatalf("makeClient: %v", err)
	}
	assertEqual(t, "base URL", client.BaseURL(), "https://acme.atlassian.net/rest/api/3")
	if !strings.HasPrefix(client.AuthHeader(), "Basic ") {
		t.Errorf("auth header = %q, want Basic prefix", client.AuthHeader())
	}
}

func TestMakeClient_ServerUsesBearerAndV2(t *testing.T) {
	t.Parallel()
	client, err := makeClient(&config.Config{}, "https://jira.acme.com", "", "tok", "server")
	if err != nil {
		t.Fatalf("makeClient: %v", err)
	}
	assertEqual(t, "base URL", client.BaseURL(), "https://jira.acme.com/rest/api/2")
	if !strings.HasPrefix(client.AuthHeader(), "Bearer ") {
		t.Errorf("auth header = %q, want Bearer prefix", client.AuthHeader())
	}
}

func TestMakeClient_BadTLSErrors(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{}
	cfg.Jira.TLS.CAFile = filepath.Join(t.TempDir(), "absent.pem")
	if _, err := makeClient(cfg, "https://acme.atlassian.net", "u@a.com", "tok", "cloud"); err == nil {
		t.Error("expected TLS setup error")
	}
}

func TestTestConnection_OKReturnsNil(t *testing.T) {
	t.Parallel()
	server := stubServer(t, http.StatusOK)
	client, err := makeClient(&config.Config{}, server.URL, "u@a.com", "tok", "cloud")
	if err != nil {
		t.Fatalf("makeClient: %v", err)
	}
	if err := testConnection(client); err != nil {
		t.Errorf("testConnection: %v", err)
	}
}

func TestTestConnection_Non200Errors(t *testing.T) {
	t.Parallel()
	server := stubServer(t, http.StatusUnauthorized)
	client, err := makeClient(&config.Config{}, server.URL, "u@a.com", "tok", "cloud")
	if err != nil {
		t.Fatalf("makeClient: %v", err)
	}
	if err := testConnection(client); err == nil {
		t.Error("expected error for HTTP 401")
	}
}

func TestTestConnection_UnreachableHostErrors(t *testing.T) {
	t.Parallel()
	client, err := makeClient(&config.Config{}, "http://127.0.0.1:1", "u@a.com", "tok", "cloud")
	if err != nil {
		t.Fatalf("makeClient: %v", err)
	}
	if err := testConnection(client); err == nil {
		t.Error("expected error for unreachable host")
	}
}

func TestResolveClient_SavedCloudCredentials(t *testing.T) {
	t.Setenv("LAZYJIRA_CONFIG_DIR", t.TempDir())
	if err := config.SaveCredentials(&config.Credentials{
		Host: "https://acme.atlassian.net", Email: "u@a.com", Token: "tok", ServerType: "cloud",
	}); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}

	client, method, err := resolveClient(&config.Config{})
	if err != nil {
		t.Fatalf("resolveClient: %v", err)
	}
	assertEqual(t, "auth method", method, tui.AuthSaved)
	assertEqual(t, "base URL", client.BaseURL(), "https://acme.atlassian.net/rest/api/3")
}

func TestResolveClient_SavedServerCredentialsWithoutEmail(t *testing.T) {
	t.Setenv("LAZYJIRA_CONFIG_DIR", t.TempDir())
	if err := config.SaveCredentials(&config.Credentials{
		Host: "https://jira.acme.com", Token: "tok", ServerType: "server",
	}); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}

	client, method, err := resolveClient(&config.Config{})
	if err != nil {
		t.Fatalf("resolveClient: %v", err)
	}
	assertEqual(t, "auth method", method, tui.AuthSaved)
	assertEqual(t, "base URL", client.BaseURL(), "https://jira.acme.com/rest/api/2")
}

func TestResolveClient_EnvVarsWhenNoSavedCreds(t *testing.T) {
	t.Setenv("LAZYJIRA_CONFIG_DIR", t.TempDir())
	cfg := &config.Config{}
	cfg.Jira.Host = "https://acme.atlassian.net"
	cfg.Jira.Email = "u@a.com"
	cfg.Jira.Token = "tok"

	client, method, err := resolveClient(cfg)
	if err != nil {
		t.Fatalf("resolveClient: %v", err)
	}
	assertEqual(t, "auth method", method, tui.AuthEnv)
	assertEqual(t, "base URL", client.BaseURL(), "https://acme.atlassian.net/rest/api/3")
}

func TestResolveClient_SavedCloudWithoutEmailFallsBackToEnv(t *testing.T) {
	t.Setenv("LAZYJIRA_CONFIG_DIR", t.TempDir())
	if err := config.SaveCredentials(&config.Credentials{
		Host: "https://saved.atlassian.net", Token: "tok", ServerType: "cloud",
	}); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}
	cfg := &config.Config{}
	cfg.Jira.Host = "https://env.atlassian.net"
	cfg.Jira.Email = "u@a.com"
	cfg.Jira.Token = "tok"

	client, method, err := resolveClient(cfg)
	if err != nil {
		t.Fatalf("resolveClient: %v", err)
	}
	assertEqual(t, "auth method", method, tui.AuthEnv)
	assertEqual(t, "base URL", client.BaseURL(), "https://env.atlassian.net/rest/api/3")
}

func TestRunSetupWizard_MissingHostErrors(t *testing.T) {
	t.Parallel()
	if _, err := runSetupWizard(&config.Config{}, strings.NewReader("1\n\n")); err == nil {
		t.Error("expected host required error")
	}
}

func TestRunSetupWizard_MissingEmailErrors(t *testing.T) {
	t.Parallel()
	if _, err := runSetupWizard(&config.Config{}, strings.NewReader("1\nhttps://acme.atlassian.net\n\n")); err == nil {
		t.Error("expected email required error")
	}
}

func TestRunSetupWizard_MissingTokenErrors(t *testing.T) {
	t.Parallel()
	if _, err := runSetupWizard(&config.Config{}, strings.NewReader("1\nhttps://acme.atlassian.net\nu@a.com\n\n")); err == nil {
		t.Error("expected token required error")
	}
}

func TestRunSetupWizard_CloudSuccessSavesCredentials(t *testing.T) {
	t.Setenv("LAZYJIRA_CONFIG_DIR", t.TempDir())
	server := stubServer(t, http.StatusOK)

	client, err := runSetupWizard(&config.Config{}, strings.NewReader("1\n"+server.URL+"\nu@a.com\ntok\n"))
	if err != nil {
		t.Fatalf("runSetupWizard: %v", err)
	}
	if client == nil {
		t.Fatal("client = nil, want non-nil")
	}

	creds, err := config.LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials: %v", err)
	}
	if creds == nil {
		t.Fatal("creds = nil, want saved credentials")
	}
	assertEqual(t, "saved email", creds.Email, "u@a.com")
	assertEqual(t, "saved server type", creds.ServerType, "cloud")
}

func TestRunSetupWizard_ServerSuccessNeedsNoEmail(t *testing.T) {
	t.Setenv("LAZYJIRA_CONFIG_DIR", t.TempDir())
	server := stubServer(t, http.StatusOK)

	client, err := runSetupWizard(&config.Config{}, strings.NewReader("2\n"+server.URL+"\ntok\n"))
	if err != nil {
		t.Fatalf("runSetupWizard: %v", err)
	}
	if client == nil {
		t.Fatal("client = nil, want non-nil")
	}
	assertEqual(t, "base URL", client.BaseURL(), server.URL+"/rest/api/2")
}

func TestRunSetupWizard_ConnectionFailureErrors(t *testing.T) {
	t.Setenv("LAZYJIRA_CONFIG_DIR", t.TempDir())
	server := stubServer(t, http.StatusUnauthorized)

	if _, err := runSetupWizard(&config.Config{}, strings.NewReader("2\n"+server.URL+"\ntok\n")); err == nil {
		t.Error("expected connection test failed error")
	}
}

func TestDispatch_VersionPrints(t *testing.T) {
	t.Parallel()
	for _, arg := range []string{"--version", "version"} {
		var stdout, stderr strings.Builder
		code := dispatch([]string{arg}, &stdout, &stderr)
		assertEqual(t, "exit code", code, 0)
		if !strings.Contains(stdout.String(), "lazyjira ") {
			t.Errorf("version output = %q, want lazyjira prefix", stdout.String())
		}
	}
}

func TestDispatch_LogoutClearsCredentials(t *testing.T) {
	t.Setenv("LAZYJIRA_CONFIG_DIR", t.TempDir())
	if err := config.SaveCredentials(&config.Credentials{Host: "https://acme.atlassian.net", Token: "tok"}); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}

	var stdout, stderr strings.Builder
	code := dispatch([]string{"logout"}, &stdout, &stderr)

	assertEqual(t, "exit code", code, 0)
	if !strings.Contains(stdout.String(), "Credentials cleared") {
		t.Errorf("logout output = %q", stdout.String())
	}
	if creds, _ := config.LoadCredentials(); creds != nil {
		t.Error("credentials should be removed after logout")
	}
}

func TestRunAuth_WizardSuccess(t *testing.T) {
	t.Setenv("LAZYJIRA_CONFIG_DIR", t.TempDir())
	server := stubServer(t, http.StatusOK)

	if err := runAuth(nil, strings.NewReader("2\n"+server.URL+"\ntok\n")); err != nil {
		t.Errorf("runAuth: %v", err)
	}
}

func TestRunAuth_WizardFailureReturnsError(t *testing.T) {
	t.Setenv("LAZYJIRA_CONFIG_DIR", t.TempDir())

	if err := runAuth(nil, strings.NewReader("1\n\n")); err == nil {
		t.Error("expected error when host is missing")
	}
}
