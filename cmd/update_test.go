package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadSubscriptionURL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", ".proxyctl-subscription")
	const want = "https://example.com/subscription?id=123"
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("old\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := saveSubscriptionURL(path, "  "+want+" \n"); err != nil {
		t.Fatal(err)
	}
	got, err := loadSubscriptionURL(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("loadSubscriptionURL() = %q, want %q", got, want)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("subscription file mode = %o, want 600", info.Mode().Perm())
	}
}

func TestLoadSubscriptionURLRejectsEmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".proxyctl-subscription")
	if err := os.WriteFile(path, []byte(" \n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadSubscriptionURL(path); err == nil {
		t.Fatal("loadSubscriptionURL() accepted an empty file")
	}
}
