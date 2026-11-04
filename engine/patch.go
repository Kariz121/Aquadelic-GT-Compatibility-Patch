package engine

import (
	"aquadelic-gt-integrated-patcher/core"
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

const (
	Version          = "1.1.0"
	LinuxBaseSHA     = "7283966d6ce1891cc2634514d525263689135803baba3fda297c17e5d8cc7409"
	LinuxAnalogSHA   = "183288309da895b8dc37eb472719854539dd349aa392e6d6185bf504811712b3"
	LinuxStepSHA     = "22f5a62da4593828b0eacfb92b39d3dc25b03accc1bd0351e1766808c9444af2"
	WindowsBaseSHA   = "b61fb5d2ef4a365a13494286579203ef4a79568e2758def0a3c89b1d1b8c167c"
	WindowsAnalogSHA = "118e80677b1f5851d98055a423b9aeb3a1c2535e5c4cb35eb4f7f03db41076dd"
	WindowsStepSHA   = "c41ba830d92ddb497def36a2f1be8d90ec072685651ee6e7f525daaa6b2fcf9c"
	XMLHydroSHA      = "a38b4ae440e543ae91c0a39721ae57668cfd393c8ac727c65ac70c3b5ad7544e"
)

// Deltas contain only changed bytes / added instructions, never the game's executables.
//
//go:embed patches/linux_analog.aqpt
var linuxAnalog []byte

//go:embed patches/linux_step.aqpt
var linuxStep []byte

//go:embed patches/windows_analog.aqpt
var windowsAnalog []byte

//go:embed patches/windows_step.aqpt
var windowsStep []byte

//go:embed patches/launcher_mesa.aqpt
var launcherMesa []byte

func SHA(data []byte) string { v := sha256.Sum256(data); return hex.EncodeToString(v[:]) }
func Platform() string {
	if runtime.GOOS == "windows" {
		return "windows"
	}
	return "linux"
}
func ModeValid(m string) bool { return m == "none" || m == "analog" || m == "step" }
func ModeLabel(m string) string {
	switch m {
	case "analog":
		return "Boat Analog"
	case "step":
		return "Boat Step"
	case "none":
		return "No controller changes"
	}
	return "Unknown"
}

// ApplyAQPT accepts only an exact supported source hash and validates every
// output byte through the full golden SHA256. It cannot patch unknown versions.
func ApplyAQPT(original, delta []byte, inHash, outHash string) ([]byte, error) {
	if SHA(original) != inHash {
		return nil, fmt.Errorf("Unknown file version (SHA-256 %s, expected %s)", SHA(original), inHash)
	}
	if len(delta) < 16 || string(delta[:4]) != "AQPT" {
		return nil, errors.New("Invalid patch header")
	}
	baseSize, targetSize, count := binary.LittleEndian.Uint32(delta[4:8]), binary.LittleEndian.Uint32(delta[8:12]), binary.LittleEndian.Uint32(delta[12:16])
	if uint64(baseSize) != uint64(len(original)) || count > 10000 || targetSize > uint32(len(original))+0x100000 {
		return nil, errors.New("Invalid patch dimensions")
	}
	result := make([]byte, int(targetSize))
	copy(result, original)
	at := 16
	var prev uint32
	for j := uint32(0); j < count; j++ {
		if at+8 > len(delta) {
			return nil, errors.New("Truncated patch record")
		}
		off, n := binary.LittleEndian.Uint32(delta[at:at+4]), binary.LittleEndian.Uint32(delta[at+4:at+8])
		at += 8
		if n == 0 || off < prev || off > targetSize || n > targetSize-off || uint64(n) > uint64(len(delta)-at) {
			return nil, errors.New("Invalid patch range")
		}
		copy(result[int(off):int(off+n)], delta[at:at+int(n)])
		at += int(n)
		prev = off + n
	}
	if at != len(delta) || SHA(result) != outHash {
		return nil, errors.New("Patched output SHA-256 mismatch")
	}
	return result, nil
}

var settingRX = regexp.MustCompile(`(?s)<IRSetting\b[^>]*>.*?</IRSetting>`)

func HydroXML(ori []byte) ([]byte, error) {
	if SHA(ori) != core.ControllerFinalSHA {
		return nil, fmt.Errorf("Unsupported ControllerSchemes.xml SHA-256 %s", SHA(ori))
	}
	all := settingRX.FindAllIndex(ori, -1)
	if len(all) != 34 {
		return nil, errors.New("ControllerSchemes: expected 34 entries")
	}
	hit := all[27]
	orig := ori[hit[0]:hit[1]]
	replacements := [][2][]byte{{[]byte(`AxisReverse="False"`), []byte(`AxisReverse="True"`)}, {[]byte(`<KeyCode>5</KeyCode>`), []byte(`<KeyCode>0</KeyCode>`)}, {[]byte(`<Caption>RT</Caption>`), []byte(`<Caption>LS-U</Caption>`)}}
	changed := append([]byte(nil), orig...)
	for _, r := range replacements {
		if bytes.Count(changed, r[0]) != 1 {
			return nil, fmt.Errorf("Unexpected hydroplane controller scheme: %s", r[0])
		}
		changed = bytes.Replace(changed, r[0], r[1], 1)
	}
	result := make([]byte, 0, len(ori)+3)
	result = append(result, ori[:hit[0]]...)
	result = append(result, changed...)
	result = append(result, ori[hit[1]:]...)
	if SHA(result) != XMLHydroSHA {
		return nil, fmt.Errorf("Resulting controller XML differs from tested output: %s", SHA(result))
	}
	return result, nil
}

func PatchSteamRun(ori []byte, platform, mode string) ([]byte, error) {
	if SHA(ori) != core.SteamOriginal {
		return nil, fmt.Errorf("Steam 1.4 requires clean Run.exe SHA-256 %s (found %s)", core.SteamOriginal, SHA(ori))
	}
	if platform != "linux" && platform != "windows" {
		return nil, errors.New("Unsupported platform")
	}
	if !ModeValid(mode) {
		return nil, errors.New("Unknown controller mode")
	}
	var err error
	b := ori
	if platform == "linux" {
		b, err = core.ApplyMD5Steam(b)
		if err != nil {
			return nil, err
		}
	}
	b, err = core.PatchBroadcastV2(b)
	if err != nil {
		return nil, err
	}
	if mode == "none" {
		return b, nil
	}
	if platform == "linux" {
		b, err = core.PatchSteamControls(b)
		if err != nil {
			return nil, err
		}
		target := LinuxAnalogSHA
		delta := linuxAnalog
		if mode == "step" {
			target, delta = LinuxStepSHA, linuxStep
		}
		return ApplyAQPT(b, delta, LinuxBaseSHA, target)
	}
	b, err = core.PatchSteamControlsNoMD5(b)
	if err != nil {
		return nil, err
	}
	target := WindowsAnalogSHA
	delta := windowsAnalog
	if mode == "step" {
		target, delta = WindowsStepSHA, windowsStep
	}
	return ApplyAQPT(b, delta, WindowsBaseSHA, target)
}
func PatchLauncher(b []byte, platform string) ([]byte, error) {
	hash := SHA(b)
	if platform == "windows" {
		if hash == core.Launcher140Mesa {
			return append([]byte(nil), b...), nil
		}
		if hash == core.Launcher140Full {
			return nil, errors.New("Windows: the launcher contains a legacy integrity patch. Restore the original Steam 1.4 launcher before installing. The Windows build does not apply launcher MD5 fixes")
		}
		return ApplyAQPT(b, launcherMesa, core.Launcher140Original, core.Launcher140Mesa)
	}
	if hash == core.Launcher140Full {
		return append([]byte(nil), b...), nil
	}
	return core.ApplyLauncher140(b)
}

func safeRead(game, rel string) ([]byte, error) {
	if rel == "" || filepath.IsAbs(rel) || strings.Contains(filepath.ToSlash(rel), "..") {
		return nil, errors.New("Invalid relative path")
	}
	file := filepath.Join(game, filepath.FromSlash(rel))
	st, e := os.Lstat(file)
	if e != nil {
		return nil, e
	}
	if !st.Mode().IsRegular() {
		return nil, fmt.Errorf("Not a regular file: %s", rel)
	}
	return os.ReadFile(file)
}
func findLanguage(game, lang string) (string, error) {
	var found []string
	err := filepath.WalkDir(game, func(path string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			if d.Name() == "AquadelicGT-Patcher-Backups" || d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() || !strings.EqualFold(d.Name(), "Language."+lang) {
			return nil
		}
		rel, e := filepath.Rel(game, path)
		if e != nil {
			return e
		}
		found = append(found, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return "", err
	}
	if len(found) != 1 {
		return "", fmt.Errorf("Language.%s: found %d files (expected 1)", lang, len(found))
	}
	return found[0], nil
}
func findFonts(game string) ([]string, error) {
	var dirs []string
	guiEntries, e := os.ReadDir(game)
	if e != nil {
		return nil, e
	}
	for _, g := range guiEntries {
		if !g.IsDir() || !strings.EqualFold(g.Name(), "Gui") {
			continue
		}
		sub, e := os.ReadDir(filepath.Join(game, g.Name()))
		if e != nil {
			return nil, e
		}
		for _, folder := range sub {
			if !folder.IsDir() || !strings.EqualFold(folder.Name(), "Fonts") {
				continue
			}
			rel := filepath.Join(g.Name(), folder.Name())
			st, e := os.Lstat(filepath.Join(game, rel))
			if e != nil {
				return nil, e
			}
			if !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
				continue
			}
			dirs = append(dirs, rel)
		}
	}
	var candidates [][]string
	for _, dir := range dirs {
		children, e := os.ReadDir(filepath.Join(game, dir))
		if e != nil {
			return nil, e
		}
		var files []string
		ok := true
		for _, name := range core.FontNames {
			var matched string
			for _, d := range children {
				if d.Type().IsRegular() && strings.EqualFold(d.Name(), name) {
					if matched != "" {
						return nil, fmt.Errorf("Duplicate font %s", name)
					}
					matched = d.Name()
				}
			}
			if matched == "" {
				ok = false
				break
			}
			files = append(files, filepath.ToSlash(filepath.Join(dir, matched)))
		}
		if ok {
			candidates = append(candidates, files)
		}
	}
	if len(candidates) != 1 {
		return nil, fmt.Errorf("Ambiguous font directories: %d", len(candidates))
	}
	return candidates[0], nil
}
