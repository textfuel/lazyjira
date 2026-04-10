package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_TLSEnvVars(t *testing.T) {
	t.Setenv("CONFIG_DIR", t.TempDir())
	t.Setenv("JIRA_SERVER_TYPE", "server")
	t.Setenv("JIRA_TLS_CERT", "/tmp/cert.pem")
	t.Setenv("JIRA_TLS_KEY", "/tmp/key.pem")
	t.Setenv("JIRA_TLS_CA", "/tmp/ca.pem")
	t.Setenv("JIRA_TLS_INSECURE", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Jira.ServerType != "server" {
		t.Errorf("ServerType = %q, want server", cfg.Jira.ServerType)
	}
	if cfg.Jira.TLS.CertFile != "/tmp/cert.pem" {
		t.Errorf("CertFile = %q", cfg.Jira.TLS.CertFile)
	}
	if cfg.Jira.TLS.KeyFile != "/tmp/key.pem" {
		t.Errorf("KeyFile = %q", cfg.Jira.TLS.KeyFile)
	}
	if cfg.Jira.TLS.CAFile != "/tmp/ca.pem" {
		t.Errorf("CAFile = %q", cfg.Jira.TLS.CAFile)
	}
	if !cfg.Jira.TLS.Insecure {
		t.Error("Insecure should be true")
	}
}

func TestLoad_CustomCommandRefreshFromYAML(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CONFIG_DIR", dir)

	cfgYAML := `customCommands:
  - key: "y"
    name: "copy"
    command: "echo {{.Key}}"
    suspend: false
  - key: "w"
    name: "log work"
    command: "echo {{.Key}}"
    refresh: true
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(cfgYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.CustomCommands[0].Refresh {
		t.Error("first command: Refresh should default to false")
	}
	if !cfg.CustomCommands[1].Refresh {
		t.Error("second command: Refresh should be true")
	}
}

func TestLoad_InvalidCustomCommandTemplate(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CONFIG_DIR", dir)

	cfgYAML := `customCommands:
  - key: "y"
    name: "broken"
    command: "echo {{.Unclosed"
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(cfgYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid template, got nil")
	}
	if !strings.Contains(err.Error(), "template parse error") {
		t.Errorf("error = %q, want it to mention template parse error", err)
	}
}
