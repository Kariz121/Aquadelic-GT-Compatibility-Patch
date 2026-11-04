package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// RecentGame is a convenience hint only. Backups and installation state are
// stored inside the game directory; this hint is never used to bypass checks.
type recentPath struct {
	Path string `json:"path"`
}

func recentFile() string {
	config, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(config, "AquadelicGT-Compatibility-Patcher", "last-game.json")
}
func recentGame() string {
	file := recentFile()
	if file == "" {
		return ""
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return ""
	}
	var v recentPath
	if json.Unmarshal(data, &v) != nil {
		return ""
	}
	if v.Path == "" {
		return ""
	}
	return v.Path
}

func RememberGame(path string) {
	g, err := NormalizeGame(path)
	if err != nil {
		return
	}
	file := recentFile()
	if file == "" {
		return
	}
	if os.MkdirAll(filepath.Dir(file), 0700) != nil {
		return
	}
	b, err := json.Marshal(recentPath{Path: g})
	if err != nil {
		return
	}
	_ = os.WriteFile(file, append(b, '\n'), 0600)
}
