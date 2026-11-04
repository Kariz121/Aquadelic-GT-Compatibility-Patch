package engine

import (
	"aquadelic-gt-integrated-patcher/core"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileChange struct {
	Rel    string `json:"rel"`
	Before []byte `json:"-"`
	After  []byte `json:"-"`
}
type Record struct {
	Rel    string `json:"rel"`
	OldSHA string `json:"old_sha256"`
	NewSHA string `json:"new_sha256"`
	Backup string `json:"backup"`
}
type Manifest struct {
	Version  string   `json:"version"`
	Edition  string   `json:"edition"`
	Game     string   `json:"game"`
	Platform string   `json:"platform"`
	Mode     string   `json:"mode"`
	Snapshot string   `json:"snapshot"`
	Files    []Record `json:"files"`
}
type Status struct {
	Game      string `json:"game"`
	Exists    bool   `json:"exists"`
	Edition   string `json:"edition"`
	Platform  string `json:"platform"`
	Mode      string `json:"mode"`
	RunSHA    string `json:"run_sha256"`
	Active    bool   `json:"active"`
	Ready     bool   `json:"ready"`
	Info      string `json:"info"`
	FileCount int    `json:"file_count"`
}

func NormalizeGame(s string) (string, error) {
	s = strings.Trim(strings.TrimSpace(s), "\"' ")
	if strings.EqualFold(filepath.Base(s), "Run.exe") {
		s = filepath.Dir(s)
	}
	if s == "" {
		return "", errors.New("Select a game folder")
	}
	abs, e := filepath.Abs(s)
	if e != nil {
		return "", e
	}
	st, e := os.Stat(abs)
	if e != nil {
		return "", e
	}
	if !st.IsDir() {
		return "", errors.New("The selected path is not a game folder")
	}
	// Avoid patching a symbolic-link replacement of the game's executable.
	_, e = safeRead(abs, "Run.exe")
	if e != nil {
		return "", fmt.Errorf("Run.exe: %w", e)
	}
	return abs, nil
}
func dataDir(game string) string    { return filepath.Join(game, "AquadelicGT-Patcher-Backups", "v1.1.0") }
func activePath(game string) string { return filepath.Join(dataDir(game), "active.json") }
func loadManifest(game string) (*Manifest, error) {
	b, e := os.ReadFile(activePath(game))
	if os.IsNotExist(e) {
		return nil, nil
	}
	if e != nil {
		return nil, e
	}
	var m Manifest
	if e = json.Unmarshal(b, &m); e != nil {
		return nil, e
	}
	if m.Version != Version || !sameGameFolder(m.Game, game) || m.Platform != Platform() || m.Snapshot == "" || len(m.Files) == 0 || !ModeValid(m.Mode) {
		return nil, errors.New("Unrecognized or corrupt installation manifest. No files were changed")
	}
	return &m, nil
}
func verifyManifest(game string, m *Manifest) (map[string][]byte, error) {
	backups := map[string][]byte{}
	for _, r := range m.Files {
		if r.Rel == "" || strings.Contains(r.Rel, "..") || filepath.IsAbs(r.Rel) || r.Backup == "" || strings.Contains(r.Backup, "..") || filepath.IsAbs(r.Backup) {
			return nil, errors.New("Invalid file path in manifest")
		}
		current, e := safeRead(game, r.Rel)
		if e != nil {
			return nil, e
		}
		if SHA(current) != r.NewSHA && SHA(current) != r.OldSHA {
			return nil, fmt.Errorf("%s changed since installation (SHA=%s). Operation blocked to protect your files; do not delete your backup", r.Rel, SHA(current))
		}
		old, e := os.ReadFile(filepath.Join(dataDir(game), m.Snapshot, r.Backup))
		if e != nil {
			return nil, e
		}
		if SHA(old) != r.OldSHA {
			return nil, fmt.Errorf("Corrupted backup for %s", r.Rel)
		}
		backups[r.Rel] = old
	}
	return backups, nil
}
func Check(game string) Status {
	p := Platform()
	st := Status{Game: game, Platform: p, Mode: "none"}
	g, e := NormalizeGame(game)
	if e != nil {
		st.Info = e.Error()
		return st
	}
	st.Game = g
	st.Exists = true
	run, e := safeRead(g, "Run.exe")
	if e != nil {
		st.Info = e.Error()
		return st
	}
	st.RunSHA = SHA(run)
	m, e := loadManifest(g)
	if e != nil {
		st.Info = "MANIFEST ERROR: " + e.Error()
		return st
	}
	if m != nil {
		st.Active = true
		st.Edition = m.Edition
		st.Mode = m.Mode
		st.FileCount = len(m.Files)
		if _, e = verifyManifest(g, m); e != nil {
			st.Info = "REVIEW REQUIRED: " + e.Error()
			return st
		}
		st.Ready = true
		st.Info = "Patcher 1.1.0 is installed. You can switch modes or restore the original files."
		return st
	}
	switch st.RunSHA {
	case core.RetailOriginal:
		st.Edition = "retail-1.2"
		st.Ready = true
		st.Info = "Polish retail 1.2 detected — legacy V44 patch supported"
	case core.RetailPatched:
		st.Edition = "retail-1.2"
		st.Ready = true
		st.Info = "Retail 1.2 already contains the V44 patch. Restore it with the original installer backup."
	case core.SteamOriginal:
		st.Edition = "steam-1.4"
		st.Ready = true
		st.Info = "Original Steam 1.4 detected — ready to install"
	default:
		st.Info = "Unrecognized or previously modified Run.exe. Restore clean game files before installing."
	}
	if st.Ready {
		markers, err := inspectLegacy(g)
		if err != nil {
			st.Ready = false
			st.Info = "STOP: " + err.Error()
			return st
		}
		// A clean Run.exe alone is insufficient: old installers could leave
		// the launcher, fonts or other files independently patched.
		if st.Edition == "steam-1.4" {
			if _, _, err := buildPlan(g, p, "none", nil); err != nil {
				st.Ready = false
				st.Info = "Files need attention: " + err.Error()
				return st
			}
		}
		if len(markers) != 0 {
			st.Info += "\nLegacy manifests (" + strings.Join(markers, ", ") + ") refer to restored files and will not block installation."
		}
	}
	return st
}
func buildPlan(game, platform, mode string, orig map[string][]byte) (string, []FileChange, error) {
	if !ModeValid(mode) {
		return "", nil, errors.New("Unknown controller mode")
	}
	var changes []FileChange
	get := func(rel string) ([]byte, error) {
		if old, ok := orig[rel]; ok {
			return old, nil
		}
		return safeRead(game, rel)
	}
	add := func(rel string, fn func([]byte) ([]byte, error)) error {
		b, e := get(rel)
		if e != nil {
			return fmt.Errorf("%s: %w", rel, e)
		}
		c, e := fn(b)
		if e != nil {
			return fmt.Errorf("%s: %w", rel, e)
		}
		if len(c) == 0 {
			return fmt.Errorf("%s: patch returned an empty file", rel)
		}
		changes = append(changes, FileChange{Rel: rel, Before: b, After: c})
		return nil
	}
	run, e := get("Run.exe")
	if e != nil {
		return "", nil, e
	}
	edition := ""
	switch SHA(run) {
	case core.RetailOriginal:
		edition = "retail-1.2"
		if mode != "none" {
			return "", nil, errors.New("Retail 1.2 supports the legacy V44 patch only. Controller modes require Steam 1.4.")
		}
		if e = add("Run.exe", core.ApplyRetail); e != nil {
			return "", nil, e
		}
		return edition, changes, nil
	case core.SteamOriginal:
		edition = "steam-1.4"
	default:
		return "", nil, fmt.Errorf("Unsupported original Run.exe (SHA-256 %s)", SHA(run))
	}
	if e = add("Run.exe", func(b []byte) ([]byte, error) { return PatchSteamRun(b, platform, mode) }); e != nil {
		return "", nil, e
	}
	if e = add("Launcher.exe", func(b []byte) ([]byte, error) { return PatchLauncher(b, platform) }); e != nil {
		return "", nil, e
	}
	for _, rel := range core.BloomFiles {
		if e = add(rel, core.PatchBloomXML); e != nil {
			return "", nil, e
		}
	}
	for _, lang := range []string{"pl", "esp", "fr", "hu"} {
		rel, e := findLanguage(game, lang)
		if e != nil {
			return "", nil, e
		}
		label := lang
		if e = add(rel, func(b []byte) ([]byte, error) {
			ori, patched, ok := core.LanguageTargetHash(label)
			if !ok {
				return nil, fmt.Errorf("Missing localization patch table for %s", label)
			}
			if SHA(b) == patched {
				return append([]byte(nil), b...), nil
			}
			if SHA(b) != ori {
				return nil, fmt.Errorf("Language.%s has unrecognized SHA-256 %s", label, SHA(b))
			}
			return core.PatchLanguage(label, b)
		}); e != nil {
			return "", nil, e
		}
	}
	// Include XML in manifest even when unchanged, to allow later None->Analog/Step.
	if e = add("Configs/ControllerSchemes.xml", func(b []byte) ([]byte, error) {
		if mode == "none" {
			if SHA(b) != core.ControllerOriginalSHA {
				return nil, errors.New("Unrecognized original controller scheme")
			}
			return append([]byte(nil), b...), nil
		}
		x, err := core.PatchControllerXML(b)
		if err != nil {
			return nil, err
		}
		return HydroXML(x)
	}); e != nil {
		return "", nil, e
	}
	fonts, e := findFonts(game)
	if e != nil {
		return "", nil, e
	}
	for _, rel := range fonts {
		if platform == "linux" {
			if e = add(rel, func(b []byte) ([]byte, error) { return core.SetFontRevision(b, core.FontCompatRevision) }); e != nil {
				return "", nil, e
			}
		} else {
			b, err := get(rel)
			if err != nil {
				return "", nil, err
			}
			rev, err := core.FontRevision(b)
			if err != nil {
				return "", nil, err
			}
			if rev != core.FontOriginalRevision {
				return "", nil, fmt.Errorf("Windows: font %s was modified (revision %d). Restore the original TTF. The Windows patcher does not modify fonts", rel, rev)
			}
		}
	}
	return edition, changes, nil
}
func newSnapshot() (string, error) {
	b := make([]byte, 6)
	if _, e := rand.Read(b); e != nil {
		return "", e
	}
	return time.Now().UTC().Format("20060102T150405Z") + "-" + hex.EncodeToString(b), nil
}
func atomicReplace(path string, data []byte) error {
	st, e := os.Lstat(path)
	if e != nil {
		return e
	}
	if !st.Mode().IsRegular() {
		return errors.New("Not a regular file or symlink: " + path)
	}
	f, e := os.CreateTemp(filepath.Dir(path), ".aqgt-v110-*")
	if e != nil {
		return e
	}
	temp := f.Name()
	defer os.Remove(temp)
	if _, e = f.Write(data); e != nil {
		f.Close()
		return e
	}
	if e = f.Sync(); e != nil {
		f.Close()
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	if e = os.Chmod(temp, st.Mode().Perm()); e != nil {
		return e
	}
	if e = swapFile(temp, path); e != nil {
		return e
	}
	after, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	if SHA(after) != SHA(data) {
		return errors.New("Written file failed SHA-256 verification")
	}
	return nil
}
func writeNew(path string, data []byte) error {
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	if _, e = f.Write(data); e != nil {
		f.Close()
		return e
	}
	if e = f.Sync(); e != nil {
		f.Close()
		return e
	}
	return f.Close()
}
func writeManifest(path string, m Manifest) error {
	b, e := json.MarshalIndent(m, "", "  ")
	if e != nil {
		return e
	}
	return writeNew(path, append(b, '\n'))
}
func recoverChanges(game string, applied []FileChange) {
	for j := len(applied) - 1; j >= 0; j-- {
		_ = atomicReplace(filepath.Join(game, filepath.FromSlash(applied[j].Rel)), applied[j].Before)
	}
}
func Install(path, mode string) (string, error) {
	game, e := NormalizeGame(path)
	if e != nil {
		return "", e
	}
	status := Check(game)
	if !status.Ready {
		return "", errors.New(status.Info)
	}
	if status.Active {
		return Switch(game, mode)
	}
	if status.RunSHA == core.RetailPatched {
		if mode != "none" {
			return "", errors.New("Retail 1.2 supports only its original controller configuration")
		}
		return "Retail 1.2: V44 is already installed. No changes needed.", nil
	}
	edition, changes, e := buildPlan(game, Platform(), mode, nil)
	if e != nil {
		return "", e
	}
	snap, e := newSnapshot()
	if e != nil {
		return "", e
	}
	folder := filepath.Join(dataDir(game), snap)
	if e = os.MkdirAll(folder, 0700); e != nil {
		return "", e
	}
	manifest := Manifest{Version: Version, Edition: edition, Game: game, Platform: Platform(), Mode: mode, Snapshot: snap}
	for i, c := range changes {
		backup := fmt.Sprintf("%02d.original", i)
		if e = writeNew(filepath.Join(folder, backup), c.Before); e != nil {
			return "", fmt.Errorf("backup %s: %w", c.Rel, e)
		}
		manifest.Files = append(manifest.Files, Record{Rel: c.Rel, OldSHA: SHA(c.Before), NewSHA: SHA(c.After), Backup: backup})
	}
	var applied []FileChange
	for _, c := range changes {
		if SHA(c.Before) == SHA(c.After) {
			continue
		}
		if e = atomicReplace(filepath.Join(game, filepath.FromSlash(c.Rel)), c.After); e != nil {
			recoverChanges(game, append(applied, c))
			return "", fmt.Errorf("Writing %s: %w; attempting rollback", c.Rel, e)
		}
		applied = append(applied, c)
	}
	if e = writeManifest(activePath(game), manifest); e != nil {
		recoverChanges(game, applied)
		return "", fmt.Errorf("Manifest: %w; attempting rollback", e)
	}
	return fmt.Sprintf("Installed %s / %s. Backup: %s. Changed %d files.", edition, ModeLabel(mode), folder, len(applied)), nil
}
func Switch(path, mode string) (string, error) {
	game, e := NormalizeGame(path)
	if e != nil {
		return "", e
	}
	m, e := loadManifest(game)
	if e != nil {
		return "", e
	}
	if m == nil {
		return "", errors.New("No v1.1.0 installation found to switch modes")
	}
	orig, e := verifyManifest(game, m)
	if e != nil {
		return "", e
	}
	if m.Mode == mode {
		return "Selected controller mode is already active. No changes needed.", nil
	}
	if m.Edition == "retail-1.2" {
		return "", errors.New("Retail 1.2 has no alternative throttle modes")
	}
	_, fresh, e := buildPlan(game, Platform(), mode, orig)
	if e != nil {
		return "", e
	}
	planned := map[string][]byte{}
	for _, c := range fresh {
		planned[c.Rel] = c.After
	}
	// Never write another module during a mode switch; unchanged resources are
	// compared to the expected manifest and remain untouched.
	var changes []FileChange
	for _, r := range m.Files {
		if r.Rel != "Run.exe" && r.Rel != "Configs/ControllerSchemes.xml" {
			continue
		}
		current, err := safeRead(game, r.Rel)
		if err != nil {
			return "", err
		}
		result, ok := planned[r.Rel]
		if !ok {
			return "", fmt.Errorf("Missing controller-mode output for %s", r.Rel)
		}
		changes = append(changes, FileChange{Rel: r.Rel, Before: current, After: result})
	}
	var applied []FileChange
	for _, c := range changes {
		if SHA(c.Before) == SHA(c.After) {
			continue
		}
		if err := atomicReplace(filepath.Join(game, filepath.FromSlash(c.Rel)), c.After); err != nil {
			recoverChanges(game, append(applied, c))
			return "", err
		}
		applied = append(applied, c)
	}
	next := *m
	next.Mode = mode
	next.Files = append([]Record(nil), m.Files...)
	for i := range next.Files {
		if value, ok := planned[next.Files[i].Rel]; ok {
			next.Files[i].NewSHA = SHA(value)
		}
	}
	bytes, e := json.MarshalIndent(next, "", "  ")
	if e != nil {
		recoverChanges(game, applied)
		return "", e
	}
	if e = atomicReplace(activePath(game), append(bytes, '\n')); e != nil {
		recoverChanges(game, applied)
		return "", e
	}
	return "Controller mode changed to " + ModeLabel(mode) + " without overwriting original backups.", nil
}
func Restore(path string) (string, error) {
	game, e := NormalizeGame(path)
	if e != nil {
		return "", e
	}
	m, e := loadManifest(game)
	if e != nil {
		return "", e
	}
	if m == nil {
		return "No active v1.1.0 installation to restore. For older patches, use their original installers.", nil
	}
	orig, e := verifyManifest(game, m)
	if e != nil {
		return "", e
	}
	var applied []FileChange
	for _, r := range m.Files {
		current, e := safeRead(game, r.Rel)
		if e != nil {
			recoverChanges(game, applied)
			return "", e
		}
		old := orig[r.Rel]
		if SHA(current) == SHA(old) {
			continue
		}
		if e = atomicReplace(filepath.Join(game, filepath.FromSlash(r.Rel)), old); e != nil {
			// The attempted file may already have been replaced before the
			// final verification failed. Roll it back as well.
			recoverChanges(game, append(applied, FileChange{Rel: r.Rel, Before: current, After: old}))
			return "", e
		}
		applied = append(applied, FileChange{Rel: r.Rel, Before: current, After: old})
	}
	// Archive the manifest without destroying historical backups.
	archived := filepath.Join(dataDir(game), m.Snapshot, "restored-manifest.json")
	if e = swapFile(activePath(game), archived); e != nil {
		recoverChanges(game, applied)
		return "", e
	}
	return "Original files restored. Backup history was preserved at " + filepath.Join(dataDir(game), m.Snapshot), nil
}
