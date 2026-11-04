package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupLegacy(t *testing.T) (game, manifestPath, content string) {
	t.Helper()
	game = t.TempDir()
	live := []byte("launcher original bytes")
	if err := os.WriteFile(filepath.Join(game, "Launcher.exe"), live, 0600); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(game, "AquadelicGT-Patcher-Backups", "launcher")
	if err := os.MkdirAll(filepath.Join(dir, "snapshot-test"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "snapshot-test", "00.original"), live, 0600); err != nil {
		t.Fatal(err)
	}
	m := legacyManifest{Format: 1, Game: game, Module: "launcher", Snapshot: "snapshot-test", Files: []legacyRecord{{Rel: "Launcher.exe", Before: SHA(live), After: SHA([]byte("patched launcher bytes")), Backup: "00.original"}}}
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	manifestPath = filepath.Join(dir, "active.json")
	if err := os.WriteFile(manifestPath, b, 0600); err != nil {
		t.Fatal(err)
	}
	return game, manifestPath, string(b)
}

func TestLegacyManifestAlreadyRestored(t *testing.T) {
	game, path, before := setupLegacy(t)
	dormant, err := inspectLegacy(game)
	if err != nil || len(dormant) != 1 || dormant[0] != "launcher" {
		t.Fatalf("expected stale launcher: %v / %v", dormant, err)
	}
	raw, err := os.ReadFile(path)
	if err != nil || string(raw) != before {
		t.Fatal("inspection must not alter historical manifest")
	}
}
func TestLegacyManifestStillActive(t *testing.T) {
	game, _, _ := setupLegacy(t)
	if err := os.WriteFile(filepath.Join(game, "Launcher.exe"), []byte("patched launcher bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := inspectLegacy(game)
	if err == nil || !strings.Contains(err.Error(), "still active") {
		t.Fatalf("expected active STOP: %v", err)
	}
}
func TestLegacyManifestUnexpectedState(t *testing.T) {
	game, _, _ := setupLegacy(t)
	if err := os.WriteFile(filepath.Join(game, "Launcher.exe"), []byte("third-party modifications"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := inspectLegacy(game)
	if err == nil || !strings.Contains(err.Error(), "unknown content") {
		t.Fatalf("expected unknown STOP: %v", err)
	}
}
func TestLegacyBackupIntegrity(t *testing.T) {
	game, _, _ := setupLegacy(t)
	backup := filepath.Join(game, "AquadelicGT-Patcher-Backups", "launcher", "snapshot-test", "00.original")
	if err := os.WriteFile(backup, []byte("tampered"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := inspectLegacy(game)
	if err == nil || !strings.Contains(err.Error(), "corrupted copy") {
		t.Fatalf("expected corrupt STOP: %v", err)
	}
}
func TestRestoreWhenNothingInstalled(t *testing.T) {
	game := t.TempDir()
	if err := os.WriteFile(filepath.Join(game, "Run.exe"), []byte("placeholder"), 0600); err != nil {
		t.Fatal(err)
	}
	msg, err := Restore(game)
	if err != nil || !strings.Contains(msg, "No active") {
		t.Fatalf("no-op restore: %q / %v", msg, err)
	}
}
