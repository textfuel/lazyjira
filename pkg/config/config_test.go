package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestConfigDir_LazyJiraConfigDirPrecedence(t *testing.T) {
	lazyjiraDir := t.TempDir()
	xdgDir := t.TempDir()

	t.Setenv("LAZYJIRA_CONFIG_DIR", lazyjiraDir)
	t.Setenv("XDG_CONFIG_HOME", xdgDir)

	if got := ConfigDir(); got != lazyjiraDir {
		t.Errorf("ConfigDir() = %q, want %q", got, lazyjiraDir)
	}
}

func TestConfigDir_IgnoresLegacyConfigDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("XDG_CONFIG_HOME precedence is not used on Windows")
	}

	legacyDir := t.TempDir()
	xdgDir := t.TempDir()

	t.Setenv("CONFIG_DIR", legacyDir)
	t.Setenv("XDG_CONFIG_HOME", xdgDir)

	want := filepath.Join(xdgDir, "lazyjira")
	if got := ConfigDir(); got != want {
		t.Errorf("ConfigDir() = %q, want %q", got, want)
	}
}

func TestLoad_TLSEnvVars(t *testing.T) {
	t.Setenv("LAZYJIRA_CONFIG_DIR", t.TempDir())
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
	t.Setenv("LAZYJIRA_CONFIG_DIR", dir)

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
	t.Setenv("LAZYJIRA_CONFIG_DIR", dir)

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

func pointerTo[T any](v T) *T { return &v }

func TestResolveMaxResults(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		global *int
		tab    IssueTabConfig
		want   int
	}{
		{"all unset → default", nil, IssueTabConfig{}, DefaultMaxResults},
		{"global only", pointerTo(25), IssueTabConfig{}, 25},
		{"tab overrides global", pointerTo(25), IssueTabConfig{MaxResults: pointerTo(75)}, 75},
		{"negative global ignored", pointerTo(-5), IssueTabConfig{}, DefaultMaxResults},
		{"zero tab falls back to global", pointerTo(40), IssueTabConfig{MaxResults: pointerTo(0)}, 40},
		{"large global not clamped", pointerTo(500), IssueTabConfig{}, 500},
		{"large tab override not clamped", pointerTo(50), IssueTabConfig{MaxResults: pointerTo(1000)}, 1000},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := &Config{MaxResults: tc.global}
			if got := c.ResolveMaxResults(tc.tab); got != tc.want {
				t.Errorf("ResolveMaxResults = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestResolveGlobalMaxResults(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		global *int
		want   int
	}{
		{"nil → default", nil, DefaultMaxResults},
		{"zero → default", pointerTo(0), DefaultMaxResults},
		{"negative → default", pointerTo(-1), DefaultMaxResults},
		{"set", pointerTo(125), 125},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := &Config{MaxResults: tc.global}
			if got := c.ResolveGlobalMaxResults(); got != tc.want {
				t.Errorf("ResolveGlobalMaxResults = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestDefaultConfig_MaxResults(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	if cfg.MaxResults != nil {
		t.Errorf("default MaxResults should be nil (unset), got %d", *cfg.MaxResults)
	}
	if got := cfg.ResolveGlobalMaxResults(); got != DefaultMaxResults {
		t.Errorf("ResolveGlobalMaxResults on defaults = %d, want %d", got, DefaultMaxResults)
	}
}

func TestLoad_ProjectsAcceptsStringShorthand(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LAZYJIRA_CONFIG_DIR", dir)

	cfgYAML := `projects:
  - ORCH
  - key: DATA
    boardId: 7
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(cfgYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := len(cfg.Projects); got != 2 {
		t.Fatalf("len(Projects) = %d, want 2", got)
	}
	if cfg.Projects[0].Key != "ORCH" || cfg.Projects[0].BoardID != 0 {
		t.Errorf("Projects[0] = %+v, want {Key:ORCH BoardID:0}", cfg.Projects[0])
	}
	if cfg.Projects[1].Key != "DATA" || cfg.Projects[1].BoardID != 7 {
		t.Errorf("Projects[1] = %+v, want {Key:DATA BoardID:7}", cfg.Projects[1])
	}
}

func TestValidateConverter(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty defaults to builtin", "", false},
		{"explicit builtin", ConverterBuiltin, false},
		{"adf-converter", ConverterAdfConverter, false},
		{"unknown value errors", "foo", true},
		{"typo errors", "adfconverter", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateConverter(tc.value)
			if tc.wantErr && err == nil {
				t.Errorf("validateConverter(%q) = nil, want error", tc.value)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("validateConverter(%q) = %v, want nil", tc.value, err)
			}
			if tc.wantErr && err != nil && !strings.Contains(err.Error(), tc.value) {
				t.Errorf("error %q should include the bad value %q", err.Error(), tc.value)
			}
		})
	}
}

func TestValidateRenderer(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty defaults to builtin", "", false},
		{"explicit builtin", RendererBuiltin, false},
		{"glamour", RendererGlamour, false},
		{"unknown value errors", "foo", true},
		{"typo errors", "glamor", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateRenderer(tc.value)
			if tc.wantErr && err == nil {
				t.Errorf("validateRenderer(%q) = nil, want error", tc.value)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("validateRenderer(%q) = %v, want nil", tc.value, err)
			}
			if tc.wantErr && err != nil && !strings.Contains(err.Error(), tc.value) {
				t.Errorf("error %q should include the bad value %q", err.Error(), tc.value)
			}
		})
	}
}

func TestLoad_RejectsUnknownConverter(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LAZYJIRA_CONFIG_DIR", dir)
	cfgYAML := "converter: not-a-real-converter\n"
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(cfgYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(); err == nil {
		t.Fatal("Load() with unknown converter should error")
	} else if !strings.Contains(err.Error(), "not-a-real-converter") {
		t.Errorf("error should name the invalid value, got: %v", err)
	}
}

func TestJiraConfig_IsCloud(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		serverType string
		want       bool
	}{
		{"empty defaults to cloud", "", true},
		{"explicit cloud", "cloud", true},
		{"server", "server", false},
		{"datacenter", "datacenter", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			jiraConfig := JiraConfig{ServerType: tc.serverType}
			if got := jiraConfig.IsCloud(); got != tc.want {
				t.Errorf("IsCloud() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGUIConfig_TriStateDefaults(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value *bool
		want  bool
	}{
		{"nil defaults to true", nil, true},
		{"explicit true", pointerTo(true), true},
		{"explicit false", pointerTo(false), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gui := GUIConfig{PrefillFromTab: tc.value, SelectCreatedIssue: tc.value}
			if got := gui.ShouldPrefillFromTab(); got != tc.want {
				t.Errorf("ShouldPrefillFromTab() = %v, want %v", got, tc.want)
			}
			if got := gui.ShouldSelectCreatedIssue(); got != tc.want {
				t.Errorf("ShouldSelectCreatedIssue() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCustomCommandConfig_ShouldSuspend(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		suspend *bool
		want    bool
	}{
		{"nil defaults to true", nil, true},
		{"explicit true", pointerTo(true), true},
		{"explicit false", pointerTo(false), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			command := CustomCommandConfig{Suspend: tc.suspend}
			if got := command.ShouldSuspend(); got != tc.want {
				t.Errorf("ShouldSuspend() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestValidateRendererStyle(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty defaults to auto", "", false},
		{"auto", RendererStyleAuto, false},
		{"dark", RendererStyleDark, false},
		{"light", RendererStyleLight, false},
		{"notty", RendererStyleNoTTY, false},
		{"unknown value errors", "solarized", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateRendererStyle(tc.value)
			if tc.wantErr && err == nil {
				t.Errorf("validateRendererStyle(%q) = nil, want error", tc.value)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("validateRendererStyle(%q) = %v, want nil", tc.value, err)
			}
			if tc.wantErr && err != nil && !strings.Contains(err.Error(), tc.value) {
				t.Errorf("error %q should include the bad value %q", err.Error(), tc.value)
			}
		})
	}
}

func TestConfigDir_FallsBackToUserConfigDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("env fallback chain differs on Windows")
	}

	t.Setenv("LAZYJIRA_CONFIG_DIR", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("os.UserConfigDir: %v", err)
	}
	want := filepath.Join(userConfigDir, "lazyjira")
	if got := ConfigDir(); got != want {
		t.Errorf("ConfigDir() = %q, want %q", got, want)
	}
}

func TestConfigDir_FallsBackToDotConfigWithoutHome(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("env fallback chain differs on Windows")
	}

	t.Setenv("LAZYJIRA_CONFIG_DIR", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "")

	want := filepath.Join(".", ".config", "lazyjira")
	if got := ConfigDir(); got != want {
		t.Errorf("ConfigDir() = %q, want %q", got, want)
	}
}

func TestLoad_ReadErrorWhenConfigFileIsDirectory(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LAZYJIRA_CONFIG_DIR", dir)

	if err := os.Mkdir(filepath.Join(dir, "config.yml"), 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(); err == nil {
		t.Fatal("Load() with unreadable config file should error")
	}
}

func TestLoad_InvalidYAMLErrors(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LAZYJIRA_CONFIG_DIR", dir)

	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte("gui: [unclosed"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(); err == nil {
		t.Fatal("Load() with invalid YAML should error")
	}
}

func TestLoad_MigratesDeprecatedCustomFields(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LAZYJIRA_CONFIG_DIR", dir)

	cfgYAML := `customFields:
  - id: "customfield_10015"
    name: "Story Points"
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(cfgYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Fields) != 1 || cfg.Fields[0].ID != "customfield_10015" {
		t.Errorf("Fields = %+v, want the deprecated customFields entry migrated", cfg.Fields)
	}
	if cfg.DeprecatedFields != nil {
		t.Errorf("DeprecatedFields = %+v, want nil after migration", cfg.DeprecatedFields)
	}
}

func TestLoad_FieldsWinOverDeprecatedCustomFields(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LAZYJIRA_CONFIG_DIR", dir)

	cfgYAML := `fields:
  - id: status
customFields:
  - id: "customfield_10015"
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(cfgYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Fields) != 1 || cfg.Fields[0].ID != "status" {
		t.Errorf("Fields = %+v, want only the explicit fields list", cfg.Fields)
	}
}

func TestLoad_RejectsUnknownRendererAndStyle(t *testing.T) {
	tests := []struct {
		name    string
		cfgYAML string
		wantErr string
	}{
		{"unknown renderer", "renderer: not-a-renderer\n", "not-a-renderer"},
		{"unknown rendererStyle", "rendererStyle: not-a-style\n", "not-a-style"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("LAZYJIRA_CONFIG_DIR", dir)
			if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(tc.cfgYAML), 0o644); err != nil {
				t.Fatal(err)
			}
			_, err := Load()
			if err == nil {
				t.Fatal("Load() should reject the invalid value")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error should name the invalid value, got: %v", err)
			}
		})
	}
}

func TestSave_RoundTripsWithoutToken(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LAZYJIRA_CONFIG_DIR", dir)
	t.Setenv("JIRA_HOST", "")

	cfg := DefaultConfig()
	cfg.Jira.Host = "https://example.atlassian.net"
	cfg.Jira.Token = "super-secret-token"

	if err := Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	written, err := os.ReadFile(ConfigPath())
	if err != nil {
		t.Fatalf("read saved config: %v", err)
	}
	if strings.Contains(string(written), "super-secret-token") {
		t.Error("saved config must not contain the API token")
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load after Save: %v", err)
	}
	if loaded.Jira.Host != "https://example.atlassian.net" {
		t.Errorf("Host = %q, want the saved value", loaded.Jira.Host)
	}
}

func TestSave_CreatesMissingConfigDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "lazyjira")
	t.Setenv("LAZYJIRA_CONFIG_DIR", dir)

	if err := Save(DefaultConfig()); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(ConfigPath()); err != nil {
		t.Errorf("config file missing after Save: %v", err)
	}
}

func TestSave_FailsWhenConfigDirIsFile(t *testing.T) {
	blockingFile := filepath.Join(t.TempDir(), "occupied")
	if err := os.WriteFile(blockingFile, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LAZYJIRA_CONFIG_DIR", blockingFile)

	if err := Save(DefaultConfig()); err == nil {
		t.Fatal("Save() into a file path should error")
	}
}
