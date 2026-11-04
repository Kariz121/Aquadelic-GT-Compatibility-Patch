package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyManifestAfterRestartAllowsOriginalOrPatched(t *testing.T) {
	game := t.TempDir()
	oldData := []byte("original-bytes")
	newData := []byte("patched-bytes")
	rel := "Run.exe"
	snap := "test-snapshot"
	folder := filepath.Join(dataDir(game), snap)
	if err := os.MkdirAll(folder, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, "00.original"), oldData, 0600); err != nil {
		t.Fatal(err)
	}
	m := Manifest{Version: Version, Edition: "steam-1.4", Game: game, Platform: Platform(), Mode: "step", Snapshot: snap,
		Files: []Record{{Rel: rel, OldSHA: SHA(oldData), NewSHA: SHA(newData), Backup: "00.original"}}}
	if err := writeManifest(activePath(game), m); err != nil {
		t.Fatal(err)
	}
	// Simulate a new process by rereading the durable manifest, rather than
	// retaining an in-memory pointer created during installation.
	readBack, err := loadManifest(game)
	if err != nil || readBack == nil {
		t.Fatalf("manifest reread: %v", err)
	}
	for _, valid := range [][]byte{newData, oldData} {
		if err := os.WriteFile(filepath.Join(game, rel), valid, 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := verifyManifest(game, readBack); err != nil {
			t.Fatalf("expected reversible state: %v", err)
		}
	}
	if err := os.WriteFile(filepath.Join(game, rel), []byte("unknown user changes"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := verifyManifest(game, readBack); err == nil {
		t.Fatal("must reject unknown modification")
	}
}
