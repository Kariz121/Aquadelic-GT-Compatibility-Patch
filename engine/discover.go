package engine

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

type Candidate struct {
	Path    string `json:"path"`
	Edition string `json:"edition"`
	Status  string `json:"status"`
	Active  bool   `json:"active"`
}

var libraryPathRX = regexp.MustCompile(`(?i)"path"\s*"([^"]+)"`)

// Search only conventional Steam roots and libraryfolders.vdf. Do not walk
// unrelated user disks recursively. Always offer manual selection as fallback.
func Discover() []Candidate {
	roots := []string{}
	// A previously selected nonstandard game directory must be discoverable after restart.
	addRecent := recentGame()
	addRoot := func(s string) {
		if s != "" {
			roots = append(roots, filepath.Clean(s))
		}
	}
	addRoot(addRecent)
	if exe, e := os.Executable(); e == nil {
		addRoot(filepath.Dir(exe))
	}
	if cwd, e := os.Getwd(); e == nil {
		addRoot(cwd)
	}
	home, _ := os.UserHomeDir()
	if runtime.GOOS == "windows" {
		for _, dir := range []string{os.Getenv("PROGRAMFILES(X86)"), os.Getenv("PROGRAMFILES"), os.Getenv("LOCALAPPDATA")} {
			if dir != "" {
				addRoot(filepath.Join(dir, "Steam"))
				addRoot(dir) // boxed retail 1.2 is often installed directly under Program Files
			}
		}
		if home != "" {
			addRoot(filepath.Join(home, "Steam"))
			addRoot(filepath.Join(home, "Games", "Aquadelic GT"))
		}
		for drive := 'C'; drive <= 'H'; drive++ {
			d := string(drive) + ":\\"
			for _, name := range []string{"Steam", "SteamLibrary", "Games", "Games\\SteamLibrary"} {
				addRoot(filepath.Join(d, name))
			}
		}
	} else {
		for _, dir := range []string{".steam/debian-installation", ".steam/steam", ".local/share/Steam", ".steam/root", ".var/app/com.valvesoftware.Steam/.local/share/Steam", "Games/SteamLibrary", "SteamLibrary"} {
			if home != "" {
				addRoot(filepath.Join(home, filepath.FromSlash(dir)))
			}
		}
		addRoot("/mnt/games/SteamLibrary")
		addRoot("/media")
	}
	tried := map[string]bool{}
	result := []Candidate{}
	addCandidate := func(path string) {
		path = filepath.Clean(path)
		if tried[path] {
			return
		}
		tried[path] = true
		if _, e := os.Stat(filepath.Join(path, "Run.exe")); e != nil {
			return
		}
		st := Check(path)
		edition := st.Edition
		if edition == "" {
			edition = "Unknown edition"
		}
		result = append(result, Candidate{Path: path, Edition: edition, Status: st.Info, Active: st.Active})
	}
	for _, root := range roots {
		// The directory of the application may already be the game directory.
		addCandidate(root)
		addCandidate(filepath.Join(root, "Aquadelic GT"))
		addCandidate(filepath.Join(root, "steamapps", "common", "Aquadelic GT"))
		addCandidate(filepath.Join(root, "Steam", "steamapps", "common", "Aquadelic GT"))
		vdf := filepath.Join(root, "steamapps", "libraryfolders.vdf")
		raw, e := os.ReadFile(vdf)
		if e != nil || len(raw) > 2*1024*1024 {
			continue
		}
		for _, m := range libraryPathRX.FindAllStringSubmatch(string(raw), -1) {
			lib := strings.ReplaceAll(m[1], "\\\\", "\\")
			addCandidate(filepath.Join(lib, "steamapps", "common", "Aquadelic GT"))
		}
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].Active && !result[j].Active })
	return result
}
