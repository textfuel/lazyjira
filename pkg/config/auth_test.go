package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestAuthPath_JoinsConfigDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LAZYJIRA_CONFIG_DIR", dir)

	testkit.AssertEqual(t, "AuthPath", AuthPath(), filepath.Join(dir, "auth.json"))
}

func TestLoadCredentials_MissingReturnsNilWithoutError(t *testing.T) {
	t.Setenv("LAZYJIRA_CONFIG_DIR", t.TempDir())

	creds, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials: %v", err)
	}
	if creds != nil {
		t.Errorf("creds = %#v, want nil for missing file", creds)
	}
}

func TestSaveCredentials_RoundTripsAndIsPrivate(t *testing.T) {
	t.Setenv("LAZYJIRA_CONFIG_DIR", t.TempDir())

	want := &Credentials{
		Host:        "https://test.atlassian.net",
		Email:       "user@test.com",
		Token:       "secret",
		ServerType:  "cloud",
		LastProject: "PLAT",
	}
	if err := SaveCredentials(want); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}

	info, err := os.Stat(AuthPath())
	if err != nil {
		t.Fatalf("stat auth.json: %v", err)
	}
	testkit.AssertEqual(t, "file mode", info.Mode().Perm(), os.FileMode(0o600))

	got, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials: %v", err)
	}
	if got == nil {
		t.Fatal("LoadCredentials returned nil after save")
	}
	testkit.AssertEqual(t, "round-tripped credentials", *got, *want)
}

func TestLoadCredentials_InvalidJSONReturnsError(t *testing.T) {
	t.Setenv("LAZYJIRA_CONFIG_DIR", t.TempDir())
	if err := os.WriteFile(AuthPath(), []byte("{not valid"), 0o600); err != nil {
		t.Fatalf("write auth.json: %v", err)
	}

	if _, err := LoadCredentials(); err == nil {
		t.Error("expected error for malformed auth.json, got nil")
	}
}

func TestClearCredentials_RemovesFile(t *testing.T) {
	t.Setenv("LAZYJIRA_CONFIG_DIR", t.TempDir())
	if err := SaveCredentials(&Credentials{Host: "h"}); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}

	if err := ClearCredentials(); err != nil {
		t.Fatalf("ClearCredentials: %v", err)
	}
	if _, err := os.Stat(AuthPath()); !os.IsNotExist(err) {
		t.Errorf("auth.json still present after ClearCredentials, stat err = %v", err)
	}
}

func TestClearCredentials_MissingIsNotAnError(t *testing.T) {
	t.Setenv("LAZYJIRA_CONFIG_DIR", t.TempDir())

	if err := ClearCredentials(); err != nil {
		t.Errorf("ClearCredentials on missing file = %v, want nil", err)
	}
}

func TestLoadCredentials_ReadErrorWhenAuthPathIsDirectory(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LAZYJIRA_CONFIG_DIR", dir)

	if err := os.Mkdir(AuthPath(), 0o755); err != nil {
		t.Fatal(err)
	}

	credentials, err := LoadCredentials()
	if err == nil {
		t.Fatal("LoadCredentials() with unreadable auth.json should error")
	}
	if credentials != nil {
		t.Errorf("credentials = %+v, want nil on error", credentials)
	}
}

func TestSaveCredentials_FailsWhenConfigDirIsFile(t *testing.T) {
	blockingFile := filepath.Join(t.TempDir(), "occupied")
	if err := os.WriteFile(blockingFile, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LAZYJIRA_CONFIG_DIR", blockingFile)

	if err := SaveCredentials(&Credentials{Host: "https://example.atlassian.net"}); err == nil {
		t.Fatal("SaveCredentials() into a file path should error")
	}
}
