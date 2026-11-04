package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Historical 1.5.x installers used per-module active.json records. A leftover
// manifest is not by itself proof that a game file is still patched. Never
// delete/modify historical backups while inspecting them.
type legacyRecord struct {
	Rel    string `json:"rel"`
	Before string `json:"before"`
	After  string `json:"after"`
	Backup string `json:"backup"`
}
type legacyManifest struct {
	Format   int            `json:"format"`
	Game     string         `json:"game"`
	Module   string         `json:"module"`
	Snapshot string         `json:"snapshot"`
	Files    []legacyRecord `json:"files"`
}

var legacyModules = []string{"run", "launcher", "bloom", "fonts", "broadcast", "controls", "languages"}

func isSingleComponent(s string) bool {
	return s != "" && s != "." && s != ".." && filepath.Base(s) == s &&
		!strings.ContainsAny(s, `/\\`) && !filepath.IsAbs(s)
}

func readRegular(path string) ([]byte, error) {
	st, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() {
		return nil, errors.New("Not a regular file")
	}
	return os.ReadFile(path)
}

// sameGameFolder accepts distinct path spellings only when they resolve to the
// same existing directory. This matters on SteamOS, where ~/.steam/steam and
// ~/.local/share/Steam commonly lead to the same Steam library. An unrelated
// directory, even with a matching name, must not be accepted.
func sameGameFolder(recorded, current string) bool {
	if recorded == "" || current == "" {
		return false
	}
	if filepath.Clean(recorded) == filepath.Clean(current) {
		return true
	}
	// Windows folder paths are case insensitive. Stat/SameFile additionally
	// covers junctions, while Linux SameFile covers symlinks and bind mounts.
	if Platform() == "windows" && strings.EqualFold(filepath.Clean(recorded), filepath.Clean(current)) {
		return true
	}
	former, err1 := os.Stat(recorded)
	live, err2 := os.Stat(current)
	return err1 == nil && err2 == nil && former.IsDir() && live.IsDir() && os.SameFile(former, live)
}

// inspectLegacy validates the complete historical manifest, original backup
// bytes, and current file hashes. It permits a purely archival manifest ONLY
// when every covered game file exactly matches its recorded pre-install SHA.
// Any still-patched or unknown state remains a hard stop.
func inspectLegacy(game string) ([]string, error) {
	var oldMarkers []string
	for _, module := range legacyModules {
		directory := filepath.Join(game, "AquadelicGT-Patcher-Backups", module)
		manifestFile := filepath.Join(directory, "active.json")
		data, err := readRegular(manifestFile)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("Legacy backup %s: cannot read manifest: %w", module, err)
		}
		var man legacyManifest
		if err := json.Unmarshal(data, &man); err != nil {
			return nil, fmt.Errorf("Legacy backup %s: unreadable manifest: %w", module, err)
		}
		if man.Format != 1 || man.Module != module || !isSingleComponent(man.Snapshot) || len(man.Files) == 0 {
			return nil, fmt.Errorf("Legacy backup %s has invalid manifest metadata (format, module, snapshot or files). Game files remain untouched", module)
		}
		if !sameGameFolder(man.Game, game) {
			return nil, fmt.Errorf("Legacy backup %s references another game directory (%q instead of %q). Game files remain untouched", module, man.Game, game)
		}
		for _, entry := range man.Files {
			if entry.Rel == "" || filepath.IsAbs(entry.Rel) || strings.Contains(filepath.ToSlash(entry.Rel), "..") ||
				!isSingleComponent(entry.Backup) || len(entry.Before) != 64 || len(entry.After) != 64 {
				return nil, fmt.Errorf("Legacy backup %s: invalid record %q", module, entry.Rel)
			}
			backup, err := readRegular(filepath.Join(directory, man.Snapshot, entry.Backup))
			if err != nil || !strings.EqualFold(SHA(backup), entry.Before) {
				return nil, fmt.Errorf("Legacy backup %s: missing or corrupted copy of %s. No changes made", module, entry.Rel)
			}
			live, err := safeRead(game, entry.Rel)
			if err != nil {
				return nil, fmt.Errorf("Legacy backup %s: %s: %w", module, entry.Rel, err)
			}
			actual := SHA(live)
			if !strings.EqualFold(actual, entry.Before) {
				if strings.EqualFold(actual, entry.After) {
					return nil, fmt.Errorf("Legacy patch is still active (%s: %s). Restore that module using its original installer. The v1.1.0 Restore action only restores v1.1.0 changes", module, entry.Rel)
				}
				return nil, fmt.Errorf("Legacy backup %s: %s has unknown content (SHA-256 %s). Inspect before installing; no changes made", module, entry.Rel, actual)
			}
		}
		oldMarkers = append(oldMarkers, module)
	}
	return oldMarkers, nil
}
