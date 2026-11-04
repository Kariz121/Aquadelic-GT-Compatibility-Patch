package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLegacyManifestViaSymlinkedSteamLibrary(t *testing.T) {
	game, manifestPath, _ := setupLegacy(t)
	root := t.TempDir()
	link := filepath.Join(root, "Steam-library-alias")
	if err := os.Symlink(game, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	// The old installer recorded the real directory while RC4 opened the
	// same game via SteamOS's ~/.steam/steam alias (or vice versa).
	dormant, err := inspectLegacy(link)
	if err != nil || len(dormant) != 1 || dormant[0] != "launcher" {
		t.Fatalf("equivalent Steam path must be accepted: %v / %v", dormant, err)
	}
	// The historical manifest and backup must never be modified.
	b, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var m legacyManifest
	if err := json.Unmarshal(b, &m); err != nil || m.Game != game {
		t.Fatalf("historic manifest altered: %s %v", b, err)
	}
}

func TestLegacyAliasDoesNotBypassActivePatch(t *testing.T) {
	game, _, _ := setupLegacy(t)
	link := filepath.Join(t.TempDir(), "other-steam-alias")
	if err := os.Symlink(game, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := os.WriteFile(filepath.Join(game, "Launcher.exe"), []byte("patched launcher bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := inspectLegacy(link)
	if err == nil || !strings.Contains(err.Error(), "still active") {
		t.Fatalf("must still stop active legacy patch: %v", err)
	}
}

func TestLegacyUnrelatedDirectoryStillBlocked(t *testing.T) {
	game, manifestPath, _ := setupLegacy(t)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Launcher.exe"), []byte("launcher original bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "AquadelicGT-Patcher-Backups", "launcher")
	if err := os.MkdirAll(target, 0700); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "active.json"), b, 0600); err != nil {
		t.Fatal(err)
	}
	_, err = inspectLegacy(dir)
	if err == nil || !strings.Contains(err.Error(), "another game directory") {
		t.Fatalf("must reject unrelated directory: %v, original %s", err, game)
	}
}

func TestCurrentPatcherManifestViaSymlinkedSteamLibrary(t *testing.T) {
	game := t.TempDir()
	snap := "test-snapshot"
	dir := filepath.Join(dataDir(game), snap)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "00.original"), []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(game, "Run.exe"), []byte("patched"), 0600); err != nil {
		t.Fatal(err)
	}
	m := Manifest{Version: Version, Edition: "steam-1.4", Game: game, Platform: Platform(), Mode: "step", Snapshot: snap, Files: []Record{{Rel: "Run.exe", OldSHA: SHA([]byte("original")), NewSHA: SHA([]byte("patched")), Backup: "00.original"}}}
	if err := writeManifest(activePath(game), m); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(game, alias); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	mm, err := loadManifest(alias)
	if err != nil || mm == nil {
		t.Fatalf("same-directory RC4 manifest should be recognized after restart: %v", err)
	}
	if _, err := verifyManifest(alias, mm); err != nil {
		t.Fatal(err)
	}
}
